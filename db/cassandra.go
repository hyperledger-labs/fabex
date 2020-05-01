package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

type Cassandra struct {
	Host         string
	User         string
	Password     string
	Keyspace     string
	Columnfamily string
	Session      *gocql.Session
}

var nsKeySep = []byte{0x00}

const (
	CHANNEL_ID      = "ChannelId"
	TXID            = "Txid"
	HASH            = "Hash"
	PREVIOUS_HASH   = "PreviousHash"
	BLOCKNUM        = "Blocknum"
	PAYLOAD         = "Payload"
	VALIDATION_CODE = "ValidationCode"
	TIME            = "Time"
	PAYLOADKEYS     = "Payloadkeys"
)

func NewCassandraClient(host, user, password, keyspace, columnfamily string) *Cassandra {
	return &Cassandra{host, user, password, keyspace, columnfamily, nil}
}

func (c *Cassandra) Connect() error {
	var err error
	cluster := gocql.NewCluster(c.Host)
	cluster.Timeout = 15 * time.Second
	cluster.ConnectTimeout = 15 * time.Second
	cluster.Keyspace = "system"
	cluster.Consistency = gocql.One
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: c.User,
		Password: c.Password,
	}
	c.Session, err = cluster.CreateSession()
	if err != nil {
		return errors.Wrap(err, "cassandra system session creation failed")
	}
	if err := c.Session.Query(fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };", c.Keyspace)).Exec(); err != nil {
		return errors.Wrap(err, "failed to create keyspace")
	}

	// reconnect with new keyspace
	cluster.Keyspace = c.Keyspace
	c.Session, err = cluster.CreateSession()
	if err != nil {
		return errors.Wrap(err, "cassandra client creation failed")
	}
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (ID UUID, %s text, %s text, %s text, %s text, %s bigint, %s text, %s int, %s int, %s list<text>, PRIMARY KEY(ID,%s))
		WITH CLUSTERING ORDER BY (%s DESC);`, c.Columnfamily, CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, PAYLOADKEYS, BLOCKNUM, BLOCKNUM)
	err = c.Session.Query(query).Exec()
	if err != nil {
		return errors.Wrapf(err, "failed to create column family: %s", c.Columnfamily)
	}

	// create payload index
	index := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS payload ON %s(%s);`, c.Columnfamily, PAYLOADKEYS)
	err = c.Session.Query(index).Exec()
	if err != nil {
		return errors.Wrapf(err, "failed to create index: %s", c.Columnfamily)
	}

	return nil
}

func (c *Cassandra) Insert(tx Tx) error {
	insert := fmt.Sprintf("INSERT INTO %s (ID, %s, %s, %s, %s, %s, %s, %s, %s, %s) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", c.Columnfamily,
		CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, PAYLOADKEYS)

	var Payload []RW
	err := json.Unmarshal([]byte(tx.Payload), &Payload)
	if err != nil {
		return err
	}

	// extract metadata from RWSet and insert to column (clear from separators)
	var payloadkeys []string
	for _, kv := range Payload {
		if bytes.Index([]byte(kv.Key), nsKeySep) != -1 {
			split := bytes.SplitN([]byte(kv.Key), nsKeySep, -1)
			payloadkeys = append(payloadkeys, string(split[2][0:]))
		} else {
			payloadkeys = append(payloadkeys, kv.Key)
		}
	}

	if err := c.Session.Query(insert, gocql.TimeUUID(), tx.ChannelId, tx.Txid, tx.Hash, tx.PreviousHash,
		tx.Blocknum, tx.Payload, tx.ValidationCode, tx.Time, payloadkeys).Exec(); err != nil {
		return err
	}

	return nil
}

func (c *Cassandra) GetBlockInfoByPayload(searchExpression string) ([]Tx, error) {
	query := fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE %s CONTAINS '%s'",
		CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, c.Columnfamily, PAYLOADKEYS, searchExpression)
	txs, err := c.getByFilter(query, "")
	return txs, err
}

func (c *Cassandra) QueryBlockByHash(hash string) ([]Tx, error) {
	query := fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE %s = ?",
		CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, c.Columnfamily, HASH)
	return c.getByFilter(query, hash)
}

func (c *Cassandra) GetByTxId(txID string) ([]Tx, error) {
	return c.getByFilter(fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE %s = ?",
		CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, c.Columnfamily, TXID), txID)
}

func (c *Cassandra) QueryAll() ([]Tx, error) {
	return c.getByFilter(fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s FROM %s",
		CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, c.Columnfamily), "")
}

func (c *Cassandra) GetByBlocknum(blocknum uint64) ([]Tx, error) {
	return c.getByFilter(fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE %s = ?",
		CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, c.Columnfamily, BLOCKNUM), strconv.FormatUint(blocknum, 10))
}

func (c *Cassandra) GetLastEntry() (Tx, error) {
	var (
		tx       Tx
		lastID string
	)

	/*
	    There are no nested queries in Cassandra, so we do this two-step shit for getting last tx
	 */

	// id (UUID) includes timestamp, so we use it for getting last tx ID
	err := c.Session.Query(fmt.Sprintf("SELECT MAX(id) FROM %s", c.Columnfamily)).Scan(&lastID)
	if err != nil {
		return Tx{}, err
	}
	if lastID == "" {
		return Tx{}, errors.New("not found")
	}

	// get last tx using id as filter
	err = c.Session.Query(fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE id = ? LIMIT 1",
		CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, c.Columnfamily), lastID).Scan(
		&tx.ChannelId, &tx.Txid, &tx.Hash, &tx.PreviousHash, &tx.Blocknum, &tx.Payload, &tx.ValidationCode, &tx.Time)

	return tx, err
}

func (c *Cassandra) getByFilter(sel string, filter string) ([]Tx, error) {
	var tx Tx
	var txs []Tx
	var it *gocql.Iter
	if filter != "" {
		it = c.Session.Query(fmt.Sprintf("%s ALLOW FILTERING", sel), filter).Iter()
	} else {
		it = c.Session.Query(fmt.Sprintf("%s", sel)).Iter()
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
