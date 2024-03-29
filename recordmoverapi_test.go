package main

import (
	"fmt"
	"testing"

	pbgd "github.com/brotherlogic/godiscogs/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordmover/proto"
	pbro "github.com/brotherlogic/recordsorganiser/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type testCol struct {
	fail           bool
	failSecond     bool
	noLocate       bool
	noLocateSecond bool
	count          int
}

func (t *testCol) getRecords(ctx context.Context, rec *pbrc.GetRecordsRequest) (*pbrc.GetRecordsResponse, error) {
	if t.fail || (t.failSecond && t.count > 0) {
		return &pbrc.GetRecordsResponse{}, fmt.Errorf("recs Built to fail")
	}

	if t.noLocate || (t.noLocateSecond && t.count > 0) {
		return &pbrc.GetRecordsResponse{}, nil
	}

	t.count++
	return &pbrc.GetRecordsResponse{Records: []*pbrc.Record{&pbrc.Record{Release: &pbgd.Release{InstanceId: rec.Filter.Release.InstanceId}, Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_FULL_MATCH}}}}, nil

}

type testOrg struct {
	reorgs    int
	failCount int

	failLocate bool

	emptyLocate bool

	flipLocate  bool
	rflipLocate bool
}

func (t *testOrg) reorgLocation(ctx context.Context, folder int32) error {
	t.reorgs++
	t.failCount--

	if t.failCount <= 0 {
		return fmt.Errorf("Built to fail")
	}
	return nil
}

func (t *testOrg) locate(ctx context.Context, req *pbro.LocateRequest) (*pbro.LocateResponse, error) {
	if t.failLocate {
		return &pbro.LocateResponse{}, fmt.Errorf("Locate Built to fail")
	}

	if t.emptyLocate {
		return &pbro.LocateResponse{FoundLocation: &pbro.Location{Name: "madeup",
			ReleasesLocation: []*pbro.ReleasePlacement{}}}, nil
	}

	if t.flipLocate {
		return &pbro.LocateResponse{FoundLocation: &pbro.Location{Name: "madeup",
			ReleasesLocation: []*pbro.ReleasePlacement{
				&pbro.ReleasePlacement{InstanceId: 1},
				&pbro.ReleasePlacement{InstanceId: 10},
				&pbro.ReleasePlacement{InstanceId: 20},
			}}}, nil

	}
	if t.rflipLocate {
		return &pbro.LocateResponse{FoundLocation: &pbro.Location{Name: "madeup",
			ReleasesLocation: []*pbro.ReleasePlacement{
				&pbro.ReleasePlacement{InstanceId: 10},
				&pbro.ReleasePlacement{InstanceId: 20},
				&pbro.ReleasePlacement{InstanceId: 1},
			}}}, nil

	}

	return &pbro.LocateResponse{FoundLocation: &pbro.Location{Name: "madeup",
		ReleasesLocation: []*pbro.ReleasePlacement{
			&pbro.ReleasePlacement{InstanceId: 10, Slot: 1},
			&pbro.ReleasePlacement{InstanceId: 1, Slot: 2},
			&pbro.ReleasePlacement{InstanceId: 20, Slot: 3},
		}}}, nil
}

func TestAddWithLocateFail(t *testing.T) {
	s := InitTest()
	s.organiser = &testOrg{emptyLocate: true}

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3}})
	if err == nil {
		t.Fatalf("Move did not fail")
	}
}

func TestAddWithLocateFailOne(t *testing.T) {
	s := InitTest()
	s.organiser = &testOrg{failLocate: true}

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3}})
	if err == nil {
		t.Fatalf("Move did not fail")
	}
}

func TestAddDouble(t *testing.T) {
	s := InitTest()

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 2}})
	if err != nil {
		t.Fatalf("Error moving record: %v", err)
	}

	moves, err := s.ListMoves(context.Background(), &pb.ListRequest{})
	if err != nil {
		t.Fatalf("Error listing moves: %v", err)
	}

	if len(moves.GetMoves()) != 0 {
		t.Errorf("Moves have been recorded: %v", moves)
	}
}

func TestRunThrough(t *testing.T) {
	s := InitTest()

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3}})

	if err != nil {
		t.Fatalf("Error moving record: %v", err)
	}

	moves, err := s.ListMoves(context.Background(), &pb.ListRequest{InstanceId: 1})

	if err != nil {
		t.Fatalf("Error listing records: %v", err)
	}

	if len(moves.GetMoves()) != 1 || moves.GetMoves()[0].MoveDate <= 0 || moves.GetMoves()[0].InstanceId != 1 {
		t.Fatalf("Move is a problem: %v", moves)
	}

	_, err = s.ClearMove(context.Background(), &pb.ClearRequest{InstanceId: 123456})
	if err == nil {
		t.Fatalf("No error on clearing fake move")
	}

	_, err = s.ClearMove(context.Background(), &pb.ClearRequest{InstanceId: 1})

	if err != nil {
		t.Fatalf("Error clearing moves: %v", err)
	}

	moves, err = s.ListMoves(context.Background(), &pb.ListRequest{})

	if err != nil {
		t.Fatalf("Error listing records: %v", err)
	}

	if len(moves.GetMoves()) != 0 {
		t.Fatalf("Move is a problem: %v", moves)
	}

}

func TestAppendArchive(t *testing.T) {
	s := InitTest()

	s.addToArchive(context.Background(), &pb.RecordedMove{InstanceId: 1, MoveLocation: "blah", MoveTime: 123})

}

func TestForceUpdate(t *testing.T) {
	s := InitTest()
	_, err := s.ClientUpdate(context.Background(), &pbrc.ClientUpdateRequest{})
	if status.Convert(err).Code() != codes.OK {
		t.Errorf("Bad force: %v", err)
	}
}
func TestForceUpdateFail(t *testing.T) {
	s := InitTest()
	s.getter = &testGetter{failGet: true}
	_, err := s.ClientUpdate(context.Background(), &pbrc.ClientUpdateRequest{})
	if err == nil {
		t.Errorf("Bad force")
	}
}

func TestAddCausesUpdateMissingRecord(t *testing.T) {
	s := InitTest()
	testOrg := &testOrg{failCount: 100}
	s.organiser = testOrg

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{FromFolder: 2, ToFolder: 3}})
	if err == nil {
		t.Fatalf("Move did not fail")
	}
}

func TestMoveWithoutWidthButItsDigital(t *testing.T) {
	s := InitTest()

	// A record in the listening pile that has no width
	s.getter = &testGetter{rec: &pbrc.Record{Release: &pbgd.Release{InstanceId: 123, FolderId: 812802},
		Metadata: &pbrc.ReleaseMetadata{Match: pbrc.ReleaseMetadata_NO_MATCH, GoalFolder: int32(2274270)}}}

	_, err := s.ClientUpdate(context.Background(), &pbrc.ClientUpdateRequest{InstanceId: 123})

	if status.Convert(err).Code() == codes.InvalidArgument {
		t.Errorf("Moving a record without valid width from listening pile did not fail: %v", err)
	}
}
