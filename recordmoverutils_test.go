package main

import (
	"errors"
	"fmt"
	"testing"

	keystoreclient "github.com/brotherlogic/keystore/client"
	"golang.org/x/net/context"

	gdpb "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordmover/proto"
)

type testRipper struct {
	ripped bool
}

func (t *testRipper) isRipped(ctx context.Context, ID int32) bool {
	return t.ripped
}

type testGetter struct {
	rec     *pbrc.Record
	failGet bool
}

func (t *testGetter) getRecordsSince(ctx context.Context, since int64) ([]int32, error) {
	return []int32{t.rec.GetRelease().InstanceId}, nil
}

func (t *testGetter) getRecord(ctx context.Context, instanceID int32) (*pbrc.Record, error) {
	if t.failGet {
		return nil, fmt.Errorf("Error getting record")
	}
	return t.rec, nil
}

func (t *testGetter) getRecords(ctx context.Context) ([]*pbrc.Record, error) {
	return []*pbrc.Record{t.rec}, nil
}

func (t *testGetter) update(ctx context.Context, instanceID int32, reason string, moveFolder int32) error {
	t.rec = &pbrc.Record{Release: &gdpb.Release{InstanceId: instanceID}, Metadata: &pbrc.ReleaseMetadata{MoveFolder: moveFolder}}
	return nil
}

type testFailGetter struct {
	grf          bool
	lastCategory pbrc.ReleaseMetadata_Category
}

func (t *testFailGetter) getRecordsSince(ctx context.Context, since int64) ([]int32, error) {
	if t.grf {
		return []int32{int32(12)}, nil
	}
	return []int32{}, fmt.Errorf("Built to fail")
}

func (t *testFailGetter) getRecord(ctx context.Context, instanceID int32) (*pbrc.Record, error) {
	if t.grf {
		return &pbrc.Record{Release: &gdpb.Release{FolderId: 1}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_UNLISTENED, GoalFolder: 123, Match: pbrc.ReleaseMetadata_FULL_MATCH}}, nil
	}
	return nil, fmt.Errorf("Built to fail")
}

func (t *testFailGetter) getRecords(ctx context.Context) ([]*pbrc.Record, error) {
	if t.grf {
		return []*pbrc.Record{&pbrc.Record{Release: &gdpb.Release{FolderId: 1}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_UNLISTENED, GoalFolder: 123}}}, nil
	}
	return nil, errors.New("Built to fail")
}

func (t *testFailGetter) update(ctx context.Context, instanceID int32, reason string, moveFolder int32) error {
	if !t.grf {
		return nil
	}
	return errors.New("Built to fail")
}

func InitTest() *Server {
	s := Init()
	s.SkipLog = true
	s.SkipIssue = true
	s.getter = &testGetter{rec: &pbrc.Record{Release: &gdpb.Release{InstanceId: 1}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, GoalFolder: 2}}}
	s.GoServer.KSclient = *keystoreclient.GetTestClient("./testing")
	s.GoServer.KSclient.Save(context.Background(), ConfigKey, &pb.Config{})
	s.cdproc = &testRipper{}
	s.organiser = &testOrg{failCount: 100}
	s.testing = true

	return s
}

func TestBadGetter(t *testing.T) {
	s := InitTest()
	tg := &testFailGetter{}
	s.getter = tg
}

var movetests = []struct {
	in  *pbrc.Record
	out int32
}{
	{&pbrc.Record{Release: &gdpb.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_PARENTS, GoalFolder: 1234}}, 1727264},
	{&pbrc.Record{Release: &gdpb.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_ASSESS, GoalFolder: 1234}}, 1362206},
	{&pbrc.Record{Release: &gdpb.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_NO_LABELS, GoalFolder: 1234}}, 1362206},
	{&pbrc.Record{Release: &gdpb.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_GOOGLE_PLAY, GoalFolder: 1234}}, 1433217},
	{&pbrc.Record{Release: &gdpb.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_SOLD_ARCHIVE, GoalFolder: 1234}}, 1613206},
	{&pbrc.Record{Release: &gdpb.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_ASSESS_FOR_SALE, GoalFolder: 1234}}, 1362206},
	{&pbrc.Record{Release: &gdpb.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL, GoalFolder: 1234}}, 673768},
	{&pbrc.Record{Release: &gdpb.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_LISTED_TO_SELL, GoalFolder: 1234}}, 488127},
	{&pbrc.Record{Release: &gdpb.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_HIGH_SCHOOL, GoalFolder: 1234}}, 673768},
	{&pbrc.Record{Release: &gdpb.Release{FolderId: 2259637}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_PRE_VALIDATE, GoalFolder: 2259637}}, 812802},
	{&pbrc.Record{Release: &gdpb.Release{FolderId: 4321}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_PRE_SOPHMORE, GoalFolder: 1234}}, 1234},
	{&pbrc.Record{Release: &gdpb.Release{FolderId: 1249, Rating: 5, InstanceId: 19867493}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, DateAdded: 1368884100, FilePath: "1450170", LastCache: 1543338069, Category: pbrc.ReleaseMetadata_STAGED_TO_SELL, GoalFolder: 242018, LastSyncTime: 1544561649, Purgatory: pbrc.Purgatory_NEEDS_STOCK_CHECK, LastStockCheck: 1544490181, OverallScore: 4}}, 812802},
	{&pbrc.Record{Release: &gdpb.Release{FolderId: 4321}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_STALE_SALE, GoalFolder: 1234}}, 1708299},
	{&pbrc.Record{Release: &gdpb.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_UNLISTENED, GoalFolder: 1234}}, 812802},
	{&pbrc.Record{Release: &gdpb.Release{Rating: 4, FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_DIGITAL, GoalFolder: 1234}}, 268147},
	{&pbrc.Record{Release: &gdpb.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_SOLD, GoalFolder: 1234}}, 488127},
	{&pbrc.Record{Release: &gdpb.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_STAGED, GoalFolder: 1234}}, 673768},
	{&pbrc.Record{Release: &gdpb.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, GoalFolder: 820, Category: pbrc.ReleaseMetadata_PROFESSOR}}, 820},
	{&pbrc.Record{Release: &gdpb.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_PRE_FRESHMAN, GoalFolder: 1234}}, 812802},
	{&pbrc.Record{Release: &gdpb.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, GoalFolder: 820, Category: pbrc.ReleaseMetadata_FRESHMAN}}, 820},
	{&pbrc.Record{Release: &gdpb.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_STAGED_TO_SELL, GoalFolder: 1234}}, 812802},
}

