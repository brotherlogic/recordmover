package main

import (
	"context"
	"testing"

	pb "github.com/brotherlogic/recordmover/proto"
)

func InitTestServer() *Server {
	return Init()
}

func TestAddDouble(t *testing.T) {
	s := InitTestServer()

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
	s := InitTestServer()

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3}})

	if err != nil {
		t.Fatalf("Error moving record: %v", err)
	}

	moves, err := s.ListMoves(context.Background(), &pb.ListRequest{})

	if err != nil {
		t.Fatalf("Error listing records: %v", err)
	}

	if len(moves.GetMoves()) != 1 || moves.GetMoves()[0].MoveDate <= 0 {
		t.Fatalf("Move is a problem: %v", moves)
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
