package main

import (
	"encoding/json"
	"fmt"

	"github.com/vadiminshakov/fabex/blockfetcher"
	pb "github.com/vadiminshakov/fabex/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
)

var (
	client pb.FabexClient
	addr = "0.0.0.0"
	port = "6000"
)

func init() {
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", addr, port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %s", err)
	}

	client = pb.NewFabexClient(conn)
}

func ReadStream(startblock, endblock int) error {

	stream, err := client.Explore(context.Background(), &pb.RequestRange{Startblock: int64(startblock), Endblock: int64(endblock)})
	if err != nil {
		return err
	}

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
		log.Printf("\nBlock number: %d\nBlock hash: %s\nTx id: %s\nPayload:\n", in.Blocknum, in.Hash, in.Txid)
		var cc []blockfetcher.Chaincode
		err = json.Unmarshal(in.Payload, &cc)
		if err != nil {
			log.Printf("Unmarshalling error: %s", err)
		}
		for _, val := range cc {

			fmt.Printf("Key: %s\nValue: %s\n", val.Key, in.Payload)
		}
	}

	return nil
}

func GetByTxId(filter *pb.RequestFilter) error {

	stream, err := client.GetByTxId(context.Background(), filter)
	if err != nil {
		return err
	}

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
		log.Printf("\nBlock number: %d\nBlock hash: %s\nTx id: %s\nPayload:\n", in.Blocknum, in.Hash, in.Txid)
		var cc []blockfetcher.Chaincode
		err = json.Unmarshal(in.Payload, &cc)
		if err != nil {
			log.Printf("Unmarshalling error: %s", err)
		}
		for _, val := range cc {

			fmt.Printf("Key: %s\nValue: %s\n", val.Key, in.Payload)
		}
	}

	return nil
}

func GetByBlocknum(filter *pb.RequestFilter) error {

	stream, err := client.GetByBlocknum(context.Background(), filter)
	if err != nil {
		return err
	}

	log.Println("Started stream")

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Println("Steam is empty")
			return nil
		}
		if err != nil {
			return err
		}
		log.Printf("\nBlock number: %d\nBlock hash: %s\nTx id: %s\nPayload: %s\n", in.Blocknum, in.Hash, in.Txid, in.Payload)


	}

	return nil
}

func main() {
	//ReadStream(1, 15)
	//err := GetByTxId(&pb.RequestFilter{TxId:2})
	err := GetByBlocknum(&pb.RequestFilter{Blocknum:2})
	if err != nil { log.Fatal(err) }

}
