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
	Blocknum  int    `json:"blocknum"`
	A         int    `json:"a"`
	B         int    `json:"b"`
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
	_, err := db.Instance.Exec(`
        CREATE TABLE txs (
    		txid TEXT PRIMARY KEY,
    		blockhash TEXT,
    	    blocknum INT,
    	    a INT,
    	    b INT
    	)`)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) Insert(txid, blockhash string, blocknum uint64, a, b int64) error {
	query := `INSERT INTO public.txs (txid, blockhash, blocknum, a, b) VALUES ($1, $2, $3, $4, $5);`
	_, err := db.Instance.Exec(query, txid, blockhash, blocknum, a, b)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) QueryAll(table string) ([]QueryResult, error) {
	arr := []QueryResult{}
	query := fmt.Sprintf(`
        SELECT * FROM public.%s;`, table)
	rows, err := db.Instance.Query(query)
	if err != nil {
		return []QueryResult{}, err
	}

	defer rows.Close()
	for rows.Next() {

		var (
			txid, blockhash string
			blocknum, a, b  int
		)

		err := rows.Scan(&txid, &blockhash, &blocknum, &a, &b)
		if err != nil {
			return []QueryResult{}, err
		}
		arr = append(arr, QueryResult{txid, blockhash, blocknum, a, b})
	}
	err = rows.Err()
	if err != nil {
		log.Fatal(err)
	}

	return arr, nil
}
