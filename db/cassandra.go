package db

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/gocql/gocql"
	"github.com/pkg/errors"
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
		return errors.WithStack(errors.Wrap(err, "cassandra system session creation failed"))
	}
	if err := c.Session.Query(fmt.Sprintf("CREATE KEYSPACE IF NOT EXISTS %s WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };", c.Keyspace)).Exec(); err != nil {
		return errors.WithStack(errors.Wrap(err, "failed to create keyspace"))
	}

	// reconnect with new keyspace
	cluster.Keyspace = c.Keyspace
	c.Session, err = cluster.CreateSession()
	if err != nil {
		return errors.WithStack(errors.Wrap(err, "cassandra client creation failed"))
	}
	return nil
}

func (c *Cassandra) Init(ch string) error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (ID UUID, %s text, %s text, %s text, %s text, %s bigint, %s text, %s int, %s int, %s list<text>, PRIMARY KEY(ID,%s))
		WITH CLUSTERING ORDER BY (%s DESC);`, fmt.Sprintf("%s_%s", ch, c.Columnfamily), CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, PAYLOADKEYS, BLOCKNUM, BLOCKNUM)
	if err := c.Session.Query(query).Exec(); err != nil {
		return errors.Wrapf(err, "failed to create column family: %s", c.Columnfamily)
	}

	// create payload index
	indexPayload := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS payload ON %s(%s);`, fmt.Sprintf("%s_%s", ch, c.Columnfamily), PAYLOADKEYS)
	if err := c.Session.Query(indexPayload).Exec(); err != nil {
		return errors.Wrapf(err, "failed to create index: %s", c.Columnfamily)
	}

	// create hash index
	indexHash := fmt.Sprintf(`CREATE INDEX IF NOT EXISTS hash ON %s(%s);`, fmt.Sprintf("%s_%s", ch, c.Columnfamily), HASH)
	if err := c.Session.Query(indexHash).Exec(); err != nil {
		return errors.Wrapf(err, "failed to create index: %s", c.Columnfamily)
	}

	// Normalization. We can't use slow aggregation queries, so create column family with last entry
	aggregationTable := fmt.Sprintf("CREATE TABLE IF NOT EXISTS MAX_%s (fortable text PRIMARY KEY, id UUID, blocknum bigint);", ch)
	if err := c.Session.Query(aggregationTable).Exec(); err != nil {
		return errors.Wrap(err, "failed to create column family: MAX")
	}
	return nil
}

func (c *Cassandra) Insert(ch string, tx Tx) error {
	insert := fmt.Sprintf("INSERT INTO %s (ID, %s, %s, %s, %s, %s, %s, %s, %s, %s) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", fmt.Sprintf("%s_%s", ch, c.Columnfamily),
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

	err = c.UpdateMax(ch, id, tx.Blocknum)

	return err
}

func (c *Cassandra) UpdateMax(ch string, id gocql.UUID, blocknum uint64) error {
	err := c.Session.Query(fmt.Sprintf("SELECT id FROM MAX_%s LIMIT 1;", ch)).Exec()
	if err != nil && err.Error() != NOT_FOUND_ERR {
		return errors.WithStack(err)
	}
	if err != nil && err.Error() == NOT_FOUND_ERR {
		if err = c.Session.Query(fmt.Sprintf("INSERT INTO MAX_%s (fortable, id, blocknum) VALUES (?, ?, ?)", ch), fmt.Sprintf("%s_%s", ch, c.Columnfamily), id, blocknum).Exec(); err != nil {
			return errors.WithStack(err)
		}

	}
	err = c.Session.Query(fmt.Sprintf(`UPDATE MAX_%s SET id = ?, blocknum = ? where fortable = ?;`, ch), id, blocknum, fmt.Sprintf("%s_%s", ch, c.Columnfamily)).Exec()
	return errors.WithStack(err)
}

func (c *Cassandra) GetBlockInfoByPayload(ch string, searchExpression string) ([]Tx, error) {
	query := fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE %s CONTAINS '%s'",
		CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, fmt.Sprintf("%s_%s", ch, c.Columnfamily), PAYLOADKEYS, searchExpression)
	txs, err := c.getByFilter(query, "")
	return txs, err
}

func (c *Cassandra) QueryBlockByHash(ch string, hash string) ([]Tx, error) {
	query := fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE %s = ?",
		CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, fmt.Sprintf("%s_%s", ch, c.Columnfamily), HASH)
	return c.getByFilter(query, hash)
}

func (c *Cassandra) GetByTxId(ch string, txID string) ([]Tx, error) {
	return c.getByFilter(fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE %s = ?",
		CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, fmt.Sprintf("%s_%s", ch, c.Columnfamily), TXID), txID)
}

func (c *Cassandra) QueryAll(ch string) ([]Tx, error) {
	return c.getByFilter(fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s FROM %s",
		CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, fmt.Sprintf("%s_%s", ch, c.Columnfamily)), "")
}

func (c *Cassandra) GetByBlocknum(ch string, blocknum uint64) ([]Tx, error) {
	return c.getByFilter(fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE %s = ?",
		CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, fmt.Sprintf("%s_%s", ch, c.Columnfamily), BLOCKNUM), strconv.FormatUint(blocknum, 10))
}

func (c *Cassandra) GetLastEntry(ch string) (Tx, error) {
	var (
		tx     Tx
		lastID string
	)

	/*
	   There are no nested queries in Cassandra, so we do this two-step shit for getting last tx
	*/

	// id (UUID) includes timestamp, so we use it for getting last tx ID
	err := c.Session.Query(fmt.Sprintf("SELECT id FROM MAX_%s where fortable = ?;", ch), fmt.Sprintf("%s_%s", ch, c.Columnfamily)).Scan(&lastID)
	if err != nil {
		return Tx{}, err
	}
	if lastID == "" {
		return Tx{}, errors.New("not found")
	}

	// get last tx using id as filter
	err = c.Session.Query(fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s FROM %s WHERE id = ? LIMIT 1",
		CHANNEL_ID, TXID, HASH, PREVIOUS_HASH, BLOCKNUM, PAYLOAD, VALIDATION_CODE, TIME, fmt.Sprintf("%s_%s", ch, c.Columnfamily)), lastID).Scan(
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
