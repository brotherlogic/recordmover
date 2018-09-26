package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	pbrc "github.com/brotherlogic/recordcollection/proto"
)

type getter interface {
	getRecords() ([]*pbrc.Record, error)
	update(*pbrc.Record) error
}

func (s *Server) moveRecords(ctx context.Context) {
	records, err := s.getter.getRecords()

	if err != nil {
		return
	}

	count := int64(0)
	for _, record := range records {
		update := s.moveRecord(record)
		if update != nil {
			count++
			err := s.getter.update(update)
			if err != nil {
				s.Log(fmt.Sprintf("Error moving record: %v", err))
			}
		}
	}

	s.lastProc = time.Now()
	s.lastCount = count
}

func (s *Server) canMove(r *pbrc.Record) bool {
	//We can always move to digital
	if r.GetMetadata() != nil && r.GetMetadata().GoalFolder == 268147 {
		return true
	}

	for _, f := range r.GetRelease().GetFormats() {
		if f.Name == "CD" {
			if !s.cdproc.isRipped(r.GetRelease().Id) {
				return false
			}
		}
	}

	return true
}

func (s *Server) moveRecord(r *pbrc.Record) *pbrc.Record {
	if !s.canMove(r) {
		return nil
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
		if r.GetRelease().Rating == 0 && r.GetRelease().FolderId != 812802 {
			r.GetMetadata().MoveFolder = 812802
			return r
		}

		if r.GetRelease().Rating > 0 && r.GetRelease().FolderId != 268147 && r.GetMetadata().MoveFolder != 268147 {
			r.GetMetadata().MoveFolder = 268147
			return r
		}
	}

	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOLD && r.GetRelease().FolderId != 488127 {
		r.GetMetadata().MoveFolder = 488127
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
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_POSTDOC ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_GRADUATE ||
		r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_SOPHMORE {
		if r.GetMetadata().GetGoalFolder() != 0 && (r.GetRelease().FolderId != r.GetMetadata().GetGoalFolder() && r.GetMetadata().MoveFolder != r.GetMetadata().GetGoalFolder()) {
			s.Log(fmt.Sprintf("Setting move folder: %v", r))
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
