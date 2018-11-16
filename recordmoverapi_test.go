package main

import (
	"fmt"
	"testing"

	"golang.org/x/net/context"

	pbgd "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordmover/proto"
)

type testOrg struct {
	reorgs    int
	failCount int
}

func (t *testOrg) reorgLocation(ctx context.Context, folder int32) error {
	t.reorgs++
	t.failCount--

	if t.failCount <= 0 {
		return fmt.Errorf("Built to fail")
	}
	return nil
}

func TestAddCausesUpdate(t *testing.T) {
	s := InitTest()
	testOrg := &testOrg{failCount: 100}
	s.organiser = testOrg

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3, Record: &pbrc.Record{}}})
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

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3, Record: &pbrc.Record{}}})
	if err == nil {
		t.Fatalf("No error")
	}
}

func TestAddCausesUpdateFail2(t *testing.T) {
	s := InitTest()
	testOrg := &testOrg{failCount: 2}
	s.organiser = testOrg

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3, Record: &pbrc.Record{}}})
	if err == nil {
		t.Fatalf("No error")
	}
}

func TestAddDouble(t *testing.T) {
	s := InitTest()

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 2, Record: &pbrc.Record{}}})
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

	if len(moves.GetMoves()) != 0 {
		t.Fatalf("Move is a problem: %v", moves)
	}

}
