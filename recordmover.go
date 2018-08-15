package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/goserver/utils"
	"github.com/brotherlogic/keystore/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbgd "github.com/brotherlogic/godiscogs"
	pbg "github.com/brotherlogic/goserver/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordmover/proto"
	pbt "github.com/brotherlogic/tracer/proto"
)

//Server main server type
type Server struct {
	*goserver.GoServer
	getter    getter
	lastProc  time.Time
	lastCount int64
	moves     map[int32]*pb.RecordMove
}

const (
	//KEY is where we store moves
	KEY = "github.com/brotherlogic/recordmover/moves"
)

type prodGetter struct {
	getIP func(string) (string, int32, error)
}

func (s *Server) readMoves(ctx context.Context) error {
	s.moves = make(map[int32]*pb.RecordMove)

	movelist := &pb.Moves{}
	data, _, err := s.KSclient.Read(ctx, KEY, movelist)

	if err != nil {
		return err
	}

	movelist = data.(*pb.Moves)
	for _, m := range movelist.GetMoves() {
		s.moves[m.InstanceId] = m
	}

	return nil
}

func (s *Server) saveMoves(ctx context.Context) {
	s.LogTrace(ctx, "saveMoves", time.Now(), pbt.Milestone_START_FUNCTION)
	moves := &pb.Moves{Moves: make([]*pb.RecordMove, 0)}
	for _, move := range s.moves {
		moves.Moves = append(moves.Moves, move)
	}
	s.KSclient.Save(ctx, KEY, moves)
	s.LogTrace(ctx, "saveMoves", time.Now(), pbt.Milestone_END_FUNCTION)
}

func (p prodGetter) getRecords() ([]*pbrc.Record, error) {
	ip, port, err := p.getIP("recordcollection")
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pbrc.NewRecordCollectionServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	resp, err := client.GetRecords(ctx, &pbrc.GetRecordsRequest{Filter: &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{Dirty: false, MoveFolder: 0}, Release: &pbgd.Release{}}}, grpc.MaxCallRecvMsgSize(1024*1024*1024))
	if err != nil {
		return nil, err
	}
	return resp.GetRecords(), nil
}

func (p prodGetter) update(r *pbrc.Record) error {
	ip, port, err := p.getIP("recordcollection")
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbrc.NewRecordCollectionServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err = client.UpdateRecord(ctx, &pbrc.UpdateRecordRequest{Requestor: "recordmover", Update: r})
	if err != nil {
		return err
	}
	return nil
}

// Init builds the server
func Init() *Server {
	s := &Server{GoServer: &goserver.GoServer{}}
	s.moves = make(map[int32]*pb.RecordMove)
	s.getter = &prodGetter{getIP: utils.Resolve}
	return s
}

// DoRegister does RPC registration
func (s *Server) DoRegister(server *grpc.Server) {
	pb.RegisterMoveServiceServer(server, s)
}

// ReportHealth alerts if we're not healthy
func (s *Server) ReportHealth() bool {
	return true
}

// Mote promotes/demotes this server
func (s *Server) Mote(ctx context.Context, master bool) error {
	if master {
		err := s.readMoves(ctx)
		return err
	}

	return nil
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {
	return []*pbg.State{
		&pbg.State{Key: "last_proc", TimeValue: s.lastProc.Unix(), Value: s.lastCount},
		&pbg.State{Key: "moves", Value: int64(len(s.moves))},
	}
}

func main() {
	var quiet = flag.Bool("quiet", false, "Show all output")
	flag.Parse()

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
	server := Init()
	server.GoServer.KSclient = *keystoreclient.GetClient(server.GetIP)
	server.PrepServer()
	server.Register = server

	server.RegisterServer("recordmover", false)
	server.RegisterRepeatingTask(server.moveRecords, time.Minute)
	fmt.Printf("%v\n", server.Serve())
}
