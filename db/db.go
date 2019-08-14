package db

type DbManager interface {
	Connect() error
	Init() error
	Insert(txid, blockhash string, blocknum uint64, payload []byte) error
	QueryBlockByHash(hash string) (*QueryResult, error)
	QueryAll() ([]QueryResult, error)
}

type QueryResult struct {
	Txid      string `json:"txid"`
	Blockhash string `json:"blockhash"`
	Blocknum  uint64 `json:"blocknum"`
	Payload   []byte `json:"payload"`
}

type QueryResultMongo struct {
	Txid      string                       `json:"txid"`
	Blockhash string                       `json:"blockhash"`
	Blocknum  map[string]uint64            `json:"blocknum"`
	Payload   map[string]map[string]string `json:"payload"`
}
