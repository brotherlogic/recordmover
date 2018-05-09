package main

import (
	"errors"
	"testing"

	"github.com/brotherlogic/keystore/client"
	"golang.org/x/net/context"

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
	s.GoServer.KSclient = *keystoreclient.GetTestClient("./testing")

	return s
}

func TestEmptyUpdate(t *testing.T) {
	s := InitTest()
	s.moveRecords(context.Background())
}

func TestBadGetter(t *testing.T) {
	s := InitTest()
	tg := testFailGetter{}
	s.getter = tg
	s.moveRecords(context.Background())
}

var movetests = []struct {
	in  *pbrc.Record
	out int32
}{
	{&pbrc.Record{Release: &pbgd.Release{}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_ASSESS}}, 1362206},
	{&pbrc.Record{Release: &pbgd.Release{}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_NO_LABELS}}, 1362206},
}

func TestMoves(t *testing.T) {
	for _, test := range movetests {
		s := InitTest()
		tg := testGetter{rec: test.in}
		s.getter = &tg
		s.moveRecords(context.Background())

		if tg.rec.GetMetadata().MoveFolder != test.out {
			t.Fatalf("Error moving record: %v -> %v (ended up in %v)", test.in, test.out, tg.rec.GetMetadata().MoveFolder)
		}
	}
}

func TestUpdateFailOnUpdate(t *testing.T) {
	s := InitTest()
	tg := testFailGetter{grf: true}
	s.getter = tg
	s.moveRecords(context.Background())
}

func TestUpdateToUnlistend(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_UNLISTENED}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 812802 {
		t.Errorf("Folder has not been updated: %v", tg.rec)
	}
}

func TestUpdateToDigital(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{Rating: 0, FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_DIGITAL}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 812802 {
		t.Errorf("Folder has not been updated: %v", tg.rec)
	}
}

func TestUpdateToDigitalDone(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{Rating: 4, FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_DIGITAL}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 268147 {
		t.Errorf("Folder has not been updated: %v", tg.rec)
	}
}

func TestUpdateToSold(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_SOLD}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 488127 {
		t.Errorf("Folder has not been updated: %v", tg.rec)
	}
}

func TestUpdateToStaged(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_STAGED}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 673768 {
		t.Errorf("Folder has not been updated: %v", tg.rec)
	}
}

func TestUpdatePreProfessorToListeningPile(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_PROFESSOR}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 812802 {
		t.Errorf("Pre Freshman has not been updated: %v", tg.rec)
	}
}

func TestUpdateProfessorToPurgatory(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PROFESSOR}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 1362206 {
		t.Errorf("Pre Freshman has not been updated: %v", tg.rec)
	}
}

func TestUpdateProfessorToFilled(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{GoalFolder: 820, Category: pbrc.ReleaseMetadata_PROFESSOR}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 820 {
		t.Errorf("Freshman has not been moved correctly: %v", tg.rec)
	}
}

func TestUpdatePreFreshmanToListeningPile(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_PRE_FRESHMAN}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 812802 {
		t.Errorf("Pre Freshman has not been updated: %v", tg.rec)
	}
}

func TestUpdateFreshmanToPurgatory(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_FRESHMAN}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 1362206 {
		t.Errorf("Pre Freshman has not been updated: %v", tg.rec)
	}
}

func TestUpdateFreshmanToFilled(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{GoalFolder: 820, Category: pbrc.ReleaseMetadata_FRESHMAN}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 820 {
		t.Errorf("Freshman has not been moved correctly: %v", tg.rec)
	}
}
