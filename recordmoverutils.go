package main

import (
	"fmt"
	"time"

	pb "github.com/brotherlogic/recordmover/proto"
	pbro "github.com/brotherlogic/recordsorganiser/proto"
	"golang.org/x/net/context"

	pbrc "github.com/brotherlogic/recordcollection/proto"
)

func (s *Server) incrementCount(ctx context.Context, id int32) error {
	if s.lastID == id {
		s.lastIDCount++
	} else {
		s.lastIDCount = 1
		s.lastID = id
	}

	if s.lastIDCount > 100 {
		s.RaiseIssue(ctx, "Stuck move", fmt.Sprintf("%v cannot be moved", id), false)
		return fmt.Errorf("Stuck Move")
	}

	return nil
}

type getter interface {
	getRecordsSince(ctx context.Context, since int64) ([]int32, error)
	getRecord(ctx context.Context, instanceID int32) (*pbrc.Record, error)
	update(ctx context.Context, rec *pbrc.Record) error
}

func (s *Server) refreshMoves(ctx context.Context) error {
	for _, r := range s.config.Moves {
		if time.Now().Sub(time.Unix(r.LastUpdate, 0)) > time.Hour {
			err := s.refreshMove(ctx, r)
			if err == nil {
				s.saveMoves(ctx)
				return nil
			}
		}
	}

	return nil
}

func (s *Server) refreshMove(ctx context.Context, move *pb.RecordMove) error {
	s.Log(fmt.Sprintf("Refreshing: %b", move.InstanceId))
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

	if move.GetAfterContext() != nil && move.GetAfterContext().Location != "" {
		s.updateArchive(&pb.RecordedMove{InstanceId: move.InstanceId, MoveLocation: move.GetAfterContext().Location, MoveTime: time.Now().Unix()})
	}

	move.LastUpdate = time.Now().Unix()
	return nil
}

func (s *Server) moveRecords(ctx context.Context) error {
	return s.moveRecordsHelper(ctx, 0)
}

func (s *Server) moveRecordInternal(ctx context.Context, record *pbrc.Record) error {
	update, rule := s.moveRecord(ctx, record)
	s.Log(fmt.Sprintf("MOVED: %v, %v", update, rule))
	if update != nil {
		err := s.getter.update(ctx, update)
		if err != nil {
			return err
		}
		s.incrementCount(ctx, record.GetRelease().InstanceId)
	} else {
		if len(rule) > 0 {
			return fmt.Errorf("Unable to move record: %v", rule)
		}
	}

	return nil
}