func TestMoves(t *testing.T) {
	for _, test := range movetests {
		s := InitTest()
		tg := testGetter{rec: test.in}
		s.getter = &tg

		s.ClientUpdate(context.Background(), &pbrc.ClientUpdateRequest{})

		if tg.rec.GetMetadata().MoveFolder != test.out {
			t.Fatalf("Error moving record: %v -> %v (ended up in %v)", test.in, test.out, tg.rec.GetMetadata().MoveFolder)
		}
	}
}

func TestMoveUnripped(t *testing.T) {
	s := InitTest()
	val, _ := s.moveRecord(context.Background(), &pbrc.Record{Release: &gdpb.Release{Id: 123, Formats: []*gdpb.Format{&gdpb.Format{Name: "CD"}}}, Metadata: &pbrc.ReleaseMetadata{}})

	if val > 0 {
		t.Errorf("moved: %v", val)
	}
}

func TestMoveUnrippedButDigital(t *testing.T) {
	s := InitTest()
	val, _ := s.moveRecord(context.Background(), &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{GoalFolder: 268147}, Release: &gdpb.Release{Id: 123, Formats: []*gdpb.Format{&gdpb.Format{Name: "CD"}}}})

	if val > 0 {
		t.Errorf("moved: %v", val)
	}
}

func TestUpdateRipThenSellToListeningPile(t *testing.T) {
	s := InitTest()
	s.testing = false
	tg := testGetter{rec: &pbrc.Record{Release: &gdpb.Release{InstanceId: 12, FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, GoalFolder: 820, Category: pbrc.ReleaseMetadata_RIP_THEN_SELL}}}
	s.getter = &tg
	s.ClientUpdate(context.Background(), &pbrc.ClientUpdateRequest{})

	if tg.rec.GetMetadata().MoveFolder != 812802 {
		t.Errorf("RIP THEN SELL has not been moved correctly: %v", tg.rec)
	}

	_, err := s.ListMoves(context.Background(), &pb.ListRequest{InstanceId: 12})
	if err != nil {
		t.Fatalf("ERR: %v", err)
	}
}

func TestTriggerIncrement(t *testing.T) {
	s := InitTest()

	errCount := 0
	for i := 0; i < 200; i++ {
		err := s.incrementCount(context.Background(), int32(12))
		if err != nil {
			errCount++
		}
	}

	if errCount == 0 {
		t.Errorf("Triggered Increment")
	}
}

func TestNoTriggerIncrement(t *testing.T) {
	s := InitTest()

	errCount := 0
	for i := 0; i < 200; i++ {
		err := s.incrementCount(context.Background(), int32(i))
		if err != nil {
			errCount++
		}
	}

	if errCount != 0 {
		t.Errorf("An increment was triggered incorrectly")
	}
}

func TestCanMoveFail(t *testing.T) {
	s := InitTest()

	if s.canMove(context.Background(), &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{Dirty: true}}) == nil {
		t.Errorf("Should not be able to move dirty record")
	}
}

func TestCanMoveFailInternal(t *testing.T) {
	s := InitTest()

	if s.moveRecordInternal(context.Background(), &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{Dirty: true}}) == nil {
		t.Errorf("Should not be able to move dirty record")
	}
}

func TestCanMoveFailInternalNoMatch(t *testing.T) {
	s := InitTest()

	if s.moveRecordInternal(context.Background(), &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{GoalFolder: 1234}}) == nil {
		t.Errorf("Should not be able to move dirty record")
	}
}

func TestCanMoveCDWithNoMatch(t *testing.T) {
	s := InitTest()

	if s.canMove(context.Background(), &pbrc.Record{Release: &gdpb.Release{Formats: []*gdpb.Format{&gdpb.Format{Name: "CD"}}}, Metadata: &pbrc.ReleaseMetadata{GoalFolder: 1234}}) == nil {
		t.Errorf("Should not be able to move dirty record")
	}
}

func TestCanMoveCD(t *testing.T) {
	s := InitTest()

	if s.canMove(context.Background(), &pbrc.Record{Release: &gdpb.Release{Formats: []*gdpb.Format{&gdpb.Format{Name: "CD"}}}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, GoalFolder: 1234}}) == nil {
		t.Errorf("Should not be able to move dirty record")
	}
}

func TestMoveFailOnAfterLocatePull(t *testing.T) {
	s := InitTest()
	s.organiser = &testOrg{failLocate: true}
	err := s.refreshMove(context.Background(), &pb.RecordMove{InstanceId: 12, BeforeContext: &pb.Context{Location: "blah"}})
	if err == nil {
		t.Errorf("Should have failed")
	}
}

func TestAddToArchiveFail(t *testing.T) {
	s := InitTest()
	s.testing = true

	err := s.addToArchive(context.Background(), &pb.RecordedMove{})
	if err == nil {
		t.Errorf("Should have failed")
	}
}
