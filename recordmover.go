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

	pbcdp "github.com/brotherlogic/cdprocessor/proto"
	pbgd "github.com/brotherlogic/godiscogs"
	pbg "github.com/brotherlogic/goserver/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordmover/proto"
	pbro "github.com/brotherlogic/recordsorganiser/proto"
	pbt "github.com/brotherlogic/tracer/proto"
)

//Server main server type
type Server struct {
	*goserver.GoServer
	getter           getter
	lastProc         time.Time
	lastCount        int64
	moves            map[int32]*pb.RecordMove
	cdproc           cdproc
	organiser        organiser
	recordcollection recordcollection
	config           *pb.Config
	lastArch         time.Duration
}

const (
	//KEY is where we store moves
	KEY = "github.com/brotherlogic/recordmover/moves"

	//ConfigKey is where we store the overall config
	ConfigKey = "github.com/brotherlogic/recordmover/config"
)

type recordcollection interface {
	getRecords(ctx context.Context, rec *pbrc.GetRecordsRequest) (*pbrc.GetRecordsResponse, error)
}

type prodRecordcollection struct{}

func (p *prodRecordcollection) getRecords(ctx context.Context, req *pbrc.GetRecordsRequest) (*pbrc.GetRecordsResponse, error) {
	ip, port, err := utils.Resolve("recordcollection")
	if err != nil {
		return &pbrc.GetRecordsResponse{}, err
	}

	conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		return &pbrc.GetRecordsResponse{}, err
	}
	defer conn.Close()

	client := pbrc.NewRecordCollectionServiceClient(conn)
	return client.GetRecords(ctx, req)
}

type organiser interface {
	reorgLocation(ctx context.Context, folder int32) error
	locate(ctx context.Context, req *pbro.LocateRequest) (*pbro.LocateResponse, error)
}

type prodOrganiser struct{}

func (p *prodOrganiser) reorgLocation(ctx context.Context, folder int32) error {
	ip, port, err := utils.Resolve("recordsorganiser")
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbro.NewOrganiserServiceClient(conn)
	_, err = client.GetOrganisation(ctx, &pbro.GetOrganisationRequest{ForceReorg: true, Locations: []*pbro.Location{&pbro.Location{FolderIds: []int32{folder}}}})
	return err
}

func (p *prodOrganiser) locate(ctx context.Context, req *pbro.LocateRequest) (*pbro.LocateResponse, error) {
	ip, port, err := utils.Resolve("recordsorganiser")
	if err != nil {
		return &pbro.LocateResponse{}, err
	}

	conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		return &pbro.LocateResponse{}, err
	}
	defer conn.Close()

	client := pbro.NewOrganiserServiceClient(conn)
	return client.Locate(ctx, req)
}

type cdproc interface {
	isRipped(ctx context.Context, ID int32) bool
}

type cdprocProd struct{}

func (p *cdprocProd) isRipped(ctx context.Context, ID int32) bool {
	ip, port, err := utils.Resolve("cdprocessor")
	if err != nil {
		return false
	}

	conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		return false
	}
	defer conn.Close()

	client := pbcdp.NewCDProcessorClient(conn)
	res, err := client.GetRipped(ctx, &pbcdp.GetRippedRequest{})
	if err != nil {
		return false
	}

	for _, r := range res.GetRipped() {
		if r.Id == ID {
			return true
		}
	}

	return false
}

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

		if m.InstanceId == 127363223 {
			m.LastUpdate = 0
		}
	}

	//Side load the config
	config := &pb.Config{}
	data, _, err = s.KSclient.Read(ctx, ConfigKey, config)

	if err != nil {
		return err
	}

	s.config = data.(*pb.Config)

	return nil
}

func (s *Server) saveMoves(ctx context.Context) {
	s.LogTrace(ctx, "saveMoves", time.Now(), pbt.Milestone_START_FUNCTION)
	moves := &pb.Moves{Moves: make([]*pb.RecordMove, 0)}
	for _, move := range s.moves {
		moves.Moves = append(moves.Moves, move)
	}
	s.KSclient.Save(ctx, KEY, moves)
	s.KSclient.Save(ctx, ConfigKey, s.config)
	s.LogTrace(ctx, "saveMoves", time.Now(), pbt.Milestone_END_FUNCTION)
}

func (p prodGetter) getRecords(ctx context.Context) ([]*pbrc.Record, error) {
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
	resp, err := client.GetRecords(ctx, &pbrc.GetRecordsRequest{MoveStrip: true, Filter: &pbrc.Record{Metadata: &pbrc.ReleaseMetadata{Dirty: false, MoveFolder: 0}, Release: &pbgd.Release{}}}, grpc.MaxCallRecvMsgSize(1024*1024*1024))
	if err != nil {
		return nil, err
	}
	return resp.GetRecords(), nil
}

func (p prodGetter) update(ctx context.Context, r *pbrc.Record) error {
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
	_, err = client.UpdateRecord(ctx, &pbrc.UpdateRecordRequest{Requestor: "recordmover", Update: r})
	if err != nil {
		return err
	}
	return nil
}

// Init builds the server
func Init() *Server {
	s := &Server{
		&goserver.GoServer{},
		&prodGetter{getIP: utils.Resolve},
		time.Unix(0, 1),
		0,
		make(map[int32]*pb.RecordMove),
		&cdprocProd{},
		&prodOrganiser{},
		&prodRecordcollection{},
		&pb.Config{},
		0,
	}
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
		return s.readMoves(ctx)
	}

	return nil
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {
	fromCount := int64(0)
	toCount := int64(0)
	oldest := time.Now().Unix()
	for _, m := range s.moves {
		if m.BeforeContext != nil {
			fromCount++
		}
		if m.AfterContext != nil {
			toCount++
		}

		if m.MoveDate < oldest {
			oldest = m.MoveDate
		}
	}

	return []*pbg.State{
		&pbg.State{Key: "last_proc", TimeValue: s.lastProc.Unix()},
		&pbg.State{Key: "moves", Value: int64(len(s.moves))},
		&pbg.State{Key: "moves_with_from", Value: fromCount},
		&pbg.State{Key: "moves_with_to", Value: toCount},
		&pbg.State{Key: "last_count", Value: s.lastCount},
		&pbg.State{Key: "config_moves", Value: int64(len(s.config.Moves))},
		&pbg.State{Key: "config_archives", Value: int64(len(s.config.MoveArchive))},
		&pbg.State{Key: "archive_process", TimeDuration: s.lastArch.Nanoseconds()},
		&pbg.State{Key: "oldest_move", TimeValue: oldest},
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
	server.RPCTracing = true

	server.RegisterServer("recordmover", false)
	server.RegisterRepeatingTask(server.moveRecords, "move_records", time.Minute)
	server.RegisterRepeatingTask(server.refreshMoves, "refresh_moves", time.Minute)
	server.RegisterRepeatingTask(server.lookForStale, "look_for_stale", time.Minute)

	fmt.Printf("%v\n", server.Serve())
}
