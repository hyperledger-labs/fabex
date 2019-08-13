package main

import (
	pb "fabex/proto"
	"fmt"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
)

func ReadStream(addr string, port string, startblock, endblock int64) error {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", addr, port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %s", err)
	}
	defer conn.Close()

	client := pb.NewFabexClient(conn)
	stream, err := client.Explore(context.Background(), &pb.Request{Startblock: startblock, Endblock: endblock})

	log.Println("Started stream")

	for {
		in, err := stream.Recv()
		log.Println("Received value")
		if err == io.EOF {
			log.Println("Steam is empty")
		}
		if err != nil {
			log.Println(err)
			return err
		}
		log.Printf("\nBlock number: %d\nBlock hash: %s\nTx id: %s\nPayload=%s\n", in.Blocknum, in.Hash, in.Txid, in.Payload)
	}

	return nil
}

func main() {
	ReadStream("0.0.0.0", "6000", 1, 15)
}
