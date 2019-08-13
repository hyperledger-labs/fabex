package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

type DB struct {
	Host     string
	Port     int
	User     string
	Password string
	DBname   string
	Instance *sql.DB
}

type QueryResult struct {
	Txid      string `json:"txid"`
	Blockhash string `json:"blockhash"`
	Blocknum  uint64 `json:"blocknum"`
	Payload   []byte `json:"payload"`
}

func CreateDBConf(host string, port int, user, password, dbname string) *DB {
	return &DB{host, port, user, password, dbname, &sql.DB{}}
}

func (db *DB) Connect() error {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		db.Host, db.Port, db.User, db.Password, db.DBname)

	dbinstance, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}

	err = dbinstance.Ping()
	if err != nil {
		return err
	}

	dbinstance.SetMaxIdleConns(15)
	db.Instance = dbinstance
	return nil
}

func (db *DB) Init() error {

	// create simple table with transaction id, block hash, block number
	// keys a and b (for simple chaincode)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s sslmode=disable",
		db.Host, db.Port, db.User, db.Password)

	dbinstance, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}

	err = dbinstance.Ping()
	if err != nil {
		return err
	}

	_, err = dbinstance.Exec(`
        CREATE DATABASE txs
    	`)
	if err != nil {
		return err
	}

	// close old connection and reconnect to new database
	dbinstance.Close()
	psqlInfo = fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		db.Host, db.Port, db.User, db.Password, "txs")

	dbinstance, err = sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}

	_, err = dbinstance.Exec(`
        CREATE TABLE txs (
    		txid TEXT PRIMARY KEY,
    		blockhash TEXT,
    	    blocknum INT,
    	    payload BYTEA
    	)`)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) Insert(txid, blockhash string, blocknum uint64, payload []byte) error {
	query := `INSERT INTO public.txs (txid, blockhash, blocknum, payload) VALUES ($1, $2, $3, $4);`
	_, err := db.Instance.Exec(query, txid, blockhash, blocknum, payload)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) QueryBlockByHash(hash string) (*QueryResult, error) {
	query := fmt.Sprintf(`
        SELECT (txid, blockhash, blocknum, payload) FROM public.txs
        WHERE blockhash='%s';`, hash)

	var (
		txid, blockhash string
		blocknum        uint64
		payload         []byte
	)

	err := db.Instance.QueryRow(query).Scan(&txid, &blockhash, &blocknum, &payload)
	if err != nil {
		return nil, err
	}

	return &QueryResult{txid, blockhash, blocknum, payload}, nil
}

func (db *DB) QueryAll() ([]QueryResult, error) {
	arr := []QueryResult{}
	query := fmt.Sprintf(`
        SELECT * FROM public.txs;`)
	rows, err := db.Instance.Query(query)
	if err != nil {
		return []QueryResult{}, err
	}

	defer rows.Close()
	for rows.Next() {

		var (
			txid, blockhash string
			blocknum        uint64
			payload         []byte
		)

		err := rows.Scan(&txid, &blockhash, &blocknum, &payload)
		if err != nil {
			return []QueryResult{}, err
		}
		arr = append(arr, QueryResult{txid, blockhash, blocknum, payload})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return arr, nil
}
