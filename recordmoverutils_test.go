package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/brotherlogic/keystore/client"
	pb "github.com/brotherlogic/recordmover/proto"
	"golang.org/x/net/context"

	pbgd "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
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

func (t *testGetter) update(ctx context.Context, r *pbrc.Record) error {
	t.rec = r
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
		return &pbrc.Record{Release: &pbgd.Release{FolderId: 1}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_UNLISTENED, GoalFolder: 123, Match: pbrc.ReleaseMetadata_FULL_MATCH}}, nil
	}
	return nil, fmt.Errorf("Built to fail")
}

func (t *testFailGetter) getRecords(ctx context.Context) ([]*pbrc.Record, error) {
	if t.grf {
		return []*pbrc.Record{&pbrc.Record{Release: &pbgd.Release{FolderId: 1}, Metadata: &pbrc.ReleaseMetadata{Category: pbrc.ReleaseMetadata_UNLISTENED, GoalFolder: 123}}}, nil
	}
	return nil, errors.New("Built to fail")
}

func (t *testFailGetter) update(ctx context.Context, r *pbrc.Record) error {
	if !t.grf {
		t.lastCategory = r.GetMetadata().GetCategory()
		return nil
	}
	return errors.New("Built to fail")
}

func InitTest() *Server {
	s := Init()
	s.SkipLog = true
	s.SkipIssue = true
	s.getter = &testGetter{rec: &pbrc.Record{Release: &pbgd.Release{InstanceId: 1}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, GoalFolder: 2}}}
	s.GoServer.KSclient = *keystoreclient.GetTestClient("./testing")
	s.cdproc = &testRipper{}
	s.organiser = &testOrg{failCount: 100}
	s.testing = true

	return s
}

func TestEmptyUpdate(t *testing.T) {
	s := InitTest()
	s.moveRecords(context.Background())
}

func TestBadGetter(t *testing.T) {
	s := InitTest()
	tg := &testFailGetter{}
	s.getter = tg
	s.moveRecords(context.Background())
}

var movetests = []struct {
	in  *pbrc.Record
	out int32
}{
	{&pbrc.Record{Release: &pbgd.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_PARENTS, GoalFolder: 1234}}, 1727264},
	{&pbrc.Record{Release: &pbgd.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_ASSESS, GoalFolder: 1234}}, 1362206},
	{&pbrc.Record{Release: &pbgd.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_NO_LABELS, GoalFolder: 1234}}, 1362206},
	{&pbrc.Record{Release: &pbgd.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_GOOGLE_PLAY, GoalFolder: 1234}}, 1433217},
	{&pbrc.Record{Release: &pbgd.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_SOLD_ARCHIVE, GoalFolder: 1234}}, 1613206},
	{&pbrc.Record{Release: &pbgd.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_ASSESS_FOR_SALE, GoalFolder: 1234}}, 1362206},
	{&pbrc.Record{Release: &pbgd.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_PRE_HIGH_SCHOOL, GoalFolder: 1234}}, 673768},
	{&pbrc.Record{Release: &pbgd.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_LISTED_TO_SELL, GoalFolder: 1234}}, 488127},
	{&pbrc.Record{Release: &pbgd.Release{}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_HIGH_SCHOOL, GoalFolder: 1234}}, 673768},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 4321}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_PRE_SOPHMORE, GoalFolder: 1234}}, 1234},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 1249, Rating: 5, InstanceId: 19867493}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, DateAdded: 1368884100, FilePath: "1450170", LastCache: 1543338069, Category: pbrc.ReleaseMetadata_STAGED_TO_SELL, GoalFolder: 242018, LastSyncTime: 1544561649, Purgatory: pbrc.Purgatory_NEEDS_STOCK_CHECK, LastStockCheck: 1544490181, OverallScore: 4}}, 812802},
	{&pbrc.Record{Release: &pbgd.Release{FolderId: 4321}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_STALE_SALE, GoalFolder: 1234}}, 1708299},
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
	tg := &testFailGetter{grf: true}
	s.getter = tg
	s.moveRecords(context.Background())
}

func TestUpdateFailOnUpdateRecordPull(t *testing.T) {
	s := InitTest()
	tg := &testGetter{failGet: true, rec: &pbrc.Record{Release: &pbgd.Release{InstanceId: 12}}}
	s.getter = tg
	s.moveRecords(context.Background())
}

func TestUpdateFailOnGet(t *testing.T) {
	s := InitTest()
	tg := &testFailGetter{grf: true}
	s.getter = tg
	s.moveRecords(context.Background())
}

func TestUpdateToUnlistend(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_UNLISTENED, GoalFolder: 1234}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 812802 {
		t.Errorf("Folder has not been updated: %v", tg.rec)
	}
}

func TestUpdateToDigitalDone(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{Rating: 4, FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_DIGITAL, GoalFolder: 1234}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 268147 {
		t.Errorf("Folder has not been updated: %v", tg.rec)
	}
}

func TestUpdateToSold(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_SOLD, GoalFolder: 1234}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 488127 {
		t.Errorf("Folder has not been updated: %v", tg.rec)
	}
}

