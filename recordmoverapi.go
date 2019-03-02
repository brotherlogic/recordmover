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
			in.GetMove().BeforeContext = &pb.Context{}
			in.GetMove().GetBeforeContext().Location = location.GetFoundLocation().Name
			in.GetMove().GetBeforeContext().Slot = location.GetFoundLocation().GetReleasesLocation()[0].Slot

			if i > 0 {
				resp, err := s.recordcollection.getRecords(ctx, &pbrc.GetRecordsRequest{Filter: &pbrc.Record{Release: &pbgd.Release{InstanceId: location.GetFoundLocation().GetReleasesLocation()[i-1].InstanceId}}})

				if err != nil {
					return &pb.MoveResponse{}, err
				}

				if len(resp.GetRecords()) != 1 {
					return &pb.MoveResponse{}, fmt.Errorf("Ambigous move")
				}

				in.GetMove().GetBeforeContext().Before = resp.GetRecords()[0]
			}

			if i < len(location.GetFoundLocation().GetReleasesLocation())-1 {
				resp, err := s.recordcollection.getRecords(ctx, &pbrc.GetRecordsRequest{Filter: &pbrc.Record{Release: &pbgd.Release{InstanceId: location.GetFoundLocation().GetReleasesLocation()[i+1].InstanceId}}})

				if err != nil {
					return &pb.MoveResponse{}, err
				}

				if len(resp.GetRecords()) != 1 {
					return &pb.MoveResponse{}, fmt.Errorf("Ambigous move")
				}

				in.GetMove().GetBeforeContext().After = resp.GetRecords()[0]
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
	s.moves[in.GetMove().InstanceId] = in.GetMove()

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

	s.Log(fmt.Sprintf("Moved %v %v -> %v", in.GetMove().InstanceId, in.GetMove().GetBeforeContext().Location, in.GetMove().GetAfterContext().Location))

	s.saveMoves(ctx)
	return &pb.MoveResponse{}, nil
}

// ListMoves list the moves made
func (s *Server) ListMoves(ctx context.Context, in *pb.ListRequest) (*pb.ListResponse, error) {
	resp := &pb.ListResponse{Moves: make([]*pb.RecordMove, 0), Archives: s.config.MoveArchive}
	for _, move := range s.moves {
		resp.Moves = append(resp.Moves, move)
	}
	return resp, nil
}

// ClearMove clears a single move
func (s *Server) ClearMove(ctx context.Context, in *pb.ClearRequest) (*pb.ClearResponse, error) {
	s.Log(fmt.Sprintf("CLEARING %v", in.InstanceId))
	if _, ok := s.moves[in.InstanceId]; !ok {
		return nil, fmt.Errorf("Instance ID not found in moves list")
	}
	delete(s.moves, in.InstanceId)
	for i, mv := range s.config.Moves {
		if mv.InstanceId == in.InstanceId {
			s.config.Moves = append(s.config.Moves[:i], s.config.Moves[i+1:]...)
		}
	}
	s.saveMoves(ctx)
	return &pb.ClearResponse{}, nil
}
