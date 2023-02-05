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

	pbrc "github.com/brotherlogic/recordcollection/proto"
	pb "github.com/brotherlogic/recordmover/proto"
	pbro "github.com/brotherlogic/recordsorganiser/proto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
)

func getRecord(ctx context.Context, instanceID int32) *pbrc.Record {
	host, port, err := utils.Resolve("recordcollection", "recordmovercli-getRecord")
	if err != nil {
		log.Fatalf("Unable to reach recordcollection: %v", err)
	}
	conn, err := grpc.Dial(host+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}

	client := pbrc.NewRecordCollectionServiceClient(conn)
	r, err := client.GetRecord(ctx, &pbrc.GetRecordRequest{InstanceId: instanceID})
	if err != nil {
		log.Fatalf("Unable to get records: %v", err)
	}

	return r.GetRecord()
}

func getFolder(ctx context.Context, folderID int32) (string, error) {
	host, port, err := utils.Resolve("recordsorganiser", "recordmovercli-getFolder")
	if err != nil {
		return "", err
	}
	conn, err := grpc.Dial(host+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		return "", err
	}

	client := pbro.NewOrganiserServiceClient(conn)
	r, err := client.GetQuota(ctx, &pbro.QuotaRequest{FolderId: folderID})
	if err != nil {
		return "", err
	}

	return r.LocationName, nil
}

func getReleaseString(ctx context.Context, instanceID int32) string {
	host, port, err := utils.Resolve("recordcollection", "recordmovercli-getReleaseString")
	if err != nil {
		log.Fatalf("Unable to reach collection: %v", err)
	}
	conn, err := grpc.Dial(host+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}

	client := pbrc.NewRecordCollectionServiceClient(conn)
	rel, err := client.GetRecord(ctx, &pbrc.GetRecordRequest{InstanceId: instanceID})
	if err != nil {
		log.Fatalf("unable to get record: %v", err)
	}
	return rel.GetRecord().GetRelease().Title + " [" + strconv.Itoa(int(instanceID)) + "]"
}

