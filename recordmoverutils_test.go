package main

import (
	"errors"
	"testing"

	pbgd "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
)

type testGetter struct {
	lastCategory *pbrc.ReleaseMetadata_Category
	rec          *pbrc.Record
}

func (t *testGetter) getRecords() ([]*pbrc.Record, error) {
	return []*pbrc.Record{t.rec}, nil
}

func (t *testGetter) update(r *pbrc.Record) error {
	t.lastCategory = &r.GetMetadata().Category
	return nil
}

type testFailGetter struct {
	grf          bool
	lastCategory pbrc.ReleaseMetadata_Category
}

func (t testFailGetter) getRecords() ([]*pbrc.Record, error) {
	if t.grf {
		return []*pbrc.Record{&pbrc.Record{Release: &pbgd.Release{FolderId: 1}}}, nil
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
