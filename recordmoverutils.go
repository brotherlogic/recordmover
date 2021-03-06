package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordmover/proto"
	pbro "github.com/brotherlogic/recordsorganiser/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func (s *Server) addToArchive(ctx context.Context, move *pb.RecordedMove) error {
	moves, err := s.readMoveArchive(ctx, move.GetInstanceId())
	if status.Convert(err).Code() != codes.OK && status.Convert(err).Code() != codes.NotFound {
		return err
	}

	for _, movea := range moves {
		if movea.GetMoveTime() == move.GetMoveTime() {
			return fmt.Errorf("This move has already been recorded")
		}
	}

	moves = append(moves, move)

	return s.saveMoveArchive(ctx, move.GetInstanceId(), moves)
}

func (s *Server) incrementCount(ctx context.Context, id int32) error {
	if s.lastID == id {
		s.lastIDCount++
	} else {
		s.lastIDCount = 1
		s.lastID = id
	}

	if s.lastIDCount > 100 {
		s.RaiseIssue("Stuck move", fmt.Sprintf("%v cannot be moved", id))
		return fmt.Errorf("Stuck Move")
	}

	return nil
}

type getter interface {
	getRecordsSince(ctx context.Context, since int64) ([]int32, error)
	getRecord(ctx context.Context, instanceID int32) (*pbrc.Record, error)
	update(ctx context.Context, instanceID int32, reason string, moveFolder int32) error
}

func (s *Server) refreshMove(ctx context.Context, move *pb.RecordMove) error {
	s.Log(fmt.Sprintf("Refreshing: %v", move.InstanceId))

	//Hydrate the origin
	if move.GetBeforeContext().GetLocation() == "" {
		loc, err := s.organiser.locate(ctx, &pbro.LocateRequest{FolderId: move.GetFromFolder()})
		if err != nil {
			return err
		}
		if move.GetBeforeContext() == nil {
			move.BeforeContext = &pb.Context{}
		}
		move.GetBeforeContext().Location = loc.GetFoundLocation().GetName()
	}

	location, err := s.organiser.locate(ctx, &pbro.LocateRequest{InstanceId: move.InstanceId})

	if err != nil {
		return err
	}

	for i, r := range location.GetFoundLocation().GetReleasesLocation() {
		if r.GetInstanceId() == move.InstanceId {
			if move.AfterContext == nil {
				move.AfterContext = &pb.Context{}
			}

			if i > 0 {
				move.GetAfterContext().Location = location.GetFoundLocation().Name
				move.GetAfterContext().Slot = location.GetFoundLocation().GetReleasesLocation()[i].Slot
				move.GetAfterContext().BeforeInstance = location.GetFoundLocation().GetReleasesLocation()[i-1].InstanceId
			} else {
				move.AfterContext.BeforeInstance = -2
			}

			if i < len(location.GetFoundLocation().GetReleasesLocation())-1 {
				move.GetAfterContext().Location = location.GetFoundLocation().Name
				move.GetAfterContext().Slot = location.GetFoundLocation().GetReleasesLocation()[i].Slot
				move.GetAfterContext().AfterInstance = location.GetFoundLocation().GetReleasesLocation()[i+1].InstanceId

			} else {
				move.AfterContext.AfterInstance = -2
			}
		}
	}

	move.LastUpdate = time.Now().Unix()
	return nil
}

var (
	backlog = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "recordmover_backlog",
		Help: "The size of the print queue",
	})
)

func (s *Server) moveRecordInternal(ctx context.Context, record *pbrc.Record) error {
	folder, rule := s.moveRecord(ctx, record)
	if record.GetRelease().GetFolderId() == 812802 && record.GetMetadata().GetRecordWidth() == 0 &&
		(record.GetMetadata().GetGoalFolder() != 2274270 && record.GetMetadata().GetGoalFolder() != 1782105) {
		s.RaiseIssue(fmt.Sprintf("%v needs record width", record.GetRelease().GetInstanceId()), fmt.Sprintf("Record is %v", record.GetRelease().GetTitle()))
		return status.Errorf(codes.InvalidArgument, "%v needs to have the record width set", record.GetRelease().GetInstanceId())
	}

	if folder > 0 || len(rule) > 0 {
		s.Log(fmt.Sprintf("MOVED: %v -> %v, %v", record.GetRelease().GetInstanceId(), folder, rule))
	}
	if folder > 0 {
		s.addToArchive(ctx, &pb.RecordedMove{
			InstanceId: record.GetRelease().GetInstanceId(),
			From:       record.GetRelease().GetFolderId(),
			To:         folder,
			MoveTime:   time.Now().Unix(),
			Rule:       rule,
		})
		err := s.getter.update(ctx, record.GetRelease().GetInstanceId(), rule, folder)
		if err != nil {
			return err
		}
		s.incrementCount(ctx, record.GetRelease().InstanceId)
		return nil
	}

	if len(rule) > 0 {
		if strings.Contains(rule, "Missing match") {
			return status.Errorf(codes.FailedPrecondition, "Temp unable to move: %v", rule)
		}
		return fmt.Errorf("Unable to move record: %v", rule)
	}

	return nil
}

