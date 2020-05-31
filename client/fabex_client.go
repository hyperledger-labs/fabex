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

package client

import (
	"fmt"
	"github.com/hyperledger-labs/fabex/db"
	pb "github.com/hyperledger-labs/fabex/proto"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"io"
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

func (fabexCli *FabexClient) Explore(startblock, endblock int) ([]db.Tx, error) {

	stream, err := fabexCli.Client.Explore(context.Background(), &pb.RequestRange{Startblock: int64(startblock), Endblock: int64(endblock)})
	if err != nil {
		return nil, err
	}

	var txs []db.Tx
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return txs, nil
		}
		if err != nil {
			return txs, err
		}
		txs = append(txs, db.Tx{ChannelId: in.Channelid, Blocknum: in.Blocknum, Hash: in.Hash, PreviousHash: in.Previoushash, Txid: in.Txid, Payload: in.Payload, Time: in.Time})
	}

}

func (fabexCli *FabexClient) Get(filter *pb.Entry) ([]db.Tx, error) {

	stream, err := fabexCli.Client.Get(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	var txs []db.Tx
	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return txs, nil
		}
		if err != nil {
			return txs, err
		}
		txs = append(txs, db.Tx{ChannelId: in.Channelid, Blocknum: in.Blocknum, Hash: in.Hash, PreviousHash: in.Previoushash, Txid: in.Txid, Payload: in.Payload, Time: in.Time})
	}
}
