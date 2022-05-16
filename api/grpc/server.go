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
	if req.Channelid == "" {
		return errors.New("no channel ID specified")
	}

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
				if err = stream.Send(&pb.Entry{
					Channelid:      queryResult.ChannelId,
					Txid:           queryResult.Txid,
					Hash:           queryResult.Hash,
					Previoushash:   queryResult.PreviousHash,
					Blocknum:       queryResult.Blocknum,
					Payload:        queryResult.Payload,
					Time:           queryResult.Time,
					Validationcode: queryResult.ValidationCode}); err != nil {
					return err
				}
			}
		}
		blockCounter++
	}

	return nil
}

func (s *FabexServer) Get(req *pb.Entry, stream pb.Fabex_GetServer) error {
	if req.Channelid == "" {
		return errors.New("no channel ID specified")
	}

	switch {
	case req.Txid != "":
		queryFunc := func() ([]db.Tx, error) {
			return s.db.GetByTxId(req.Channelid, req.Txid)
		}
		return query(stream, queryFunc)

	case req.Blocknum != 0:
		queryFunc := func() ([]db.Tx, error) {
			return s.db.GetByBlocknum(req.Channelid, req.Blocknum)
		}
		return query(stream, queryFunc)

	case req.Payload != nil:
		queryFunc := func() ([]db.Tx, error) {
			return s.db.GetBlockInfoByPayload(req.Channelid, string(req.Payload))
		}
		return query(stream, queryFunc)

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
			if err = sendStream(stream, queryResults); err != nil {
				return err
			}

			blockCounter++
		}
	}

	return nil
}

func query(stream pb.Fabex_GetServer, queryf func() ([]db.Tx, error)) error {
	queryResults, err := queryf()
	if err != nil {
		return err
	}

	return sendStream(stream, queryResults)
}

func sendStream(stream pb.Fabex_GetServer, queryResults []db.Tx) error {
	for _, qr := range queryResults {
		if err := stream.Send(&pb.Entry{
			Channelid:      qr.ChannelId,
			Txid:           qr.Txid,
			Hash:           qr.Hash,
			Previoushash:   qr.PreviousHash,
			Blocknum:       qr.Blocknum,
			Payload:        qr.Payload,
			Time:           qr.Time,
			Validationcode: qr.ValidationCode}); err != nil {
			return err
		}
	}
	return nil
}
