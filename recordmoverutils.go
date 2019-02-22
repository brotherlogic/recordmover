package main

import (
	"fmt"
	"time"

	pbgd "github.com/brotherlogic/godiscogs"
	pb "github.com/brotherlogic/recordmover/proto"
	pbro "github.com/brotherlogic/recordsorganiser/proto"
	"golang.org/x/net/context"

	pbrc "github.com/brotherlogic/recordcollection/proto"
)

type getter interface {
	getRecords(ctx context.Context) ([]*pbrc.Record, error)
	update(ctx context.Context, rec *pbrc.Record) error
}

func (s *Server) refreshMoves(ctx context.Context) {
	//Clean in folder moves
	for key, val := range s.moves {
		if val.ToFolder == val.FromFolder {
			delete(s.moves, key)
		}
	}

	for _, r := range s.moves {
		if time.Now().Sub(time.Unix(r.LastUpdate, 0)) > time.Hour {
			err := s.refreshMove(ctx, r)
			if err == nil {
				s.Log(fmt.Sprintf("Refreshing %v", r.InstanceId))
				s.saveMoves(ctx)
				return
			}
			s.Log(fmt.Sprintf("Failed refresh of %v -> %v", r.InstanceId, err))
		}
	}
}

func (s *Server) refreshMove(ctx context.Context, move *pb.RecordMove) error {
	location, err := s.organiser.locate(ctx, &pbro.LocateRequest{InstanceId: move.Record.GetRelease().InstanceId})

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
				resp, err := s.recordcollection.getRecords(ctx, &pbrc.GetRecordsRequest{Filter: &pbrc.Record{Release: &pbgd.Release{InstanceId: location.GetFoundLocation().GetReleasesLocation()[i-1].InstanceId}}})

				if err != nil {
					return err
				}

				if len(resp.GetRecords()) != 1 {
					return fmt.Errorf("Ambigous move")
				}

				move.GetAfterContext().Before = resp.GetRecords()[0]
			} else {
				move.AfterContext.Before = &pbrc.Record{Release: &pbgd.Release{Title: "START_OF_SLOT"}}
			}

			if i < len(location.GetFoundLocation().GetReleasesLocation())-1 {
				move.GetAfterContext().Location = location.GetFoundLocation().Name
				move.GetAfterContext().Slot = location.GetFoundLocation().GetReleasesLocation()[i].Slot

				resp, err := s.recordcollection.getRecords(ctx, &pbrc.GetRecordsRequest{Filter: &pbrc.Record{Release: &pbgd.Release{InstanceId: location.GetFoundLocation().GetReleasesLocation()[i+1].InstanceId}}})

				if err != nil {
					return err
				}

				if len(resp.GetRecords()) != 1 {
					return fmt.Errorf("Ambigous move")
				}
				move.GetAfterContext().After = resp.GetRecords()[0]

			} else {
				move.AfterContext.After = &pbrc.Record{Release: &pbgd.Release{Title: "END_OF_SLOT"}}
			}

		}
	}

	if move.GetAfterContext() != nil && move.GetAfterContext().Location != "" {
		s.updateArchive(&pb.RecordedMove{InstanceId: move.InstanceId, MoveLocation: move.GetAfterContext().Location, MoveTime: time.Now().Unix()})
	}

	move.LastUpdate = time.Now().Unix()
	return nil
}

func (s *Server) moveRecords(ctx context.Context) {
	records, err := s.getter.getRecords(ctx)

	if err != nil {
		return
	}

	count := int64(0)
	miss := 0
	for _, record := range records {
		update := s.moveRecord(ctx, record)
		if update != nil {
			count++
			err := s.getter.update(ctx, update)
			if err != nil {
				s.Log(fmt.Sprintf("Error moving record: %v", err))
			} else {
				break
			}
		} else {
			miss++
		}
	}

	s.lastProc = time.Now()
	s.lastCount = count
}

func (s *Server) canMove(ctx context.Context, r *pbrc.Record) bool {
	//We can always move to digital
	if r.GetMetadata() != nil && r.GetMetadata().GoalFolder == 268147 {
		return true
	}

	for _, f := range r.GetRelease().GetFormats() {
		if f.Name == "CD" {
			if len(r.GetMetadata().FilePath) == 0 {
				return false
			}
		}
	}

	return true
}

