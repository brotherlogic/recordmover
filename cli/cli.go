package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pbgd "github.com/brotherlogic/godiscogs"
	pbgs "github.com/brotherlogic/goserver/proto"
	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordmover/proto"
	pbro "github.com/brotherlogic/recordsorganiser/proto"
	pbt "github.com/brotherlogic/tracer/proto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
)

func getRecord(ctx context.Context, instanceID int32) string {
	utils.SendTrace(ctx, "getRecord", time.Now(), pbt.Milestone_START_FUNCTION, "recordmover-cli")
	host, port, err := utils.Resolve("recordcollection")
	if err != nil {
		log.Fatalf("Unable to reach recordcollection: %v", err)
	}
	conn, err := grpc.Dial(host+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}

	client := pbrc.NewRecordCollectionServiceClient(conn)
	r, err := client.GetRecords(ctx, &pbrc.GetRecordsRequest{Filter: &pbrc.Record{Release: &pbgd.Release{InstanceId: instanceID}}})
	if err != nil {
		log.Fatalf("Unable to get records: %v", err)
	}

	utils.SendTrace(ctx, "getRecord", time.Now(), pbt.Milestone_END_FUNCTION, "recordmover-cli")
	return r.GetRecords()[0].GetRelease().Title
}

func getFolder(ctx context.Context, folderID int32) string {
	utils.SendTrace(ctx, "getFolder", time.Now(), pbt.Milestone_START_FUNCTION, "recordmover-cli")
	host, port, err := utils.Resolve("recordsorganiser")
	if err != nil {
		log.Fatalf("Unable to reach organiser: %v", err)
	}
	conn, err := grpc.Dial(host+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}

	client := pbro.NewOrganiserServiceClient(conn)
	r, err := client.GetQuota(ctx, &pbro.QuotaRequest{FolderId: folderID})
	if err != nil {
		log.Fatalf("Unable to get quota: %v", err)
	}

	utils.SendTrace(ctx, "getFolder", time.Now(), pbt.Milestone_END_FUNCTION, "recordmover-cli")
	return r.LocationName
}

func main() {
	host, port, err := utils.Resolve("recordmover")
	if err != nil {
		log.Fatalf("Unable to reach organiser: %v", err)
	}
	conn, err := grpc.Dial(host+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}

	client := pb.NewMoveServiceClient(conn)
	ctx, cancel := utils.BuildContext("recordmover-cli", pbgs.ContextType_LONG)
	defer cancel()

	switch os.Args[1] {
	case "get":
		res, err := client.ListMoves(ctx, &pb.ListRequest{})
		if err != nil {
			log.Fatalf("Error on GET: %v", err)
		}
		for _, move := range res.GetMoves() {
			fmt.Printf("%v: %v -> %v\n", getRecord(ctx, move.InstanceId), getFolder(ctx, move.FromFolder), getFolder(ctx, move.ToFolder))
		}
	case "getclear":
		res, err := client.ListMoves(ctx, &pb.ListRequest{})
		if err != nil {
			log.Fatalf("Error on GET: %v", err)
		}
		for _, move := range res.GetMoves() {
			fmt.Printf("%v: %v -> %v\n", getRecord(ctx, move.InstanceId), getFolder(ctx, move.FromFolder), getFolder(ctx, move.ToFolder))
			client.ClearMove(ctx, &pb.ClearRequest{InstanceId: move.InstanceId})
		}
	}

}
