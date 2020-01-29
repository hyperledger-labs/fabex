package db

import (
	pb "github.com/vadiminshakov/fabex/proto"
)

type DbManager interface {
	Connect() error
	Init() error
	Insert(txid, blockhash string, blocknum uint64, payload string) error
	QueryBlockByHash(hash string) (Tx, error)
	GetByTxId(filter *pb.RequestFilter) ([]Tx, error)
	GetByBlocknum(req *pb.RequestFilter) ([]Tx, error)
	GetBlockInfoByPayload(req *pb.RequestFilter) ([]Tx, error)
	QueryAll() ([]Tx, error)
}

type Tx struct {
	Txid     string `json:"txid" bson:"Txid"`
	Hash     string `json:"blockhash" bson:"Blockhash"`
	Blocknum uint64 `json:"blocknum" bson:"Blocknum"`
	Payload  string `json:"payload" bson:"Payload"`
}