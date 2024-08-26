package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pbgd "github.com/brotherlogic/godiscogs/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordmover/proto"
	pbro "github.com/brotherlogic/recordsorganiser/proto"
)

func (s *Server) addToArchive(ctx context.Context, move *pb.RecordedMove) error {
	moves, err := s.readMoveArchive(ctx, move.GetInstanceId())
	if status.Convert(err).Code() != codes.OK && status.Convert(err).Code() != codes.InvalidArgument {
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
	s.CtxLog(ctx, fmt.Sprintf("Refreshing: %v", move.InstanceId))

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
		for _, r := range loc.GetFoundLocation().GetReleasesLocation() {
			if r.GetInstanceId() == move.InstanceId {
				move.GetBeforeContext().Slot = r.GetSlot()
			}
		}
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

			move.GetAfterContext().Location = location.GetFoundLocation().Name
			move.GetAfterContext().Slot = location.GetFoundLocation().GetReleasesLocation()[i].Slot

			if i > 0 {
				move.GetAfterContext().BeforeInstance = location.GetFoundLocation().GetReleasesLocation()[i-1].InstanceId
			} else {
				move.AfterContext.BeforeInstance = -2
			}

			if i < len(location.GetFoundLocation().GetReleasesLocation())-1 {
				move.GetAfterContext().AfterInstance = location.GetFoundLocation().GetReleasesLocation()[i+1].InstanceId
			} else {
				move.AfterContext.AfterInstance = -2
			}
		}
	}

	move.LastUpdate = time.Now().Unix()
	return nil
}

func isTwelve(record *pbrc.Record) bool {
	isTwelve := false
	for _, format := range record.GetRelease().GetFormats() {
		if format.GetName() == "LP" {
			isTwelve = true
		}
		for _, desc := range format.GetDescriptions() {
			if desc == "LP" || desc == "12\"" || desc == "10\"" {
				isTwelve = true
			}
		}
	}
	return isTwelve
}

func isCD(record *pbrc.Record) bool {
	isCD := false
	for _, format := range record.GetRelease().GetFormats() {
		if format.GetName() == "LP" {
			return false
		}
		if format.GetName() == "CD" || format.GetName() == "CDr" {
			isCD = true
		}
		for _, desc := range format.GetDescriptions() {
			if desc == "LP" || desc == "12\"" || desc == "10\"" {
				return false
			}
			if desc == "CD" {
				isCD = true
			}
		}
	}
	return isCD
}

func isSeven(record *pbrc.Record) bool {
	isSeven := false
	for _, format := range record.GetRelease().GetFormats() {
		if format.GetName() == "LP" {
			return false
		}
		if format.GetName() == "CD" || format.GetName() == "CDr" {
			return false
		}
		if format.GetName() == "7\"" {
			isSeven = true
		}
		for _, desc := range format.GetDescriptions() {
			if desc == "LP" || desc == "12\"" || desc == "10\"" {
				return false
			}
			if desc == "CD" {
				return false
			}
			if desc == "7\"" {
				isSeven = true
			}
		}
	}
	return isSeven
}

