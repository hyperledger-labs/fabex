package models

import (
	"fabex/db"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
)

type Fabex struct {
	Db            db.DbManager
	ChannelClient *channel.Client
	LedgerClient  *ledger.Client
}

type Db struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Dbuser   string `yaml:"dbuser"`
	Dbsecret string `yaml:"dbsecret"`
	Dbname   string `yaml:dbname`
}

type Config struct {
	DB Db `yaml:"Db"`
}

type Chaincode struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}
