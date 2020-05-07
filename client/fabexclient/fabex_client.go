/*
   Copyright 2019 Vadim Inshakov

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package fabexclient

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/hyperledger-labs/fabex/db"
	"github.com/hyperledger-labs/fabex/models"
	pb "github.com/hyperledger-labs/fabex/proto"
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
		return nil, errors.Wrap(err, "failed to connect")
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
		log.Printf("\nChannel ID: %s\nBlock number: %d\nBlock hash: %s\nPrevious hash: %s\nTx id: %s\nPayload: %s\nBlock timestamp: %d\n", in.Channelid, in.Blocknum, in.Hash, in.Previoushash, in.Txid, in.Payload, in.Time)
	}
}

func (fabexCli *FabexClient) GetByTxId(filter *pb.Entry) ([]db.Tx, error) {

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
		txs = append(txs, db.Tx{ChannelId: in.Channelid, Blocknum: in.Blocknum, Hash: in.Hash, PreviousHash: in.Previoushash, Txid: in.Txid, Payload: in.Payload, Time: in.Time})
	}
}

func (fabexCli *FabexClient) GetByBlocknum(filter *pb.Entry) ([]db.Tx, error) {

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
		txs = append(txs, db.Tx{ChannelId: in.Channelid, Blocknum: in.Blocknum, Hash: in.Hash, PreviousHash: in.Previoushash, Txid: in.Txid, Payload: in.Payload, Time: in.Time})
	}
}

func (fabexCli *FabexClient) GetBlockInfoByPayload(filter *pb.Entry) ([]db.Tx, error) {

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
		txs = append(txs, db.Tx{ChannelId: in.Channelid, Blocknum: in.Blocknum, Hash: in.Hash, PreviousHash: in.Previoushash, Txid: in.Txid, Payload: in.Payload, Time: in.Time})
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
			block = models.Block{ChannelId: in.ChannelId, Blocknum: in.Blocknum, BlockHash: in.Hash, PreviousHash: in.PreviousHash, Time: in.Time}
		}

		tx.Txid = in.Txid
		tx.ValidationCode = in.ValidationCode

		var ccData []models.Chaincode
		err := json.Unmarshal([]byte(in.Payload), &ccData)
		if err != nil {
			return nil, err
		}

		for _, item := range ccData {
			tx.KW = append(tx.KW, models.KW{item.Key, item.Value})
		}

		block.Txs = append(block.Txs, tx)
		Blocks = append(Blocks, block)
		blockAlreadyRead[in.Blocknum] = true
	}

	return Blocks, nil
}
