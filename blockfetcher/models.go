package blockfetcher

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/vadiminshakov/fabex/db"
)

type Fabex struct {
	Db            db.DbManager
	ChannelClient *channel.Client
	LedgerClient  *ledger.Client
}

type DB struct {
	Host     string
	Port     int
	Dbuser   string
	Dbsecret string
	Dbname   string
}

type GRPCServer struct {
	Host string
	Port string
}

type Fabric struct {
	User              string
	Secret            string
	Org               string
	Channel           string
	ConnectionProfile string
}

type Config struct {
	DB
	Fabric
	GRPCServer
}

type Chaincode struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}