func getLocation(ctx context.Context, rec *pbrc.Record) (string, error) {
	host, port, err := utils.Resolve("recordsorganiser", "recordmovercli-getLocation")
	if err != nil {
		return "", err
	}
	conn, err := grpc.Dial(host+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	defer conn.Close()

	if err != nil {
		return "", err
	}

	client := pbro.NewOrganiserServiceClient(conn)
	location, err := client.Locate(ctx, &pbro.LocateRequest{InstanceId: rec.GetRelease().InstanceId})
	str := ""
	if err != nil {
		str += fmt.Sprintf("Unable to locate instance (%v) because %v\n", rec.GetRelease().InstanceId, err)
	} else {
		for i, r := range location.GetFoundLocation().GetReleasesLocation() {
			if r.GetInstanceId() == rec.GetRelease().InstanceId {
				str += fmt.Sprintf("  Slot %v\n", r.GetSlot())
				if i > 0 {
					str += fmt.Sprintf("  %v. %v\n", i-1, getReleaseString(ctx, location.GetFoundLocation().GetReleasesLocation()[i-1].InstanceId))
				}
				str += fmt.Sprintf("  %v. %v\n", i, getReleaseString(ctx, location.GetFoundLocation().GetReleasesLocation()[i].InstanceId))
				if i < len(location.GetFoundLocation().GetReleasesLocation())-1 {
					str += fmt.Sprintf("  %v. %v\n", i+1, getReleaseString(ctx, location.GetFoundLocation().GetReleasesLocation()[i+1].InstanceId))
				}
			}
		}
	}

	return str, nil
}

func main() {
	ctx, cancel := utils.BuildContext("recordmover-cli-"+os.Args[1], "recordmover-cli")
	defer cancel()

	conn, err := utils.LFDialServer(ctx, "recordmover")

	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewMoveServiceClient(conn)

	switch os.Args[1] {
	case "fullping":
		ctx2, cancel2 := utils.ManualContext("recordcollectioncli-"+os.Args[1], time.Hour)
		defer cancel2()

		conn2, err := utils.LFDialServer(ctx2, "recordcollection")
		if err != nil {
			log.Fatalf("Cannot reach rc: %v", err)
		}
		defer conn2.Close()

		registry := pbrc.NewRecordCollectionServiceClient(conn2)
		ids, err := registry.QueryRecords(ctx2, &pbrc.QueryRecordsRequest{Query: &pbrc.QueryRecordsRequest_FolderId{673768}})
		if err != nil {
			log.Fatalf("Bad query: %v", err)
		}
		client2 := pbrc.NewClientUpdateServiceClient(conn)
		for _, id := range ids.GetInstanceIds() {
			ctx3, cancel3 := utils.ManualContext("fullping", time.Minute)
			res, err := client2.ClientUpdate(ctx3, &pbrc.ClientUpdateRequest{InstanceId: int32(id)})
			fmt.Printf("%v->%v\n", res, err)
			cancel3()
		}
	case "ping":
		id, err := strconv.Atoi(os.Args[2])
		sclient := pbrc.NewClientUpdateServiceClient(conn)
		_, err = sclient.ClientUpdate(ctx, &pbrc.ClientUpdateRequest{InstanceId: int32(id)})
		if err != nil {
			log.Fatalf("Error on GET: %v", err)
		}
	case "get":
		v, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("%v", err)
		}
		res, err := client.ListMoves(ctx, &pb.ListRequest{InstanceId: int32(v)})
		if err != nil {
			log.Fatalf("Error on GET: %v", err)
		}
		for _, move := range res.GetMoves() {
			if len(os.Args) == 2 || strconv.Itoa(int(move.InstanceId)) == os.Args[2] {
				fmt.Printf("Move %v -> %v\n", move.InstanceId, move)
				fmt.Printf("BEFORE %v %v %v\n", move.GetBeforeContext().GetLocation(), move.GetBeforeContext().GetBefore() == nil, move.GetBeforeContext().GetAfter() == nil)
				if move.AfterContext != nil {
					fmt.Printf("AFTER %v %v %v\n", move.AfterContext.Location, move.AfterContext.Before == nil, move.AfterContext.After == nil)
				}
				move.Record = &pbrc.Record{}
				fmt.Printf("RAW: %v\n", move)
			}
		}
		for _, move := range res.GetArchives() {
			fmt.Printf("%v\n", move)
		}
	case "getclear":
		foldermap := make(map[int32]string)
		res, err := client.ListMoves(ctx, &pb.ListRequest{})
		if err != nil {
			log.Fatalf("Error on GET: %v", err)
		}
		for _, move := range res.GetMoves() {
			err = nil
			r := getRecord(ctx, move.InstanceId)
			f1, ok := foldermap[move.FromFolder]
			if !ok {
				f1, err = getFolder(ctx, move.FromFolder)
				if err != nil {
					log.Fatalf("Folder retrieve fail")
				}
				foldermap[move.FromFolder] = f1
			}
			f2, ok := foldermap[move.ToFolder]
			if !ok {
				f2, err = getFolder(ctx, move.ToFolder)
				if err != nil {
					log.Fatalf("Folder retreive fail")
				}
				foldermap[move.ToFolder] = f2
			}
			loc, err := getLocation(ctx, r)
			if err == nil {
				fmt.Printf("  %v", loc)
				_, err := client.ClearMove(ctx, &pb.ClearRequest{InstanceId: move.InstanceId})
				fmt.Printf("%v: %v -> %v\n", r.GetRelease().Title, f1, f2)
				fmt.Printf("CLEARED: %v\n", err)
			}
		}

	case "clear":
		res, err := client.ListMoves(ctx, &pb.ListRequest{})
		if err != nil {
			log.Fatalf("Error on GET: %v", err)
		}
		for _, move := range res.GetMoves() {
			_, err := client.ClearMove(ctx, &pb.ClearRequest{InstanceId: move.InstanceId})
			fmt.Printf("CLEARED: %v\n", err)
		}
	}
}