func TestUpdateToStaged(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_STAGED, GoalFolder: 1234}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 673768 {
		t.Errorf("Folder has not been updated: %v", tg.rec)
	}
}

func TestMoveUnripped(t *testing.T) {
	s := InitTest()
	val, _ := s.moveRecord(context.Background(), &pbrc.Record{Release: &pbgd.Release{Id: 123, Formats: []*pbgd.Format{&pbgd.Format{Name: "CD"}}}, Metadata: &pbrc.ReleaseMetadata{}})

	if val != nil {
		t.Errorf("moved: %v", val)
	}
}

func TestMoveUnrippedButDigital(t *testing.T) {
	s := InitTest()
	val, _ := s.moveRecord(context.Background(), &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{GoalFolder: 268147}, Release: &pbgd.Release{Id: 123, Formats: []*pbgd.Format{&pbgd.Format{Name: "CD"}}}})

	if val != nil {
		t.Errorf("moved: %v", val)
	}
}

func TestUpdateProfessorToFilled(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, GoalFolder: 820, Category: pbrc.ReleaseMetadata_PROFESSOR}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 820 {
		t.Errorf("Freshman has not been moved correctly: %v", tg.rec)
	}
}

func TestUpdatePreFreshmanToListeningPile(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_PRE_FRESHMAN, GoalFolder: 1234}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 812802 {
		t.Errorf("Pre Freshman has not been updated: %v", tg.rec)
	}
}

func TestUpdateFreshmanToFilled(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, GoalFolder: 820, Category: pbrc.ReleaseMetadata_FRESHMAN}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 820 {
		t.Errorf("Freshman has not been moved correctly: %v", tg.rec)
	}
}

func TestUpdateRipThenSellToListeningPile(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{InstanceId: 12, FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, GoalFolder: 820, Category: pbrc.ReleaseMetadata_RIP_THEN_SELL}}}
	s.getter = &tg
	s.moveRecords(context.Background())

	if tg.rec.GetMetadata().MoveFolder != 812802 {
		t.Errorf("RIP THEN SELL has not been moved correctly: %v", tg.rec)
	}

	moves, err := s.ListMoves(context.Background(), &pb.ListRequest{InstanceId: 12})
	if err != nil {
		t.Fatalf("ERR: %v", err)
	}
	if len(moves.GetArchives()) == 0 {
		t.Errorf("Bad archival recording: %v", moves)
	}
}

func TestUpdateStagedToSellToListeningPile(t *testing.T) {
	s := InitTest()
	tg := testGetter{rec: &pbrc.Record{Release: &pbgd.Release{FolderId: 812}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, Category: pbrc.ReleaseMetadata_STAGED_TO_SELL, GoalFolder: 1234}}}
	s.getter = &tg
	s.moveRecords(context.Background())
	if tg.rec.GetMetadata().MoveFolder != 812802 {
		t.Errorf("Pre Freshman has not been updated: %v", tg.rec)
	}
}

func TestUpdateMove(t *testing.T) {
	s := InitTest()
	s.config.Moves = append(s.config.Moves, &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3})

	s.refreshMoves(context.Background())

	if s.config.Moves[0].AfterContext.Location == "" {
		t.Errorf("Update not run")
	}
}

func TestUpdateMoveWithFlip(t *testing.T) {
	s := InitTest()
	s.organiser = &testOrg{flipLocate: true}
	s.config.Moves = append(s.config.Moves, &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3})

	s.refreshMoves(context.Background())

	if s.config.Moves[0].AfterContext.Location == "" {
		t.Errorf("Update not run")
	}
}

func TestReverseUpdateMoveWithFlip(t *testing.T) {
	s := InitTest()
	s.organiser = &testOrg{rflipLocate: true}
	s.config.Moves = append(s.config.Moves, &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3})

	s.refreshMoves(context.Background())

	if s.config.Moves[0].AfterContext.Location == "" {
		t.Errorf("Update not run")
	}
}

func TestUpdateMoveFailLocate(t *testing.T) {
	s := InitTest()
	s.organiser = &testOrg{failLocate: true}
	s.config.Moves = append(s.config.Moves, &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3})

	s.refreshMoves(context.Background())

	if s.config.Moves[0].AfterContext != nil {
		t.Errorf("Update run")
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

func TestCanMoveCDWithNoMatch(t *testing.T) {
	s := InitTest()

	if s.canMove(context.Background(), &pbrc.Record{Release: &pbgd.Release{Formats: []*pbgd.Format{&pbgd.Format{Name: "CD"}}}, Metadata: &pbrc.ReleaseMetadata{GoalFolder: 1234}}) == nil {
		t.Errorf("Should not be able to move dirty record")
	}
}

func TestCanMoveCD(t *testing.T) {
	s := InitTest()

	if s.canMove(context.Background(), &pbrc.Record{Release: &pbgd.Release{Formats: []*pbgd.Format{&pbgd.Format{Name: "CD"}}}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH, GoalFolder: 1234}}) == nil {
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
