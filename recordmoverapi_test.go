package main

import (
	"fmt"
	"testing"

	pbgd "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordmover/proto"
	pbro "github.com/brotherlogic/recordsorganiser/proto"
	"golang.org/x/net/context"
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
		return &pbrc.GetRecordsResponse{}, fmt.Errorf("Recs Built to fail")
	}

	if t.noLocate || (t.noLocateSecond && t.count > 0) {
		return &pbrc.GetRecordsResponse{}, nil
	}

	t.count++
	return &pbrc.GetRecordsResponse{Records: []*pbrc.Record{&pbrc.Record{Release: &pbgd.Release{InstanceId: rec.Filter.Release.InstanceId}}}}, nil

}

type testOrg struct {
	reorgs    int
	failCount int

	failLocate bool

	emptyLocate bool
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

	return &pbro.LocateResponse{FoundLocation: &pbro.Location{Name: "madeup",
		ReleasesLocation: []*pbro.ReleasePlacement{
			&pbro.ReleasePlacement{InstanceId: 10},
			&pbro.ReleasePlacement{InstanceId: 1},
			&pbro.ReleasePlacement{InstanceId: 20},
		}}}, nil
}

func TestAddWithRecordPullFail(t *testing.T) {
	s := InitTest()
	s.recordcollection = &testCol{fail: true}

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3, Record: &pbrc.Record{Release: &pbgd.Release{InstanceId: 1}}}})
	if err == nil {
		t.Fatalf("Move did not fail")
	}
}

func TestAddWithRecordPullFailOnSecond(t *testing.T) {
	s := InitTest()
	s.recordcollection = &testCol{failSecond: true}

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3, Record: &pbrc.Record{Release: &pbgd.Release{InstanceId: 1}}}})
	if err == nil {
		t.Fatalf("Move did not fail")
	}
}

func TestAddWithLocateFail(t *testing.T) {
	s := InitTest()
	s.organiser = &testOrg{emptyLocate: true}

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3, Record: &pbrc.Record{Release: &pbgd.Release{InstanceId: 1}}}})
	if err == nil {
		t.Fatalf("Move did not fail")
	}
}

func TestAddWithLocateFailOne(t *testing.T) {
	s := InitTest()
	s.organiser = &testOrg{failLocate: true}

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3, Record: &pbrc.Record{Release: &pbgd.Release{InstanceId: 1}}}})
	if err == nil {
		t.Fatalf("Move did not fail")
	}
}

func TestAddWithLocateEmpty(t *testing.T) {
	s := InitTest()
	s.recordcollection = &testCol{noLocate: true}

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3, Record: &pbrc.Record{Release: &pbgd.Release{InstanceId: 1}}}})
	if err == nil {
		t.Fatalf("Move did not fail")
	}
}

func TestAddWithLocateEmptySecond(t *testing.T) {
	s := InitTest()
	s.recordcollection = &testCol{noLocateSecond: true}

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3, Record: &pbrc.Record{Release: &pbgd.Release{InstanceId: 1}}}})
	if err == nil {
		t.Fatalf("Move did not fail")
	}
}

func TestAddCausesUpdate(t *testing.T) {
	s := InitTest()
	testOrg := &testOrg{failCount: 100}
	s.organiser = testOrg

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3, Record: &pbrc.Record{Release: &pbgd.Release{InstanceId: 1}}}})
	if err != nil {
		t.Fatalf("Error making move: %v", err)
	}

	if testOrg.reorgs != 2 {
		t.Errorf("Moves have not caused reorgs")
	}
}

func TestAddCausesUpdateMissingRecord(t *testing.T) {
	s := InitTest()
	testOrg := &testOrg{failCount: 100}
	s.organiser = testOrg

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3}})
	if err == nil {
		t.Fatalf("Move did not fail")
	}
}

func TestAddCausesUpdateFail1(t *testing.T) {
	s := InitTest()
	testOrg := &testOrg{failCount: 1}
	s.organiser = testOrg

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3, Record: &pbrc.Record{Release: &pbgd.Release{InstanceId: 1}}}})
	if err == nil {
		t.Fatalf("No error")
	}
}

func TestAddCausesUpdateFail2(t *testing.T) {
	s := InitTest()
	testOrg := &testOrg{failCount: 2}
	s.organiser = testOrg

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3, Record: &pbrc.Record{Release: &pbgd.Release{InstanceId: 1}}}})
	if err == nil {
		t.Fatalf("No error")
	}
}

func TestAddDouble(t *testing.T) {
	s := InitTest()

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 2, Record: &pbrc.Record{Release: &pbgd.Release{InstanceId: 1}}}})
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

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3, Record: &pbrc.Record{Release: &pbgd.Release{InstanceId: 1}}}})

	if err != nil {
		t.Fatalf("Error moving record: %v", err)
	}

	moves, err := s.ListMoves(context.Background(), &pb.ListRequest{})

	if err != nil {
		t.Fatalf("Error listing records: %v", err)
	}

	if len(moves.GetMoves()) != 1 || moves.GetMoves()[0].MoveDate <= 0 || moves.GetMoves()[0].Record.Release.InstanceId != 1 {
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

	if len(moves.GetMoves()) != 0 || len(s.config.Moves) != 0 {
		t.Fatalf("Move is a problem: %v", moves)
	}

}

func TestAppendArchive(t *testing.T) {
	s := InitTest()
	s.config.MoveArchive = append(s.config.MoveArchive, &pb.RecordedMove{InstanceId: 1, MoveLocation: "blah", MoveTime: 12})

	s.updateArchive(&pb.RecordedMove{InstanceId: 1, MoveLocation: "blah", MoveTime: 123})

	if s.config.MoveArchive[0].MoveTime != 12 {
		t.Errorf("Update has failed")
	}
}