func (s *Server) moveRecord(ctx context.Context, r *pbrc.Record) *pbrc.Record {
	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_GOOGLE_PLAY && (r.GetRelease().FolderId != 1433217 && r.GetMetadata().MoveFolder != 1433217) {
		r.GetMetadata().MoveFolder = 1433217
		return r
	}

	if !s.canMove(ctx, r) {
		return nil
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_RIP_THEN_SELL && (r.GetRelease().FolderId != 812802 && r.GetMetadata().MoveFolder != 812802) {
		r.GetMetadata().MoveFolder = 812802
		return r
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_ASSESS_FOR_SALE && (r.GetRelease().FolderId != 1362206 && r.GetMetadata().MoveFolder != 1362206) {
		r.GetMetadata().MoveFolder = 1362206
		r.GetMetadata().Purgatory = pbrc.Purgatory_NEEDS_STOCK_CHECK
		return r
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_ASSESS && (r.GetRelease().FolderId != 1362206 && r.GetMetadata().MoveFolder != 1362206) {
		r.GetMetadata().MoveFolder = 1362206
		r.GetMetadata().Purgatory = pbrc.Purgatory_NEEDS_STOCK_CHECK
		return r
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_NO_LABELS && (r.GetRelease().FolderId != 1362206 && r.GetMetadata().MoveFolder != 1362206) {
		r.GetMetadata().MoveFolder = 1362206
		r.GetMetadata().Purgatory = pbrc.Purgatory_NEEDS_LABELS
		return r
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_DIGITAL {
		if r.GetRelease().FolderId != 268147 && r.GetMetadata().MoveFolder != 268147 {
			r.GetMetadata().MoveFolder = 268147
			return r
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOLD && r.GetRelease().FolderId != 488127 {
		r.GetMetadata().MoveFolder = 488127
		return r
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOLD_ARCHIVE && r.GetRelease().FolderId != 1613206 {
		r.GetMetadata().MoveFolder = 1613206
		return r
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNLISTENED && r.GetRelease().FolderId != 812802 {
		r.GetMetadata().MoveFolder = 812802
		return r
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_STAGED && r.GetRelease().FolderId != 673768 {
		r.GetMetadata().MoveFolder = 673768
		return r
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_HIGH_SCHOOL && r.GetRelease().FolderId != 673768 {
		r.GetMetadata().MoveFolder = 673768
		return r
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL && r.GetRelease().FolderId != 812802 {
		r.GetMetadata().MoveFolder = 812802
		return r
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_LISTED_TO_SELL && r.GetRelease().FolderId != 488127 {
		r.GetMetadata().MoveFolder = 488127
		return r
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_FRESHMAN && r.GetRelease().FolderId != 812802 {
		r.GetMetadata().MoveFolder = 812802
		return r
	}
	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_FRESHMAN {
		if r.GetMetadata().GetGoalFolder() != 0 && (r.GetRelease().FolderId != r.GetMetadata().GetGoalFolder() && r.GetMetadata().MoveFolder != r.GetMetadata().GetGoalFolder()) {
			r.GetMetadata().MoveFolder = r.GetMetadata().GetGoalFolder()
			return r
		}
		if r.GetMetadata().GetGoalFolder() == 0 && r.GetRelease().FolderId != 1362206 {
			r.GetMetadata().MoveFolder = 1362206
			return r
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_STAGED_TO_SELL && r.GetRelease().FolderId != 812802 && r.GetMetadata().MoveFolder != 812802 {
		r.GetMetadata().MoveFolder = 812802
		return r
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PROFESSOR ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_PROFESSOR ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_POSTDOC ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_POSTDOC ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_GRADUATE ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_GRADUATE ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOPHMORE ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_PRE_SOPHMORE {
		if r.GetMetadata().GetGoalFolder() != 0 && (r.GetRelease().FolderId != r.GetMetadata().GetGoalFolder() && r.GetMetadata().MoveFolder != r.GetMetadata().GetGoalFolder()) {
			r.GetMetadata().MoveFolder = r.GetMetadata().GetGoalFolder()
			return r
		}
		if r.GetMetadata().GetGoalFolder() == 0 && r.GetRelease().FolderId != 1362206 {
			r.GetMetadata().MoveFolder = 1362206
			return r
		}
	}

	return nil
}

func (s *Server) lookForStale(ctx context.Context) {

	for _, move := range s.moves {
		if time.Now().Sub(time.Unix(move.MoveDate, 0)) > time.Hour*24*7 {
			s.RaiseIssue(ctx, "Stale Move", fmt.Sprintf("Move has been stuck for over a week: %v", move.InstanceId), false)
		}
	}
}
