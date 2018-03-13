package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver/utils"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/recordmover/proto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
)

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
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	switch os.Args[1] {
	case "get":
		res, err := client.ListMoves(ctx, &pb.ListRequest{})
		if err != nil {
			log.Fatalf("Error on GET: %v", err)
		}
		for _, move := range res.GetMoves() {
			fmt.Printf("%v: %v -> %v\n", move.InstanceId, move.FromFolder, move.ToFolder)
		}
	case "getclear":
		res, err := client.ListMoves(ctx, &pb.ListRequest{})
		if err != nil {
			log.Fatalf("Error on GET: %v", err)
		}
		for _, move := range res.GetMoves() {
			fmt.Printf("%v: %v -> %v\n", move.InstanceId, move.FromFolder, move.ToFolder)
			client.ClearMove(ctx, &pb.ClearRequest{InstanceId: move.InstanceId})
		}
	}

}
