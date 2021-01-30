/*
   Copyright 2019 Vadim Inshakov

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hyperledger-labs/fabex/db"
	"github.com/hyperledger-labs/fabex/helpers"
	"github.com/hyperledger-labs/fabex/ledgerclient"
	"github.com/hyperledger-labs/fabex/models"
	pb "github.com/hyperledger-labs/fabex/proto"
	"github.com/hyperledger-labs/fabex/rest"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net"
	"os"
)

var (
	lvl          = logging.INFO
	globalConfig models.Config
)

type FabexServer struct {
	pb.UnimplementedFabexServer
	Address string
	Port    string
	Conf    *models.Fabex
}

func main() {

	// parse flags
	enrolluser := flag.Bool("enrolluser", false, "enroll user (true) or not (false)")
	task := flag.String("task", "grpc", "choose the task to execute")
	blocknum := flag.Uint64("blocknum", 0, "block number")
	confpath := flag.String("configpath", "./configs/", "path to YAML config")
	confname := flag.String("configname", "config", "name of YAML config")
	databaseSelected := flag.String("db", "mongo", "select database")
	ui := flag.Bool("ui", true, "with UI or without")

	flag.Parse()

	// read config
	viper.SetConfigName(*confname)
	viper.AddConfigPath(*confpath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&globalConfig)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	log.Println("Reading connection profile..")
	c := config.FromFile(globalConfig.Fabric.ConnectionProfile)
	sdk, err := fabsdk.New(c)
	if err != nil {
		log.Printf("Failed to create new SDK: %s\n", err)
		os.Exit(1)
	}
	defer sdk.Close()

	helpers.SetupLogLevel(lvl)
	if *enrolluser {
		err = helpers.EnrollUser(sdk, globalConfig.Fabric.User, globalConfig.Fabric.Secret)
		if err != nil {
			log.Fatal(err)
		}
	}

	clientChannelContext := sdk.ChannelContext(globalConfig.Fabric.Channel, fabsdk.WithUser(globalConfig.Fabric.User), fabsdk.WithOrg(globalConfig.Fabric.Org))
	ledgerClient, err := ledger.New(clientChannelContext)
	if err != nil {
		log.Fatalf("Failed to create ledger [%s] client: %#v", globalConfig.Fabric.Channel, err)
	}

	channelclient, err := channel.New(clientChannelContext)
	if err != nil {
		log.Fatalf("Failed to create channel [%s], error: %s", globalConfig.Fabric.Channel, err)
	}

	// choose database
	var dbInstance db.Storage
	switch *databaseSelected {
	case "mongo":
		dbInstance = db.CreateDBConfMongo(globalConfig.Mongo.Host, globalConfig.Mongo.Port, globalConfig.Mongo.Dbuser, globalConfig.Mongo.Dbsecret, globalConfig.Mongo.Dbname, globalConfig.Mongo.Collection)
	case "cassandra":
		dbInstance = db.NewCassandraClient(globalConfig.Cassandra.Host, globalConfig.Cassandra.Dbuser, globalConfig.Cassandra.Dbsecret, globalConfig.Cassandra.Keyspace, globalConfig.Cassandra.Columnfamily)
	}

	err = dbInstance.Connect()
	if err != nil {
		log.Fatal("DB connection failed:", err.Error())
	}
	log.Println("Connected to database successfully")

	fabex := &models.Fabex{dbInstance, channelclient, &ledgerclient.CustomLedgerClient{ledgerClient}, clientChannelContext}

	switch *task {
	case "channelinfo":
		resp, err := helpers.QueryChannelInfo(fabex.LedgerClient.Client)
		if err != nil {
			log.Fatalf("Can't query blockchain info: %s", err)
		}
		log.Printf("Blockchain height: %d\nCurrent block hash: %s\nPrevious block hash: %s\nEndorser: %v\nStatus: %v\n", resp.BCI.Height, hex.EncodeToString(resp.BCI.CurrentBlockHash), hex.EncodeToString(resp.BCI.PreviousBlockHash), resp.Endorser, resp.Status)

	case "channelconfig":
		cfg, err := helpers.QueryChannelConfig(fabex.LedgerClient.Client)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("ChannelID: %v\nChannel Orderers: %v\nChannel Versions: %v\n", cfg.ID(), cfg.Orderers(), cfg.Versions())

	case "getblock":
		txs, err := fabex.Db.GetByBlocknum(*blocknum)
		if err != nil {
			log.Fatalf("GetBlock error: %s", err)
		}

		if txs == nil {
			log.Fatal("empty block")
		}
		var cc []models.WriteKV
		for _, tx := range txs {
			err = json.Unmarshal(tx.Payload, &cc)
			if err != nil {
				log.Fatalf("Unmarshalling error: %s", err)
			}

			fmt.Printf("Channel ID: %s\nBlock number: %d\nBlock hash: %s\nPrevious hash: %s\nTxid: %s\nTx validation code: %d\nTime: %d\nPayload:\n",
				tx.ChannelId, tx.Blocknum, tx.Hash, tx.PreviousHash, tx.Txid, tx.ValidationCode, tx.Time)
			for _, val := range cc {
				fmt.Printf("Key: %s\nValue: %s\n", val.Key, val.Value)
			}
		}

	case "explore":
		log.Fatal(helpers.Explore(fabex))
	case "getall":
		allTxs, err := fabex.Db.QueryAll()
		if err != nil {
			log.Fatal("Can't query data: ", err)
		}

		for _, singleTx := range allTxs {
			fmt.Printf("Channel ID: %s\nBlock number: %d\nBlock hash: %s\nPrevious hash: %s\n",
				singleTx.ChannelId, singleTx.Blocknum, singleTx.Hash, singleTx.PreviousHash)

			var writeSet []models.WriteKV
			err = json.Unmarshal(singleTx.Payload, &writeSet)
			if err != nil {
				log.Fatalf("Unmarshalling error: %s", err)
			}

			fmt.Printf("Txid: %s\nTx validation code: %d\nTime: %d\nPayload:\n",
				singleTx.Txid, singleTx.ValidationCode, singleTx.Time)
			for _, val := range writeSet {
				decoded, err := base64.StdEncoding.DecodeString(val.Value)
				if err != nil {
					log.Fatal("base64 decoding error", err)
				}
				fmt.Printf("Key: %s\nValue: %s\n", val.Key, string(decoded))
			}
		}

	case "grpc":
		serv := NewFabexServer(globalConfig.GRPCServer.Host, globalConfig.GRPCServer.Port, fabex)

		// rest
		go rest.Run(serv.Conf.Db, globalConfig.UI.Port, *ui)
		// grpc
		StartGrpcServ(serv, fabex)
	}
}

func NewFabexServer(addr string, port string, conf *models.Fabex) *FabexServer {
	return &FabexServer{Address: addr, Port: port, Conf: conf}
}

func (s *FabexServer) GetRange(req *pb.RequestRange, stream pb.Fabex_GetRangeServer) error {
	// set blocks counter to latest saved in db block number value
	blockCounter := req.Startblock

	// insert missing blocks/txs into db
	for blockCounter <= req.Endblock {
		QueryResults, err := s.Conf.Db.GetByBlocknum(uint64(blockCounter))
		if err != nil {
			return errors.Wrapf(err, "failed to get txs by block number %d", blockCounter)
		}
		if QueryResults != nil {
			for _, queryResult := range QueryResults {
				stream.Send(&pb.Entry{Channelid: queryResult.ChannelId, Txid: queryResult.Txid, Hash: queryResult.Hash, Previoushash: queryResult.PreviousHash, Blocknum: queryResult.Blocknum, Payload: queryResult.Payload, Time: queryResult.Time, Validationcode: queryResult.ValidationCode})
			}
		}
		blockCounter++
	}

	return nil
}

func (s *FabexServer) Get(req *pb.Entry, stream pb.Fabex_GetServer) error {

	switch {
	case req.Txid != "":
		QueryResults, err := s.Conf.Db.GetByTxId(req.Txid)
		if err != nil {
			return err
		}

		for _, queryResult := range QueryResults {
			stream.Send(&pb.Entry{Channelid: queryResult.ChannelId, Txid: queryResult.Txid, Hash: queryResult.Hash, Previoushash: queryResult.PreviousHash, Blocknum: queryResult.Blocknum, Payload: queryResult.Payload, Time: queryResult.Time, Validationcode: queryResult.ValidationCode})
		}
	case req.Blocknum != 0:
		QueryResults, err := s.Conf.Db.GetByBlocknum(req.Blocknum)
		if err != nil {
			return err
		}

		for _, queryResult := range QueryResults {
			stream.Send(&pb.Entry{Channelid: queryResult.ChannelId, Txid: queryResult.Txid, Hash: queryResult.Hash, Previoushash: queryResult.PreviousHash, Blocknum: queryResult.Blocknum, Payload: queryResult.Payload, Time: queryResult.Time, Validationcode: queryResult.ValidationCode})
		}

		// DEPRECATED: payload is not string anymore, so we can't do search
	//case len(req.Payload) != 0:
	//	QueryResults, err := s.Conf.Db.GetBlockInfoByPayload(req.Payload)
	//	if err != nil {
	//		return err
	//	}
	//
	//	for _, queryResult := range QueryResults {
	//		stream.Send(&pb.Entry{Channelid: queryResult.ChannelId, Txid: queryResult.Txid, Hash: queryResult.Hash, Previoushash: queryResult.PreviousHash, Blocknum: queryResult.Blocknum, Payload: queryResult.Payload, Time: queryResult.Time, Validationcode: queryResult.ValidationCode})
	//	}
	default:
		// set blocks counter to latest saved in db block number value
		blockCounter := 1

		// insert missing blocks/txs into db
		for {
			queryResults, err := s.Conf.Db.GetByBlocknum(uint64(blockCounter))
			if err != nil {
				return errors.Wrapf(err, "failed to get txs by block number %d", blockCounter)
			}
			if queryResults == nil {
				break
			}
			for _, queryResult := range queryResults {
				stream.Send(&pb.Entry{Channelid: queryResult.ChannelId, Txid: queryResult.Txid, Hash: queryResult.Hash, Previoushash: queryResult.PreviousHash, Blocknum: queryResult.Blocknum, Payload: queryResult.Payload, Time: queryResult.Time, Validationcode: queryResult.ValidationCode})
			}

			blockCounter++
		}
	}

	return nil
}

func StartGrpcServ(serv *FabexServer, fabex *models.Fabex) {

	grpcServer := grpc.NewServer()
	pb.RegisterFabexServer(grpcServer, serv)

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", serv.Address, serv.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Listening on tcp://%s:%s", serv.Address, serv.Port)

	// start explorer
	go func() {
		log.Fatal(helpers.Explore(fabex))
	}()

	// start server
	grpcServer.Serve(l)
}
