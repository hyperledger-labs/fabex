package db

import (
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
	UPDATE_MAX      = `UPDATE MAX SET id = ?, blocknum = ? where fortable = ?;`
)

func NewCassandraClient(host, user, password, keyspace, columnfamily string) *Cassandra {
	return &Cassandra{host, user, password, keyspace, columnfamily, nil}
}

func (c *Cassandra) Connect() error {
	var err error
	cluster := gocql.NewCluster(c.Host)
	cluster.Timeout = 1000 * time.Second
	cluster.ConnectTimeout = 30 * time.Second
	cluster.Keyspace = "system"
	cluster.Consistency = gocql.All
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

	// Normalization. We can't use slow aggregation queries, so create column family with last entry
	aggregationTable := `CREATE TABLE IF NOT EXISTS MAX (fortable text PRIMARY KEY, id UUID, blocknum bigint);`
	err = c.Session.Query(aggregationTable).Exec()
	if err != nil {
		return errors.Wrap(err, "failed to create column family: MAX")
	}

	return nil
}

func (c *Cassandra) Insert(tx Tx) error {
	insert := fmt.Sprintf("INSERT INTO %s (ID, %s, %s, %s, %s, %s, %s, %s, %s, %s) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", c.Columnfamily,
		CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, PAYLOADKEYS)

	var Payload []RW
	err := json.Unmarshal(tx.Payload, &Payload)
	if err != nil {
		return err
	}

	// extract keys from RWSet
	var payloadkeys []string
	for _, kv := range Payload {
		payloadkeys = append(payloadkeys, kv.Key)
	}

	id := gocql.TimeUUID()
	if err := c.Session.Query(insert, id, tx.ChannelId, tx.Txid, tx.Hash, tx.PreviousHash,
		tx.Blocknum, tx.Payload, tx.ValidationCode, tx.Time, payloadkeys).Exec(); err != nil {
		return err
	}

	err = c.UpdateMax(id, tx.Blocknum)

	return err
}

func (c *Cassandra) UpdateMax(id gocql.UUID, blocknum uint64) error {
	err := c.Session.Query("SELECT id FROM MAX LIMIT 1;").Exec()
	if err != nil && err.Error() != NOT_FOUND_ERR {
		return err
	}
	if err != nil && err.Error() == NOT_FOUND_ERR {
		err = c.Session.Query("INSERT INTO MAX (fortable, id, blocknum) VALUES (?, ?, ?)", c.Columnfamily, id, blocknum).Exec()
	}
	err = c.Session.Query(UPDATE_MAX, id, blocknum, c.Columnfamily).Exec()
	return err
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
		tx     Tx
		lastID string
	)

	/*
	   There are no nested queries in Cassandra, so we do this two-step shit for getting last tx
	*/

	// id (UUID) includes timestamp, so we use it for getting last tx ID
	err := c.Session.Query("SELECT id FROM MAX where fortable = ?;", c.Columnfamily).Scan(&lastID)
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
	var txs []Tx
	var sc gocql.Scanner
	if filter != "" {
		sc = c.Session.Query(fmt.Sprintf("%s ALLOW FILTERING", sel), filter).Iter().Scanner()
	} else {
		sc = c.Session.Query(fmt.Sprintf("%s", sel)).Iter().Scanner()
	}
	for sc.Next() {
		var tx Tx
		if err := sc.Scan(&tx.ChannelId, &tx.Txid, &tx.Hash, &tx.PreviousHash, &tx.Blocknum,
			&tx.Payload, &tx.ValidationCode, &tx.Time); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return txs, nil
}