func (s *Server) canMove(ctx context.Context, r *pbrc.Record) error {
	// Can't move a record with no goal
	if r.GetMetadata() != nil && r.GetMetadata().GoalFolder == 0 {
		return fmt.Errorf("No Goal")
	}

	//We can always move to digital
	if r.GetMetadata() != nil && r.GetMetadata().GoalFolder == 268147 {
		return nil
	}

	if r.GetMetadata().GetMatch() == pbrc.ReleaseMetadata_MATCH_UNKNOWN {
		s.forceMatch(ctx, r.GetRelease().GetInstanceId())
		return fmt.Errorf("Missing match: %v", r.GetRelease().GetInstanceId())
	}

	// Only check for non GOOGLE_PLAY releases
	if r.GetMetadata().GetCategory() != pbrc.ReleaseMetadata_GOOGLE_PLAY && r.GetMetadata().GetGoalFolder() != 1782105 && r.GetMetadata().GetGoalFolder() != 1433217 {
		for _, f := range r.GetRelease().GetFormats() {
			if f.Name == "CD" || f.Name == "CDr" {
				if len(r.GetMetadata().CdPath) == 0 {
					return fmt.Errorf("No CDPath: %v", r.GetMetadata())
				}
			}
		}
	}

	return nil
}

func (s *Server) moveRecord(ctx context.Context, r *pbrc.Record) (int32, string) {
	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_GOOGLE_PLAY && (r.GetRelease().FolderId != 1433217 && r.GetMetadata().MoveFolder != 1433217) {
		return 1433217, "GPLAY"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_DIGITAL && (r.GetRelease().FolderId != 268147 && r.GetMetadata().MoveFolder != 268147) && r.GetRelease().FolderId != 1433217 {
		return 268147, "DIGITAL"
	}

	// We can always move something for processing.
	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_RIP_THEN_SELL && (r.GetRelease().FolderId != 812802 && r.GetMetadata().MoveFolder != 812802) {
		return 812802, "RIP THEN SELL"
	}

	err := s.canMove(ctx, r)
	if err != nil {
		return -1, fmt.Sprintf("CANNOT MOVE %v: %v", r.GetRelease().GetInstanceId(), err)
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PARENTS && (r.GetRelease().FolderId != 1727264 && r.GetMetadata().MoveFolder != 1727264) {
		return 1727264, "PARENTS"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_ASSESS_FOR_SALE && (r.GetRelease().FolderId != 1362206 && r.GetMetadata().MoveFolder != 1362206) {
		return 1362206, "ASSESS FOR SALE"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_ASSESS && (r.GetRelease().FolderId != 1362206 && r.GetMetadata().MoveFolder != 1362206) {
		return 1362206, "ASSESSING"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_VALIDATE && (r.GetRelease().FolderId != 812802 && r.GetMetadata().MoveFolder != 812802) {
		return 812802, "VALIDATING"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_NO_LABELS && (r.GetRelease().FolderId != 1362206 && r.GetMetadata().MoveFolder != 1362206) {
		return 1362206, "NO LABELS"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOLD && r.GetRelease().FolderId != 488127 {
		return 488127, "SOLD"
	}

	if (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOLD_ARCHIVE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOLD_OFFLINE) && r.GetRelease().FolderId != 1613206 && r.GetMetadata().MoveFolder != 1613206 {
		return 1613206, "SOLD_ARCHI"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNLISTENED && r.GetRelease().FolderId != 812802 {
		return 812802, "UNLISTE"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_STAGED && r.GetRelease().FolderId != 673768 {
		return 673768, "STAGED"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_STALE_SALE && r.GetRelease().FolderId != 1708299 {
		return 1708299, "STALE SALE"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_HIGH_SCHOOL && r.GetRelease().FolderId != 673768 {
		return 673768, "HIGH SCHOOL"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL && r.GetRelease().FolderId != 673768 {
		return 673768, "PRE HIGH SCHOOL"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_LISTED_TO_SELL && r.GetRelease().FolderId != 488127 {
		return 488127, "LSITEND TO SELL"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_FRESHMAN && r.GetRelease().FolderId != 812802 {
		return 812802, "PRE FERSHMAN"
	}
	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_FRESHMAN {
		if r.GetMetadata().GetGoalFolder() != 0 && (r.GetRelease().FolderId != r.GetMetadata().GetGoalFolder() && r.GetMetadata().MoveFolder != r.GetMetadata().GetGoalFolder()) {
			return r.GetMetadata().GetGoalFolder(), "FRESHMAN MOVE"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_STAGED_TO_SELL && r.GetRelease().FolderId != 812802 && r.GetMetadata().MoveFolder != 812802 {
		r.GetMetadata().MoveFolder = 812802
		return 812802, "STAGED TO SELL"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PROFESSOR ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_PROFESSOR ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_POSTDOC ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_POSTDOC ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_GRADUATE ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_GRADUATE ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOPHMORE ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_SOPHMORE ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_DISTINGUISHED ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_DISTINGUISHED {
		if r.GetMetadata().GetGoalFolder() != 0 && (r.GetRelease().FolderId != r.GetMetadata().GetGoalFolder() && r.GetMetadata().MoveFolder != r.GetMetadata().GetGoalFolder()) {
			return r.GetMetadata().GetGoalFolder(), "GOAL FOLDER"
		}
	}

	return -1, ""
}
