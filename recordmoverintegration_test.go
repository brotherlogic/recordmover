package main

import (
	"testing"

	pbgd "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordmover/proto"
	"golang.org/x/net/context"
)

func TestUpdatingMove(t *testing.T) {
	s := InitTest()

	_, err := s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 2, ToFolder: 3, Record: &pbrc.Record{Release: &pbgd.Release{InstanceId: 1}}}})

	moves, err := s.ListMoves(context.Background(), &pb.ListRequest{})

	if err != nil {
		t.Fatalf("Error listing records: %v", err)
	}

	if len(moves.GetMoves()) != 1 || moves.GetMoves()[0].MoveDate <= 0 || moves.GetMoves()[0].Record.Release.InstanceId != 1 {
		t.Fatalf("Move is a problem: %v", moves)
	}

	//Move this record to a different folder
	_, err = s.RecordMove(context.Background(), &pb.MoveRequest{Move: &pb.RecordMove{InstanceId: 1, FromFolder: 3, ToFolder: 5, Record: &pbrc.Record{Release: &pbgd.Release{InstanceId: 1}}}})

	moves, err = s.ListMoves(context.Background(), &pb.ListRequest{})

	if err != nil {
		t.Fatalf("Error listing records: %v", err)
	}

	if len(moves.GetMoves()) != 1 || moves.GetMoves()[0].MoveDate <= 0 || moves.GetMoves()[0].Record.Release.InstanceId != 1 {
		t.Fatalf("Move has not been updated problem: %v", moves)
	}

}
