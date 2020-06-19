package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordmover/proto"
	pbro "github.com/brotherlogic/recordsorganiser/proto"
)

// RecordMove moves a record
func (s *Server) RecordMove(ctx context.Context, in *pb.MoveRequest) (*pb.MoveResponse, error) {
	config, err := s.readMoves(ctx)
	if err != nil {
		return nil, err
	}

	if in.GetMove().InstanceId == 0 {
		return nil, fmt.Errorf("You need to supply an instance ID")
	}
	location, err := s.organiser.locate(ctx, &pbro.LocateRequest{InstanceId: in.GetMove().InstanceId})
	if err != nil {
		return &pb.MoveResponse{}, err
	}

	newBefore := &pb.Context{}
	for i, r := range location.GetFoundLocation().GetReleasesLocation() {
		if r.GetInstanceId() == in.GetMove().InstanceId {
			newBefore.Location = location.GetFoundLocation().Name
			newBefore.Slot = location.GetFoundLocation().GetReleasesLocation()[0].Slot

			if i > 0 {
				newBefore.BeforeInstance = location.GetFoundLocation().GetReleasesLocation()[i-1].InstanceId
			}

			if i < len(location.GetFoundLocation().GetReleasesLocation())-1 {
				newBefore.AfterInstance = location.GetFoundLocation().GetReleasesLocation()[i+1].InstanceId
			}
		}
	}

	if newBefore.GetSlot() == 0 {
		return &pb.MoveResponse{}, fmt.Errorf("Unable to define before context: %v, given %v locations", in.GetMove().InstanceId, len(location.GetFoundLocation().GetReleasesLocation()))
	}

	if in.GetMove().ToFolder == in.GetMove().FromFolder {
		return &pb.MoveResponse{}, nil
	}

	in.GetMove().MoveDate = time.Now().Unix()

	// Overwrite existing move or create a new one
	found := false
	for i, val := range config.Moves {
		if val.InstanceId == in.GetMove().InstanceId {
			found = true
			config.Moves[i] = in.GetMove()
		}
	}

	if !found {
		in.GetMove().BeforeContext = newBefore
		config.Moves = append(config.Moves, in.GetMove())
	}

	return &pb.MoveResponse{}, s.saveMoves(ctx, config)
}

// ListMoves list the moves made
func (s *Server) ListMoves(ctx context.Context, in *pb.ListRequest) (*pb.ListResponse, error) {
	config, err := s.readMoves(ctx)
	if err != nil {
		return nil, err
	}
	resp := &pb.ListResponse{Moves: make([]*pb.RecordMove, 0), Archives: make([]*pb.RecordedMove, 0)}
	for _, move := range config.Moves {
		if in.GetInstanceId() == 0 || move.GetInstanceId() == in.GetInstanceId() {
			resp.Moves = append(resp.Moves, move)
		}
	}
	moves, _ := s.readMoveArchive(ctx, in.GetInstanceId())
	for _, move := range moves {
		if move.GetInstanceId() == in.GetInstanceId() {
			resp.Archives = append(resp.Archives, move)
		}
	}
	return resp, nil
}

// ClearMove clears a single move
func (s *Server) ClearMove(ctx context.Context, in *pb.ClearRequest) (*pb.ClearResponse, error) {
	config, err := s.readMoves(ctx)
	if err != nil {
		return nil, err
	}
	for i, mv := range config.Moves {
		if mv.InstanceId == in.InstanceId {
			config.Moves = append(config.Moves[:i], config.Moves[i+1:]...)
			return &pb.ClearResponse{}, s.saveMoves(ctx, config)
		}
	}

	return nil, fmt.Errorf("Unable to clear move: %v", in.InstanceId)
}

//ForceMove forces a move
func (s *Server) ClientUpdate(ctx context.Context, in *pbrc.ClientUpdateRequest) (*pbrc.ClientUpdateResponse, error) {
	record, err := s.getter.getRecord(ctx, in.InstanceId)
	if err != nil {
		return nil, err
	}

	return &pbrc.ClientUpdateResponse{}, s.moveRecordInternal(ctx, record)
}
