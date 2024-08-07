package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pbcdp "github.com/brotherlogic/cdprocessor/proto"
	pbg "github.com/brotherlogic/goserver/proto"
	pbgr "github.com/brotherlogic/gramophile/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	rcpb "github.com/brotherlogic/recordcollection/proto"
	rmpb "github.com/brotherlogic/recordmatcher/proto"
	pb "github.com/brotherlogic/recordmover/proto"
	pbro "github.com/brotherlogic/recordsorganiser/proto"
)

// Server main server type
type Server struct {
	*goserver.GoServer
	getter      getter
	lastProc    time.Time
	lastCount   int64
	cdproc      cdproc
	organiser   organiser
	lastArch    time.Duration
	lastID      int32
	lastIDCount int
	total       int
	count       int
	testing     bool
	configMutex *sync.Mutex
	block       *sync.Mutex
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
		0,
		int32(0),
		0,
		0,
		0,
		false,
		&sync.Mutex{},
		&sync.Mutex{},
	}
	s.getter = &prodGetter{s.FDialServer}
	s.cdproc = &cdprocProd{s.FDialServer}
	s.organiser = &prodOrganiser{s.FDialServer}
	return s
}

const (
	//ConfigKey is where we store the overall config
	ConfigKey = "github.com/brotherlogic/recordmover/config"

	//MoveKey is the base key for storing moves
	MoveKey = "github.com/brotherlogic/recordmove/archive"
)

type organiser interface {
	reorgLocation(ctx context.Context, folder int32) error
	locate(ctx context.Context, req *pbro.LocateRequest) (*pbro.LocateResponse, error)
}

type prodOrganiser struct {
	dial func(ctx context.Context, server string) (*grpc.ClientConn, error)
}

func (p *prodOrganiser) reorgLocation(ctx context.Context, folder int32) error {
	conn, err := p.dial(ctx, "recordsorganiser")
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pbro.NewOrganiserServiceClient(conn)
	_, err = client.GetOrganisation(ctx, &pbro.GetOrganisationRequest{ForceReorg: true, Locations: []*pbro.Location{&pbro.Location{FolderIds: []int32{folder}}}})
	return err
}

func (p *prodOrganiser) locate(ctx context.Context, req *pbro.LocateRequest) (*pbro.LocateResponse, error) {
	conn, err := p.dial(ctx, "recordsorganiser")
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
	dial func(ctx context.Context, server string) (*grpc.ClientConn, error)
}

func (p *cdprocProd) isRipped(ctx context.Context, ID int32) bool {
	conn, err := p.dial(ctx, "cdprocessor")
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
	dial func(ctx context.Context, server string) (*grpc.ClientConn, error)
}

func (p *prodGetter) getRecordsSince(ctx context.Context, since int64) ([]int32, error) {
	conn, err := p.dial(ctx, "recordcollection")
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
	conn, err := p.dial(ctx, "recordcollection")
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

func (s *Server) forceMatch(ctx context.Context, ID int32) {
	if s.testing {
		return
	}
	conn, err := s.FDialServer(ctx, "recordmatcher")
	if err != nil {
		return
	}
	defer conn.Close()

	client := rmpb.NewRecordMatcherServiceClient(conn)
	client.Match(ctx, &rmpb.MatchRequest{InstanceId: ID})
}

func (s *Server) readMoves(ctx context.Context) (*pb.Config, error) {
	config := &pb.Config{}
	data, _, err := s.KSclient.Read(ctx, ConfigKey, config)

	if err != nil {
		return nil, err
	}

	return data.(*pb.Config), nil
}

func (s *Server) readMoveArchive(ctx context.Context, iid int32) ([]*pb.RecordedMove, error) {
	if s.testing {
		return nil, fmt.Errorf("Bad")
	}
	config := &pb.MoveArchive{}
	data, _, err := s.KSclient.Read(ctx, fmt.Sprintf("%v-%v", MoveKey, iid), config)

	if err != nil {
		return nil, err
	}

	config = data.(*pb.MoveArchive)
	return config.GetMoves(), nil
}

func (s *Server) saveMoves(ctx context.Context, config *pb.Config) error {
	return s.KSclient.Save(ctx, ConfigKey, config)
}

func (s *Server) saveMoveArchive(ctx context.Context, iid int32, moves []*pb.RecordedMove) error {
	return s.KSclient.Save(ctx, fmt.Sprintf("%v-%v", MoveKey, iid), &pb.MoveArchive{Moves: moves})
}

func buildContext(ctx context.Context) (context.Context, context.CancelFunc, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, err
	}

	text, err := ioutil.ReadFile(fmt.Sprintf("%v/.gramophile", dirname))
	if err != nil {
		return nil, nil, err
	}

	user := &pbgr.GramophileAuth{}
	err = proto.UnmarshalText(string(text), user)
	if err != nil {
		return nil, nil, err
	}

	mContext := metadata.AppendToOutgoingContext(ctx, "auth-token", user.GetToken())
	ctx, cancel := context.WithTimeout(mContext, time.Minute)
	return ctx, cancel, nil
}

func (p prodGetter) update(ctx context.Context, instanceID int32, reason string, folder int32) error {

	// Dial gram
	conn, err := grpc.NewClient("gramophile-grpc.brotherlogic-backend.com:80", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()

	gclient := pbgr.NewGramophileEServiceClient(conn)
	nctx, cancel, gerr := buildContext(ctx)
	if gerr != nil {
		return gerr
	}
	defer cancel()

	s.CtxLog(ctx, fmt.Sprintf("Moving %v to %v", instanceID, folder))
	_, err = gclient.SetIntent(nctx, &pbgr.SetIntentRequest{
		InstanceId: int64(instanceID),
		Intent: &pbgr.Intent{
			NewFolder: folder,
		},
	})

	return err
}

// DoRegister does RPC registration
func (s *Server) DoRegister(server *grpc.Server) {
	pb.RegisterMoveServiceServer(server, s)
	rcpb.RegisterClientUpdateServiceServer(server, s)
}

// ReportHealth alerts if we're not healthy
func (s *Server) ReportHealth() bool {
	return true
}

// Shutdown the server
func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

// Mote promotes/demotes this server
func (s *Server) Mote(ctx context.Context, master bool) error {
	return nil
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {
	return []*pbg.State{}
}

func main() {
	server := Init()
	server.PrepServer("recordmover")
	server.Register = server

	err := server.RegisterServerV2(false)
	if err != nil {
		return
	}

	fmt.Printf("%v\n", server.Serve())
}
