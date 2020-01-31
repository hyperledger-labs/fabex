package db

type DbManager interface {
	Connect() error
	Init() error
	Insert(string, string, uint64, string) error
	QueryBlockByHash(hash string) ([]Tx, error)
	GetByTxId(string) ([]Tx, error)
	GetByBlocknum(uint64) ([]Tx, error)
	GetBlockInfoByPayload(string) ([]Tx, error)
	QueryAll() ([]Tx, error)
}

type Tx struct {
	Txid     string `json:"txid" bson:"Txid"`
	Hash     string `json:"blockhash" bson:"Blockhash"`
	Blocknum uint64 `json:"blocknum" bson:"Blocknum"`
	Payload  string `json:"payload" bson:"Payload"`
}

type RequestFilter struct {
	Txid     string
	Hash     string
	Blocknum uint64
	Payload  string
}