func (s *Server) moveRecordsHelper(ctx context.Context, instanceID int32) error {
	records, err := s.getter.getRecordsSince(ctx, s.config.LastPull)
	s.Log(fmt.Sprintf("Found %v records since %v", len(records), time.Unix(s.config.LastPull, 0)))
	s.total = len(records)

	if err != nil {
		return err
	}

	s.count = 0
	badRecords := []int32{}
	for _, id := range records {
		record, err := s.getter.getRecord(ctx, id)
		if err != nil {
			return err
		}
		s.count++
		if record.GetMetadata() != nil && !record.GetMetadata().Dirty {
			if instanceID == 0 || record.GetRelease().InstanceId == instanceID {
				err := s.moveRecordInternal(ctx, record)
				if err != nil {
					s.Log(fmt.Sprintf("Unable to move %v -> %v", record.GetRelease().InstanceId, err))
					badRecords = append(badRecords, record.GetRelease().InstanceId)
				}
			}
		}
	}

	if len(badRecords) > 0 {
		s.RaiseIssue(ctx, "Stuck Records", fmt.Sprintf("%v are all stuck", badRecords), false)
	}

	s.config.LastPull = time.Now().Unix()
	s.saveMoves(ctx)
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

	// Only check for non GOOGLE_PLAY releases
	if r.GetMetadata().GetCategory() != pbrc.ReleaseMetadata_GOOGLE_PLAY {
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

func (s *Server) moveRecord(ctx context.Context, r *pbrc.Record) (*pbrc.Record, string) {
	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_GOOGLE_PLAY && (r.GetRelease().FolderId != 1433217 && r.GetMetadata().MoveFolder != 1433217) {
		r.GetMetadata().MoveFolder = 1433217
		return r, "GPLAY"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_DIGITAL && (r.GetRelease().FolderId != 268147 && r.GetMetadata().MoveFolder != 268147) {
		r.GetMetadata().MoveFolder = 268147
		return r, "DIGITAL"
	}

	// We can always move something for processing.
	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_RIP_THEN_SELL && (r.GetRelease().FolderId != 812802 && r.GetMetadata().MoveFolder != 812802) {
		r.GetMetadata().MoveFolder = 812802
		return r, "RIP THEN SELL"
	}

	err := s.canMove(ctx, r)
	if err != nil {
		return nil, fmt.Sprintf("CANNOT MOVE: %v", err)
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PARENTS && (r.GetRelease().FolderId != 1727264 && r.GetMetadata().MoveFolder != 1727264) {
		r.GetMetadata().MoveFolder = 1727264
		return r, "PARENTS"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_ASSESS_FOR_SALE && (r.GetRelease().FolderId != 1362206 && r.GetMetadata().MoveFolder != 1362206) {
		r.GetMetadata().MoveFolder = 1362206
		r.GetMetadata().Purgatory = pbrc.Purgatory_NEEDS_STOCK_CHECK
		return r, "ASSESS FOR SALE"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_ASSESS && (r.GetRelease().FolderId != 1362206 && r.GetMetadata().MoveFolder != 1362206) {
		r.GetMetadata().MoveFolder = 1362206
		r.GetMetadata().Purgatory = pbrc.Purgatory_NEEDS_STOCK_CHECK
		return r, "ASSESSING"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_NO_LABELS && (r.GetRelease().FolderId != 1362206 && r.GetMetadata().MoveFolder != 1362206) {
		r.GetMetadata().MoveFolder = 1362206
		r.GetMetadata().Purgatory = pbrc.Purgatory_NEEDS_LABELS
		return r, "NO LABELS"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOLD && r.GetRelease().FolderId != 488127 {
		r.GetMetadata().MoveFolder = 488127
		return r, "SOLD"
	}

	if (r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOLD_ARCHIVE || r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOLD_OFFLINE) && r.GetRelease().FolderId != 1613206 && r.GetMetadata().MoveFolder != 1613206 {
		r.GetMetadata().MoveFolder = 1613206
		return r, "SOLD_ARCHI"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNLISTENED && r.GetRelease().FolderId != 812802 {
		r.GetMetadata().MoveFolder = 812802
		return r, "UNLISTE"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_STAGED && r.GetRelease().FolderId != 673768 {
		r.GetMetadata().MoveFolder = 673768
		return r, "STAGED"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_STALE_SALE && r.GetRelease().FolderId != 1708299 {
		r.GetMetadata().MoveFolder = 1708299
		return r, "STALE SALE"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_HIGH_SCHOOL && r.GetRelease().FolderId != 673768 {
		r.GetMetadata().MoveFolder = 673768
		return r, "HIGH SCHOOL"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL && r.GetRelease().FolderId != 673768 {
		r.GetMetadata().MoveFolder = 673768
		return r, "PRE HIGH SCHOOL"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_LISTED_TO_SELL && r.GetRelease().FolderId != 488127 {
		r.GetMetadata().MoveFolder = 488127
		return r, "LSITEND TO SELL"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_FRESHMAN && r.GetRelease().FolderId != 812802 {
		r.GetMetadata().MoveFolder = 812802
		return r, "PRE FERSHMAN"
	}
	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_FRESHMAN {
		if r.GetMetadata().GetGoalFolder() != 0 && (r.GetRelease().FolderId != r.GetMetadata().GetGoalFolder() && r.GetMetadata().MoveFolder != r.GetMetadata().GetGoalFolder()) {
			r.GetMetadata().MoveFolder = r.GetMetadata().GetGoalFolder()
			return r, "FRESHMAN MOVE"
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_STAGED_TO_SELL && r.GetRelease().FolderId != 812802 && r.GetMetadata().MoveFolder != 812802 {
		r.GetMetadata().MoveFolder = 812802
		return r, "STAGED TO SELL"
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
			r.GetMetadata().MoveFolder = r.GetMetadata().GetGoalFolder()
			return r, "GOAL FOLDER"
		}
	}

	return nil, ""
}
