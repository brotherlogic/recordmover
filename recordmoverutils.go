package main

import (
	"fmt"
	"log"
	"time"

	pbrc "github.com/brotherlogic/recordcollection/proto"
)

type getter interface {
	getRecords() ([]*pbrc.Record, error)
	update(*pbrc.Record) error
}

func (s *Server) moveRecords() {
	t := time.Now()
	records, err := s.getter.getRecords()

	log.Printf("Now %v and %v", records, err)

	if err != nil {
		s.Log(fmt.Sprintf("Error moving records: %v", err))
		return
	}

	s.Log(fmt.Sprintf("About to move %v records", len(records)))
	count := int64(0)
	for _, record := range records {
		update := s.moveRecord(record)
		if update != nil {
			count++
			log.Printf("HERE")
			err := s.getter.update(update)
			if err != nil {
				s.Log(fmt.Sprintf("Error moving record: %v", err))
			}
		}
	}

	s.lastProc = time.Now()
	s.lastCount = count
	s.Log(fmt.Sprintf("Moved %v records (touched %v) in %v", len(records), count, time.Now().Sub(t)))
}

func (s *Server) moveRecord(r *pbrc.Record) *pbrc.Record {
	if r.GetMetadata().GetCategory() == pbrc.ReleaseMetadata_UNLISTENED && r.GetRelease().FolderId != 812802 {
		r.GetMetadata().MoveFolder = 812802
		return r
	}
	return nil
}
