package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	pbgd "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordmover/proto"
	pbro "github.com/brotherlogic/recordsorganiser/proto"
	pbt "github.com/brotherlogic/tracer/proto"
)

// RecordMove moves a record
func (s *Server) RecordMove(ctx context.Context, in *pb.MoveRequest) (*pb.MoveResponse, error) {
	if in.GetMove().Record == nil {
		s.RaiseIssue(ctx, "RecordMove issue", fmt.Sprintf("Move with details %v is missing record", in.GetMove().InstanceId), false)
		return &pb.MoveResponse{}, fmt.Errorf("Missing Record on call")
	}

	location, err := s.organiser.locate(ctx, &pbro.LocateRequest{InstanceId: in.GetMove().Record.GetRelease().InstanceId})
	if err != nil {
		return &pb.MoveResponse{}, err
	}

	for i, r := range location.GetFoundLocation().GetReleasesLocation() {
		if r.GetInstanceId() == in.GetMove().InstanceId {
			if i > 0 {
				in.GetMove().BeforeContext = &pb.Context{}
				in.GetMove().GetBeforeContext().Location = location.GetFoundLocation().Name
				in.GetMove().GetBeforeContext().Slot = location.GetFoundLocation().GetReleasesLocation()[0].Slot
				resp, err := s.recordcollection.getRecords(ctx, &pbrc.GetRecordsRequest{Filter: &pbrc.Record{Release: &pbgd.Release{InstanceId: location.GetFoundLocation().GetReleasesLocation()[i-1].InstanceId}}})

				if err != nil {
					return &pb.MoveResponse{}, err
				}

				if len(resp.GetRecords()) != 1 {
					return &pb.MoveResponse{}, fmt.Errorf("Ambigous move")
				}

				in.GetMove().GetBeforeContext().Before = resp.GetRecords()[0]
			}
		}
	}

	if in.GetMove().GetBeforeContext() == nil {
		return &pb.MoveResponse{}, fmt.Errorf("Unable to define before context: %v", in.GetMove().InstanceId)
	}

	if in.GetMove().ToFolder == in.GetMove().FromFolder {
		return &pb.MoveResponse{}, nil
	}

	err = s.organiser.reorgLocation(ctx, in.Move.ToFolder)
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
