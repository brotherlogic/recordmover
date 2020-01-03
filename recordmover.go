package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/keystore/client"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbcdp "github.com/brotherlogic/cdprocessor/proto"
	pbg "github.com/brotherlogic/goserver/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordmover/proto"
	pbro "github.com/brotherlogic/recordsorganiser/proto"
)

//Server main server type
type Server struct {
	*goserver.GoServer
	getter      getter
	lastProc    time.Time
	lastCount   int64
	cdproc      cdproc
	organiser   organiser
	config      *pb.Config
	lastArch    time.Duration
	lastID      int32
	lastIDCount int
	total       int
	count       int
}

// Init builds the server
func Init() *Server {
	s := &Server{
		&goserver.GoServer{},
		&prodGetter{},
		time.Unix(0, 1),
		0,
		&cdprocProd{},
		&prodOrganiser{},
		&pb.Config{},
		0,
		int32(0),
		0,
		0,
		0,
	}
	s.getter = &prodGetter{s.DialMaster}
	s.cdproc = &cdprocProd{s.DialMaster}
	s.organiser = &prodOrganiser{s.DialMaster}
	return s
}

const (
	//ConfigKey is where we store the overall config
	ConfigKey = "github.com/brotherlogic/recordmover/config"
)

type organiser interface {
	reorgLocation(ctx context.Context, folder int32) error
	locate(ctx context.Context, req *pbro.LocateRequest) (*pbro.LocateResponse, error)
}

type prodOrganiser struct {
	dial func(server string) (*grpc.ClientConn, error)
}

func (p *prodOrganiser) reorgLocation(ctx context.Context, folder int32) error {
	conn, err := p.dial("recordsorganiser")
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbro.NewOrganiserServiceClient(conn)
	_, err = client.GetOrganisation(ctx, &pbro.GetOrganisationRequest{ForceReorg: true, Locations: []*pbro.Location{&pbro.Location{FolderIds: []int32{folder}}}})
	return err
}

func (p *prodOrganiser) locate(ctx context.Context, req *pbro.LocateRequest) (*pbro.LocateResponse, error) {
	conn, err := p.dial("recordsorganiser")
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

type cdprocProd struct {
	dial func(server string) (*grpc.ClientConn, error)
}

func (p *cdprocProd) isRipped(ctx context.Context, ID int32) bool {
	conn, err := p.dial("cdprocessor")
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
	dial func(server string) (*grpc.ClientConn, error)
}

func (p *prodGetter) getRecordsSince(ctx context.Context, since int64) ([]int32, error) {
	conn, err := p.dial("recordcollection")
	if err != nil {
		return []int32{}, err
	}
	defer conn.Close()

	client := pbrc.NewRecordCollectionServiceClient(conn)
	resp, err := client.QueryRecords(ctx, &pbrc.QueryRecordsRequest{Query: &pbrc.QueryRecordsRequest_UpdateTime{since}})

	if err != nil {
		return []int32{}, err
	}

	return resp.GetInstanceIds(), err
}
func (p *prodGetter) getRecord(ctx context.Context, instanceID int32) (*pbrc.Record, error) {
	conn, err := p.dial("recordcollection")
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	client := pbrc.NewRecordCollectionServiceClient(conn)
	resp, err := client.GetRecord(ctx, &pbrc.GetRecordRequest{InstanceId: instanceID})

	if err != nil {
		return nil, err
	}

	return resp.GetRecord(), err
}

func (s *Server) readMoves(ctx context.Context) error {
	config := &pb.Config{}
	data, _, err := s.KSclient.Read(ctx, ConfigKey, config)

	if err != nil {
		return err
	}

	s.config = data.(*pb.Config)

	return nil
}

func (s *Server) saveMoves(ctx context.Context) {
	s.KSclient.Save(ctx, ConfigKey, s.config)
}

func (p prodGetter) update(ctx context.Context, r *pbrc.Record) error {
	conn, err := p.dial("recordcollection")
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

// DoRegister does RPC registration
func (s *Server) DoRegister(server *grpc.Server) {
	pb.RegisterMoveServiceServer(server, s)
}

// ReportHealth alerts if we're not healthy
func (s *Server) ReportHealth() bool {
	return true
}

// Shutdown the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.saveMoves(ctx)
	return nil
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
	for _, m := range s.config.Moves {
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
		&pbg.State{Key: "last_pull", TimeValue: s.config.LastPull},
		&pbg.State{Key: "config_size", Value: int64(proto.Size(s.config))},
		&pbg.State{Key: "progress", Text: fmt.Sprintf("%v / %v", s.count, s.total)},
		&pbg.State{Key: "last_id_count", Value: int64(s.lastIDCount)},
		&pbg.State{Key: "last_id", Value: int64(s.lastID)},
		&pbg.State{Key: "last_proc", TimeValue: s.lastProc.Unix()},
		&pbg.State{Key: "moves_with_from", Value: fromCount},
		&pbg.State{Key: "moves_with_to", Value: toCount},
		&pbg.State{Key: "last_count", Value: s.lastCount},
		&pbg.State{Key: "config_moves", Value: int64(len(s.config.Moves))},
		&pbg.State{Key: "config_archives", Value: int64(len(s.config.MoveArchive))},
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
	server.GoServer.KSclient = *keystoreclient.GetClient(server.DialMaster)
	server.PrepServer()
	server.Register = server

	err := server.RegisterServerV2("recordmover", false, false)
	if err != nil {
		return
	}

	server.RegisterRepeatingTask(server.moveRecords, "move_records", time.Minute*5)
	server.RegisterRepeatingTask(server.refreshMoves, "refresh_moves", time.Minute)

	fmt.Printf("%v\n", server.Serve())
}
