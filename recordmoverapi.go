package main

import (
	"fmt"
	"time"

	pb "github.com/brotherlogic/recordmover/proto"
	pbt "github.com/brotherlogic/tracer/proto"
	"golang.org/x/net/context"
)

// RecordMove moves a record
func (s *Server) RecordMove(ctx context.Context, in *pb.MoveRequest) (*pb.MoveResponse, error) {
	if in.GetMove().Record == nil {
		s.RaiseIssue(ctx, "RecordMove issue", fmt.Sprintf("Move with details %v is missing record", in.GetMove().InstanceId), false)
		return &pb.MoveResponse{}, fmt.Errorf("Missing Record on call")
	}

	if in.GetMove().ToFolder == in.GetMove().FromFolder {
		return &pb.MoveResponse{}, nil
	}

	err := s.organiser.reorgLocation(ctx, in.Move.ToFolder)
	if err != nil {
		return &pb.MoveResponse{}, err
	}
	err = s.organiser.reorgLocation(ctx, in.Move.FromFolder)
	if err != nil {
		return &pb.MoveResponse{}, err
	}

	in.GetMove().MoveDate = time.Now().Unix()
	s.moves[in.GetMove().InstanceId] = in.GetMove()
	s.saveMoves(ctx)

	return &pb.MoveResponse{}, nil
}

// ListMoves list the moves made
func (s *Server) ListMoves(ctx context.Context, in *pb.ListRequest) (*pb.ListResponse, error) {
	s.LogTrace(ctx, "ListMoves", time.Now(), pbt.Milestone_START_FUNCTION)
	resp := &pb.ListResponse{Moves: make([]*pb.RecordMove, 0)}
	for _, move := range s.moves {
		resp.Moves = append(resp.Moves, move)
	}
	s.LogTrace(ctx, "ListMoves", time.Now(), pbt.Milestone_END_FUNCTION)
	return resp, nil
}

// ClearMove clears a single move
func (s *Server) ClearMove(ctx context.Context, in *pb.ClearRequest) (*pb.ClearResponse, error) {
	s.LogTrace(ctx, "ClearMove", time.Now(), pbt.Milestone_START_FUNCTION)
	if _, ok := s.moves[in.InstanceId]; !ok {
		return nil, fmt.Errorf("Instance ID not found in moves list")
	}
	delete(s.moves, in.InstanceId)
	s.saveMoves(ctx)
	s.LogTrace(ctx, "ClearMove", time.Now(), pbt.Milestone_END_FUNCTION)
	return &pb.ClearResponse{}, nil
}
