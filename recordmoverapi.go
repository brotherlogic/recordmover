package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	pb "github.com/brotherlogic/recordmover/proto"
	pbro "github.com/brotherlogic/recordsorganiser/proto"
)

func (s *Server) updateArchive(move *pb.RecordedMove) {
	t := time.Now()
	newMove := true
	for _, archMove := range s.config.MoveArchive {
		if archMove.InstanceId == move.InstanceId && archMove.MoveLocation == move.MoveLocation {
			newMove = false
		}
	}

	if newMove {
		s.config.MoveArchive = append(s.config.MoveArchive, move)
	}

	s.lastArch = time.Now().Sub(t)
}

// RecordMove moves a record
func (s *Server) RecordMove(ctx context.Context, in *pb.MoveRequest) (*pb.MoveResponse, error) {
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
	for i, val := range s.config.Moves {
		if val.InstanceId == in.GetMove().InstanceId {
			found = true
			s.config.Moves[i] = in.GetMove()
		}
	}

	if !found {
		in.GetMove().BeforeContext = newBefore
		s.config.Moves = append(s.config.Moves, in.GetMove())
	}

	s.saveMoves(ctx)
	return &pb.MoveResponse{}, nil
}

// ListMoves list the moves made
func (s *Server) ListMoves(ctx context.Context, in *pb.ListRequest) (*pb.ListResponse, error) {
	resp := &pb.ListResponse{Moves: make([]*pb.RecordMove, 0), Archives: s.config.MoveArchive}
	for _, move := range s.config.Moves {
		resp.Moves = append(resp.Moves, move)
	}
	return resp, nil
}

// ClearMove clears a single move
func (s *Server) ClearMove(ctx context.Context, in *pb.ClearRequest) (*pb.ClearResponse, error) {
	for i, mv := range s.config.Moves {
		if mv.InstanceId == in.InstanceId {
			s.config.Moves = append(s.config.Moves[:i], s.config.Moves[i+1:]...)
			s.saveMoves(ctx)
			return &pb.ClearResponse{}, nil
		}
	}

	return nil, fmt.Errorf("Unable to clear move: %v", in.InstanceId)
}

//ForceMove forces a move
func (s *Server) ForceMove(ctx context.Context, in *pb.ForceRequest) (*pb.ForceResponse, error) {
	record, err := s.getter.getRecord(ctx, in.InstanceId)
	if err != nil {
		return nil, err
	}

	return &pb.ForceResponse{}, s.moveRecordInternal(ctx, record)
}
