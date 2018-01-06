package main

import (
	"errors"
	"testing"

	pbgd "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
)

type testGetter struct {
	rec *pbrc.Record
}

func (t *testGetter) getRecords() ([]*pbrc.Record, error) {
	return []*pbrc.Record{t.rec}, nil
}

func (t *testGetter) update(r *pbrc.Record) error {
	t.rec = r
	return nil
}

type testFailGetter struct {
	grf          bool
	lastCategory pbrc.ReleaseMetadata_Category
}

func (t testFailGetter) getRecords() ([]*pbrc.Record, error) {
	if t.grf {
		return []*pbrc.Record{&pbrc.Record{Release: &pbgd.Release{FolderId: 1}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_UNLISTENED}}}, nil
	}
	return nil, errors.New("Built to fail")
}

func (t testFailGetter) update(r *pbrc.Record) error {
	if !t.grf {
		t.lastCategory = r.GetMetadata().GetCategory()
		return nil
	}
	return errors.New("Built to fail")
}

func InitTest() *Server {
	s := Init()
	s.SkipLog = true
	s.getter = &testGetter{}

	return s
}

func TestEmptyUpdate(t *testing.T) {
	s := InitTest()
	s.moveRecords()
}

func TestBadGetter(t *testing.T) {
	s := InitTest()
	tg := testFailGetter{}
	s.getter = tg
	s.moveRecords()
}

func TestUpdateFailOnUpdate(t *testing.T) {
	s := InitTest()
	tg := testFailGetter{grf: true}
	s.getter = tg
	s.moveRecords()
}

func TestUpdateToStaged(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_UNLISTENED}}}
	s.getter = &tg
	s.moveRecords()

	if tg.rec.GetMetadata().MoveFolder != 812802 {
		t.Errorf("Folder has not been updated: %v", tg.rec)
	}
}
