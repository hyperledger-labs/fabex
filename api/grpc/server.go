package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/hyperledger-labs/fabex/db"
	pb "github.com/hyperledger-labs/fabex/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func StartGrpcServ(_ context.Context, serv *FabexServer) error {
	grpcServer := grpc.NewServer()
	pb.RegisterFabexServer(grpcServer, serv)

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", serv.address, serv.port))
	if err != nil {
		return errors.WithStack(errors.Wrap(err, "failed to listen port"))
	}

	// start server
	if err := grpcServer.Serve(l); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

type FabexServer struct {
	pb.UnimplementedFabexServer
	address string
	port    string
	db      db.Storage
}

func NewFabexServer(addr string, port string, database db.Storage) *FabexServer {
	return &FabexServer{address: addr, port: port, db: database}
}

func (s *FabexServer) GetRange(req *pb.RequestRange, stream pb.Fabex_GetRangeServer) error {
	// set blocks counter to latest saved in db block number value
	blockCounter := req.Startblock

	// insert missing blocks/txs into db
	for blockCounter <= req.Endblock {
		QueryResults, err := s.db.GetByBlocknum(req.Channelid, uint64(blockCounter))
		if err != nil {
			return errors.Wrapf(err, "failed to get txs by block number %d", blockCounter)
		}
		if QueryResults != nil {
			for _, queryResult := range QueryResults {
				stream.Send(&pb.Entry{Channelid: queryResult.ChannelId, Txid: queryResult.Txid, Hash: queryResult.Hash, Previoushash: queryResult.PreviousHash, Blocknum: queryResult.Blocknum, Payload: queryResult.Payload, Time: queryResult.Time, Validationcode: queryResult.ValidationCode})
			}
		}
		blockCounter++
	}

	return nil
}

func (s *FabexServer) Get(req *pb.Entry, stream pb.Fabex_GetServer) error {
	switch {
	case req.Txid != "":
		QueryResults, err := s.db.GetByTxId(req.Channelid, req.Txid)
		if err != nil {
			return err
		}

		for _, queryResult := range QueryResults {
			stream.Send(&pb.Entry{Channelid: queryResult.ChannelId, Txid: queryResult.Txid, Hash: queryResult.Hash, Previoushash: queryResult.PreviousHash, Blocknum: queryResult.Blocknum, Payload: queryResult.Payload, Time: queryResult.Time, Validationcode: queryResult.ValidationCode})
		}
	case req.Blocknum != 0:
		QueryResults, err := s.db.GetByBlocknum(req.Channelid, req.Blocknum)
		if err != nil {
			return err
		}

		for _, queryResult := range QueryResults {
			stream.Send(&pb.Entry{Channelid: queryResult.ChannelId, Txid: queryResult.Txid, Hash: queryResult.Hash, Previoushash: queryResult.PreviousHash, Blocknum: queryResult.Blocknum, Payload: queryResult.Payload, Time: queryResult.Time, Validationcode: queryResult.ValidationCode})
		}
	default:
		// set blocks counter to latest saved in db block number value
		blockCounter := 1

		// insert missing blocks/txs into db
		for {
			queryResults, err := s.db.GetByBlocknum(req.Channelid, uint64(blockCounter))
			if err != nil {
				return errors.Wrapf(err, "failed to get txs by block number %d", blockCounter)
			}
			if queryResults == nil {
				break
			}
			for _, queryResult := range queryResults {
				stream.Send(&pb.Entry{Channelid: queryResult.ChannelId, Txid: queryResult.Txid, Hash: queryResult.Hash, Previoushash: queryResult.PreviousHash, Blocknum: queryResult.Blocknum, Payload: queryResult.Payload, Time: queryResult.Time, Validationcode: queryResult.ValidationCode})
			}

			blockCounter++
		}
	}

	return nil
}
