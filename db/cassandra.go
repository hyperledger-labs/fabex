package db

import (
	"github.com/gocql/gocql"
	"log"
	"strconv"
)

type Cassandra struct {
	Host    string
	Session *gocql.Session
}

const (
	CREATE_KEYSPACE = " CREATE KEYSPACE IF NOT EXISTS " + KEYSPACE_NAME + " WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };"
	CREATE_TABLE    = "create table if not exists blockchain.txs (id uuid PRIMARY KEY, text text);"
	KEYSPACE_NAME   = "blockchain"

	TABLE           = "txs"
	CHANNEL_ID      = "ChannelId"
	TXID            = "Txid"
	HASH            = "Hash"
	PREVIOUS_HASH   = "PreviousHash"
	BLOCKNUM        = "Blocknum"
	PAYLOAD         = "Payload"
	VALIDATION_CODE = "ValidationCode"
	TIME            = "Time"
	INSERT          = "INSERT INTO " + TABLE + " (" + CHANNEL_ID + ", " + TXID + ", " + HASH + ", " +
		PREVIOUS_HASH + ", " + BLOCKNUM + ", " + PAYLOAD + ", " + VALIDATION_CODE + ", " + TIME + ") VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	SELECT = "SELECT " + CHANNEL_ID + ", " + TXID + ", " + HASH + ", " +
		PREVIOUS_HASH + ", " + BLOCKNUM + ", " + PAYLOAD + ", " + VALIDATION_CODE + ", " + TIME + " FROM " + TABLE
	SELECT_BLOCK_BY_HASH  = SELECT + " WHERE " + HASH + " = ?"
	SELECT_BY_TX_ID       = SELECT + " FROM " + TABLE + " WHERE " + TXID + " = ?"
	SELECT_LATEST_BY_TIME = "SELECT * " + " FROM " + TABLE + "ORDER BY " + TIME + " DESC LIMIT 1"
	SELECT_BY_BLOCKNUM    = SELECT + " WHERE " + BLOCKNUM + " = ?"
)

func NewCassandraClient(host string) *Cassandra {
	var err error
	cluster := gocql.NewCluster(host)
	cluster.Keyspace = KEYSPACE_NAME
	cluster.Consistency = gocql.One
	var c Cassandra
	c.Session, err = cluster.CreateSession()
	if err := c.Session.Query(CREATE_KEYSPACE).Exec(); err != nil {
		log.Fatal("failed to create keyspace", err)
	}
	if err != nil {
		log.Fatalf("cassandra client creation failed: %s", err)
	}
	err = c.Session.Query(CREATE_TABLE).Exec()
	if err != nil {
		log.Fatalf("failed to create table: %s", err)
	}
	return c

}

func (c *Cassandra) Init() error {
	return nil
}

func (c *Cassandra) Connect() error {
	panic("implement me")
}

func (c *Cassandra) GetBlockInfoByPayload(string) ([]Tx, error) {
	panic("implement me")
}

func (c *Cassandra) Insert(tx Tx) error {
	if err := c.Session.Query(INSERT, tx.ChannelId, tx.Txid, tx.Hash, tx.PreviousHash,
		tx.Blocknum, tx.Payload, tx.ValidationCode, tx.Time).Exec(); err != nil {
		return err
	}
	return nil
}

func (c *Cassandra) QueryBlockByHash(hash string) ([]Tx, error) {
	return c.getByFilter(SELECT_BLOCK_BY_HASH, hash)
}

func (c *Cassandra) GetByTxId(txID string) ([]Tx, error) {
	return c.getByFilter(SELECT_BY_TX_ID, txID)
}

func (c *Cassandra) QueryAll() ([]Tx, error) {
	return c.getByFilter(SELECT, "")
}

func (c *Cassandra) GetByBlocknum(blocknum uint64) ([]Tx, error) {
	return c.getByFilter(SELECT_BY_BLOCKNUM, strconv.FormatUint(blocknum, 10))
}

func (c *Cassandra) GetLastEntry() (Tx, error) {
	var tx Tx
	err := c.Session.Query(SELECT_LATEST_BY_TIME).Scan(&tx.ChannelId, &tx.Txid, &tx.Hash, &tx.PreviousHash, &tx.Blocknum,
		&tx.Payload, &tx.ValidationCode, &tx.Time)
	return tx, err
}

func (c *Cassandra) getByFilter(sel string, filter string) ([]Tx, error) {
	var tx Tx
	var txs []Tx
	var it *gocql.Iter
	if filter != "" {
		it = c.Session.Query(sel, filter).Iter()
	} else {
		it = c.Session.Query(sel).Iter()
	}
	for it.Scan(&tx.ChannelId, &tx.Txid, &tx.Hash, &tx.PreviousHash, &tx.Blocknum,
		&tx.Payload, &tx.ValidationCode, &tx.Time) {
		txs = append(txs, tx)
	}
	if err := it.Close(); err != nil {
		return nil, err
	}
	return txs, nil

}
