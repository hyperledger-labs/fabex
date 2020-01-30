package fabexclient

import (
	"encoding/json"
	"fmt"
	"github.com/vadiminshakov/fabex/db"
	"github.com/vadiminshakov/fabex/models"
	pb "github.com/vadiminshakov/fabex/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
	"log"
)

type FabexClient struct {
	Client pb.FabexClient
}

func New(addr, port string) (*FabexClient, error) {

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", addr, port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("failed to connect: %s", err)
	}

	return &FabexClient{pb.NewFabexClient(conn)}, nil
}

func (fabexCli *FabexClient) Explore(startblock, endblock int) error {

	stream, err := fabexCli.Client.Explore(context.Background(), &pb.RequestRange{Startblock: int64(startblock), Endblock: int64(endblock)})
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
			log.Println(err)
			return err
		}
		log.Printf("\nBlock number: %d\nBlock hash: %s\nTx id: %s\nPayload: %s\n", in.Blocknum, in.Hash, in.Txid, in.Payload)
	}

	return nil
}

func (fabexCli *FabexClient) GetByTxId(filter *pb.RequestFilter) ([]db.Tx, error) {

	stream, err := fabexCli.Client.GetByTxId(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	log.Println("Started stream")

	var txs []db.Tx
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Println("Steam is empty")
			return txs, nil
		}
		if err != nil {
			return txs, err
		}
		txs = append(txs, db.Tx{Hash: in.Hash, Txid: in.Txid, Blocknum: in.Blocknum, Payload: in.Payload})
	}
}

func (fabexCli *FabexClient) GetByBlocknum(filter *pb.RequestFilter) ([]db.Tx, error) {

	stream, err := fabexCli.Client.GetByBlocknum(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	log.Println("Started stream")
	var txs []db.Tx
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Println("Steam is empty")
			return txs, nil
		}
		if err != nil {
			return txs, err
		}
		txs = append(txs, db.Tx{Hash: in.Hash, Txid: in.Txid, Blocknum: in.Blocknum, Payload: in.Payload})
	}
}

func (fabexCli *FabexClient) GetBlockInfoByPayload(filter *pb.RequestFilter) ([]db.Tx, error) {

	stream, err := fabexCli.Client.GetBlockInfoByPayload(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	log.Println("Started stream")

	var txs []db.Tx
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			log.Println("Steam is empty")
			return txs, nil
		}
		if err != nil {
			return txs, err
		}
		txs = append(txs, db.Tx{Hash: in.Hash, Txid: in.Txid, Blocknum: in.Blocknum, Payload: in.Payload})
	}
}

func (fabexCli *FabexClient) PackTxsToBlocks(blocks []db.Tx) ([]models.Block, error) {
	var blockAlreadyRead = make(map[uint64]bool)

	var Blocks []models.Block
	for _, in := range blocks {
		var (
			block models.Block
			tx    models.Tx
		)
		if _, ok := blockAlreadyRead[in.Blocknum]; !ok {
			block = models.Block{Blocknum: in.Blocknum, BlockHash: in.Hash}
		}
		tx.Txid = in.Txid

		var ccData []models.Chaincode
		err := json.Unmarshal([]byte(in.Payload), &ccData)
		if err != nil {
			return nil, err
		}

		for _, item := range ccData {
			tx.KW = append(tx.KW, models.KW{item.Key, item.Value})
		}

		block.Tx = append(block.Tx, tx)
		Blocks = append(Blocks, block)
		blockAlreadyRead[in.Blocknum] = true
	}

	return Blocks, nil
}
