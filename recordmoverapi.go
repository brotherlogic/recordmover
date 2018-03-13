package main

import (
	"time"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/recordmover/proto"
)

// RecordMove moves a record
func (s *Server) RecordMove(ctx context.Context, in *pb.MoveRequest) (*pb.MoveResponse, error) {
	if in.GetMove().ToFolder == in.GetMove().FromFolder {
		return &pb.MoveResponse{}, nil
	}
	in.GetMove().MoveDate = time.Now().Unix()
	s.moves[in.GetMove().InstanceId] = in.GetMove()
	s.saveMoves()
	return &pb.MoveResponse{}, nil
}

// ListMoves list the moves made
func (s *Server) ListMoves(ctx context.Context, in *pb.ListRequest) (*pb.ListResponse, error) {
	resp := &pb.ListResponse{Moves: make([]*pb.RecordMove, 0)}
	for _, move := range s.moves {
		resp.Moves = append(resp.Moves, move)
	}
	return resp, nil
}

// ClearMove clears a single move
func (s *Server) ClearMove(ctx context.Context, in *pb.ClearRequest) (*pb.ClearResponse, error) {
	delete(s.moves, in.InstanceId)
	s.saveMoves()
	return &pb.ClearResponse{}, nil
}