func (s *Server) moveRecordInternal(ctx context.Context, record *pbrc.Record) error {
	folder, rule := s.moveRecord(ctx, record)

	s.CtxLog(ctx, fmt.Sprintf("%v", record.GetRelease().GetFormats()))
	s.CtxLog(ctx, fmt.Sprintf("%v -> %v, %v", record.GetRelease().GetInstanceId(), folder, rule))

	// Move from LP to cleaning pile if required
	if time.Since(time.Unix(record.GetMetadata().GetLastCleanDate(), 0)).Hours() > 3*365*24 {
		if folder == 812802 || folder == 7651472 || folder == 7665013 {
			if record.GetMetadata().GetFiledUnder() == pbrc.ReleaseMetadata_FILE_12_INCH || record.GetMetadata().GetFiledUnder() == pbrc.ReleaseMetadata_FILE_7_INCH {
				s.CtxLog(ctx, fmt.Sprintf("Moving to clean %v", record.GetRelease().GetInstanceId()))
				folder = 3386035
			}
		}
	}

	if record.GetRelease().GetFolderId() == folder {
		return nil
	}

	s.CtxLog(ctx, fmt.Sprintf("%v -> %v, %v", record.GetRelease().GetInstanceId(), folder, rule))

	if folder > 0 || len(rule) > 0 {
		s.CtxLog(ctx, fmt.Sprintf("MOVED: %v -> %v, %v", record.GetRelease().GetInstanceId(), folder, rule))
	}
	if folder > 0 {
		s.addToArchive(ctx, &pb.RecordedMove{
			InstanceId: record.GetRelease().GetInstanceId(),
			From:       record.GetRelease().GetFolderId(),
			To:         folder,
			MoveTime:   time.Now().Unix(),
			Rule:       rule,
		})
		if record.GetRelease().GetFolderId() != folder {
			s.CtxLog(ctx, fmt.Sprintf("Moving %v to %v", record.GetRelease().GetInstanceId(), folder))

			err := s.getter.update(ctx, record.GetRelease().GetInstanceId(), rule, folder)
			if err != nil {
				return err
			}
		}
		s.incrementCount(ctx, record.GetRelease().GetInstanceId())
		return nil
	}

	if len(rule) > 0 {
		if strings.Contains(rule, "Missing match") {
			return status.Errorf(codes.FailedPrecondition, "Temp unable to move: %v", rule)
		}
		s.CtxLog(ctx, fmt.Sprintf("Unable to move record: %v", rule))
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

	return nil
}

func (s *Server) moveRecord(ctx context.Context, r *pbrc.Record) (int32, string) {

	// Prevent unclean records from moving out of the cleaning pile
	if r.GetMetadata().GetCategory() != pbrc.ReleaseMetadata_PARENTS {
		if (r.GetMetadata().GetFiledUnder() == pbrc.ReleaseMetadata_FILE_12_INCH || r.GetMetadata().GetFiledUnder() == pbrc.ReleaseMetadata_FILE_7_INCH) &&
			r.GetMetadata().GetSaleState() != pbgd.SaleState_SOLD && r.GetMetadata().GetCategory() != pbrc.ReleaseMetadata_UNKNOWN && r.GetMetadata().GetCategory() != pbrc.ReleaseMetadata_SOLD_ARCHIVE {
			if r.GetRelease().GetFolderId() == 3386035 && time.Since(time.Unix(r.GetMetadata().GetLastCleanDate(), 0)) > time.Hour*24*365*3 {
				return -1, "STILL_NOT_CLEAN"
			}
		}
	}

	if r.GetMetadata().GetBoxState() == pbrc.ReleaseMetadata_IN_THE_BOX && (r.GetRelease().FolderId != 3282985 && r.GetMetadata().MoveFolder != 3282985) {
		return 3282985, "BOX_IT_UP"
	}

	if r.GetMetadata().GetBoxState() == pbrc.ReleaseMetadata_IN_45S_BOX && (r.GetRelease().FolderId != 3291655 && r.GetMetadata().MoveFolder != 3291655) {
		return 3291655, "BOX_IT_UP"
	}

	if r.GetMetadata().GetBoxState() == pbrc.ReleaseMetadata_IN_CDS_BOX && (r.GetRelease().FolderId != 3291970 && r.GetMetadata().MoveFolder != 3291970) {
		return 3291970, "BOX_IT_UP"
	}

	if r.GetMetadata().GetBoxState() == pbrc.ReleaseMetadata_IN_TAPE_BOX && (r.GetRelease().FolderId != 3299890 && r.GetMetadata().MoveFolder != 3299890) {
		return 3299890, "BOX_IT_UP"
	}

	if r.GetMetadata().GetBoxState() == pbrc.ReleaseMetadata_IN_DIGITAL_BOX && (r.GetRelease().FolderId != 3358141 && r.GetMetadata().MoveFolder != 3358141) {
		return 3358141, "BOX_IT_UP"
	}

	if r.GetMetadata().GetBoxState() == pbrc.ReleaseMetadata_IN_BOXSET_BOX && (r.GetRelease().FolderId != 3499126 && r.GetMetadata().MoveFolder != 3499126) {
		return 3499126, "BOX_IT_UP"
	}

	s.CtxLog(ctx, fmt.Sprintf("%v, %v, %v", r.GetMetadata().GetBoxState(), r.GetRelease().GetFolderId(), r.GetMetadata().GetMoveFolder()))

	// Don't move a record that's in the box
	if r.GetMetadata().GetBoxState() != pbrc.ReleaseMetadata_BOX_UNKNOWN && r.GetMetadata().GetBoxState() != pbrc.ReleaseMetadata_OUT_OF_BOX {
		return -1, ""
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNKNOWN && (r.GetRelease().GetFolderId() != 3380098 && r.GetMetadata().GetMoveFolder() != 3380098) {
		return 3380098, "UNKNOWN"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_GOOGLE_PLAY && (r.GetRelease().FolderId != 1433217 && r.GetMetadata().MoveFolder != 1433217) {
		return 1433217, "GPLAY"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_DIGITAL && (r.GetRelease().FolderId != 268147 && r.GetMetadata().MoveFolder != 268147) && r.GetRelease().FolderId != 1433217 {
		return 268147, "DIGITAL"
	}

	// We can always move something for processing.
	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_RIP_THEN_SELL {
		return 812802, "RIP THEN SELL"
	}

	// We can always move something for processing.
	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_ARRIVED {
		if isTwelve(r) {
			return 7651472, "ARRIVED 12"
		}
		if isCD(r) {
			return 7664293, "ARRIVED CD"
		}
		if isSeven(r) {
			return 7665013, "ARRIVED 7"
		}
		return 812802, "ARRIVED"
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

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_VALIDATE {
		if isTwelve(r) {
			return 7651472, "VALIDATING 12"
		}
		if isCD(r) {
			return 7664293, "VALIDATING CD"
		}
		if isSeven(r) {
			return 7665013, "VALIDATING 7"
		}
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

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNLISTENED {
		if isTwelve(r) {
			return 7651472, "UNLISTENED 12"
		}
		if isCD(r) {
			return 7664293, "UNLISTENED CD"
		}
		if isSeven(r) {
			return 7665013, "UNLISTENED 7"
		}
		return 812802, "UNLISTE"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_STAGED {
		if isTwelve(r) {
			return 7651475, "STAGED 12"
		}
		if isCD(r) {
			return 7664296, "STAGED CD"
		}
		if isSeven(r) {
			return 7665016, "STAGED SEVEN"
		}
		return 3578980, "STAGED"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_STALE_SALE {
		return 1708299, "STALE SALE"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_HIGH_SCHOOL {
		if isTwelve(r) {
			return 7651475, "HIGH SCHOOL 12"
		}
		if isCD(r) {
			return 7664296, "HIGH SCHOOL CD"
		}
		return 7665016, "HIGH SCHOOL"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL {
		if isTwelve(r) {
			return 7651472, "PRE HIGH SCHOOL 12"
		}
		if isCD(r) {
			return 7664293, "PRE HIGH SCHOOL CD"
		}
		if isSeven(r) {
			return 7665013, "PHS 7"
		}
		return 812802, "PRE HIGH SCHOOL"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SALE_ISSUE && r.GetRelease().GetFolderId() != 6818839 {
		return 6818839, "SALE_ISSUE"
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_LISTED_TO_SELL {
		if r.GetMetadata().GetFiledUnder() == pbrc.ReleaseMetadata_FILE_12_INCH {
			if r.GetRelease().GetFolderId() != 6803737 {
				return 6803737, "LISTED 12 INCH FOR SALE"
			}
		} else if r.GetMetadata().GetFiledUnder() == pbrc.ReleaseMetadata_FILE_CD {
			if r.GetRelease().GetFolderId() != 6804694 {
				return 6804694, "LISTED CD FOR SALE"
			}
		} else if r.GetMetadata().GetFiledUnder() == pbrc.ReleaseMetadata_FILE_7_INCH {
			if r.GetRelease().GetFolderId() != 6804697 {
				return 6804697, "LISTED 7 INCH FOR SALE"
			}
		} else {
			if r.GetRelease().FolderId != 488127 {
				return 488127, "LSITEND TO SELL"
			}
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_IN_COLLECTION && r.GetRelease().FolderId != 812802 {
		if isTwelve(r) {
			return 7651472, "PRE IN COLLECTION 12"
		}
		if isCD(r) {
			return 7664293, "PRE IN COLLECTION CD"
		}
		if isSeven(r) {
			return 7665013, "PIC 7"
		}
		return 812802, "PRE IN COLLECTION"
	}
	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_IN_COLLECTION {
		if r.GetMetadata().GetGoalFolder() != 0 && (r.GetRelease().FolderId != r.GetMetadata().GetGoalFolder() && r.GetMetadata().MoveFolder != r.GetMetadata().GetGoalFolder()) {
			return r.GetMetadata().GetGoalFolder(), "COLLECTION MOVE"
		}
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
		if isTwelve(r) {
			return 7651472, "STAGED TO SELL 12"
		}
		if isCD(r) {
			return 7664293, "STAGED TO SELL CD"
		}
		if isSeven(r) {
			return 7665013, "STS 7"
		}
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
