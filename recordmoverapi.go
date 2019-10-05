package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	pbgd "github.com/brotherlogic/godiscogs"
	pbrc "github.com/brotherlogic/recordcollection/proto"
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
	if in.GetMove().Record != nil {
		return nil, fmt.Errorf("Do not specify the record in the move")
	}
	if in.GetMove().InstanceId == 0 {
		return nil, fmt.Errorf("You need to supply an instance ID")
	}
	location, err := s.organiser.locate(ctx, &pbro.LocateRequest{InstanceId: in.GetMove().InstanceId})
	if err != nil {
		return &pb.MoveResponse{}, err
	}

	for i, r := range location.GetFoundLocation().GetReleasesLocation() {
		if r.GetInstanceId() == in.GetMove().InstanceId {
			in.GetMove().BeforeContext = &pb.Context{}
			in.GetMove().GetBeforeContext().Location = location.GetFoundLocation().Name
			in.GetMove().GetBeforeContext().Slot = location.GetFoundLocation().GetReleasesLocation()[0].Slot

			if i > 0 {
				resp, err := s.recordcollection.getRecords(ctx, &pbrc.GetRecordsRequest{Caller: "recordmover-api1", Filter: &pbrc.Record{Release: &pbgd.Release{InstanceId: location.GetFoundLocation().GetReleasesLocation()[i-1].InstanceId}}})

				if err != nil {
					return &pb.MoveResponse{}, err
				}

				if len(resp.GetRecords()) != 1 {
					return &pb.MoveResponse{}, fmt.Errorf("Ambigous move")
				}

				in.GetMove().GetBeforeContext().BeforeInstance = resp.GetRecords()[0].GetRelease().InstanceId
			}

			if i < len(location.GetFoundLocation().GetReleasesLocation())-1 {
				resp, err := s.recordcollection.getRecords(ctx, &pbrc.GetRecordsRequest{Caller: "recordmover-api2", Filter: &pbrc.Record{Release: &pbgd.Release{InstanceId: location.GetFoundLocation().GetReleasesLocation()[i+1].InstanceId}}})

				if err != nil {
					return &pb.MoveResponse{}, err
				}

				if len(resp.GetRecords()) != 1 {
					return &pb.MoveResponse{}, fmt.Errorf("Ambigous move")
				}

				in.GetMove().GetBeforeContext().AfterInstance = resp.GetRecords()[0].GetRelease().InstanceId
			}

		}
	}

	if in.GetMove().GetBeforeContext() == nil {
		return &pb.MoveResponse{}, fmt.Errorf("Unable to define before context: %v, given %v locations", in.GetMove().InstanceId, len(location.GetFoundLocation().GetReleasesLocation()))
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

	// Overwrite existing move or create a new one
	found := false
	for i, val := range s.config.Moves {
		if val.InstanceId == in.GetMove().InstanceId {
			found = true
			s.config.Moves[i] = in.GetMove()
		}
	}
	if !found {
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
	return &pb.ForceResponse{}, s.moveRecordsHelper(ctx, in.InstanceId)
}
