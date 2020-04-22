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
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/vadiminshakov/fabex/blockfetcher"
	"github.com/vadiminshakov/fabex/db"
	"github.com/vadiminshakov/fabex/helpers"
	"github.com/vadiminshakov/fabex/models"
	pb "github.com/vadiminshakov/fabex/proto"
	"google.golang.org/grpc"
	"net"
	"os"
)

var (
	lvl          = logging.INFO
	globalConfig models.Config
)

func main() {

	// parse flags
	enrolluser := flag.Bool("enrolluser", false, "enroll user (true) or not (false)")
	task := flag.String("task", "query", "choose the task to execute")
	forever := flag.Bool("forever", false, "explore ledger forever")
	blocknum := flag.Uint64("blocknum", 0, "block number")
	confpath := flag.String("configpath", "./", "path to YAML config")
	confname := flag.String("configname", "config", "name of YAML config")
	databaseSelected := flag.String("db", "mongo", "select database")

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
		log.Fatalf("Failed to create channel [%s] client: %#v", globalConfig.Fabric.Channel, err)
	}

	channelclient, err := channel.New(clientChannelContext)
	if err != nil {
		log.Fatalf("Failed to create channel [%s], error: %s", globalConfig.Fabric.Channel, err)
	}

	// choose database
	var dbInstance db.DbManager
	switch *databaseSelected {
	case "mongo":
		dbInstance = db.CreateDBConfMongo(globalConfig.DB.Host, globalConfig.DB.Port, globalConfig.DB.Dbuser, globalConfig.DB.Dbsecret, globalConfig.DB.Dbname, globalConfig.DB.Collection)
	}

	var fabex *models.Fabex
	if *task != "initdb" {
		err = dbInstance.Connect()
		if err != nil {
			log.Fatal("DB connection failed:", err.Error())
		}
	}
	fabex = &models.Fabex{dbInstance, channelclient, ledgerClient}

	switch *task {
	case "initdb":
		err = fabex.Db.Init()
		if err != nil {
			log.Fatalf("Failed to create table: %s", err)
		}
		log.Println("Database and table created successfully")

	case "channelinfo":
		resp, err := helpers.QueryChannelInfo(fabex.LedgerClient)
		if err != nil {
			log.Fatalf("Can't query blockchain info: %s", err)
		}
		log.Printf("BlockChainInfo: %v\nEndorser: %v\nStatus: %v\n", resp.BCI, resp.Endorser, resp.Status)

	case "channelconfig":
		err = helpers.QueryChannelConfig(fabex.LedgerClient)
		if err != nil {
			log.Fatal(err)
		}

	case "getblock":
		customBlock, err := blockfetcher.GetBlock(fabex.LedgerClient, *blocknum)
		if err != nil {
			log.Fatalf("GetBlock error: %s", err)
		}

		if customBlock != nil {
			var cc []models.Chaincode
			for _, block := range customBlock.Txs {

				err = json.Unmarshal([]byte(block.Payload), &cc)
				if err != nil {
					log.Fatalf("Unmarshalling error: %s", err)
				}

				log.Printf("\nBlock number: %d\nBlock hash: %s\nTxid: %s\nPayload:\n", block.Blocknum, block.Hash, block.Txid)
				for _, val := range cc {
					log.Fatalf("Key: %s\nValue: %s\n", val.Key, val.Value)
				}

			}
		}

	case "explore":
		if *forever {
			for {
				err = helpers.Explore(fabex)
				if err != nil {
					log.Fatal(err)
				}
			}
		} else {
			err = helpers.Explore(fabex)
			if err != nil {
				log.Fatal(err)
			}

			log.Println("All blocks saved")
		}

	case "getall":
		txs, err := fabex.Db.QueryAll()
		if err != nil {
			log.Fatal("Can't query data: ", err)
		}

		for _, tx := range txs {

			var cc []models.Chaincode

			err = json.Unmarshal([]byte(tx.Payload), &cc)
			if err != nil {
				log.Fatalf("Unmarshalling error: %s", err)
			}

			log.Printf("\nBlock number: %d\nBlock hash: %s\nTxid: %s\nPayload:\n", tx.Blocknum, tx.Hash, tx.Txid)
			for _, val := range cc {
				log.Fatalf("Key: %s\nValue: %s\n", val.Key, val.Value)
			}

		}

	case "grpc":
		serv := NewFabexServer(globalConfig.GRPCServer.Host, globalConfig.GRPCServer.Port, fabex)
		StartGrpcServ(serv, fabex)
	}
}

type fabexServer struct {
	Address string
	Port    string
	Conf    *models.Fabex
}

func NewFabexServer(addr string, port string, conf *models.Fabex) *fabexServer {
	return &fabexServer{addr, port, conf}
}

func (s *fabexServer) Explore(req *pb.RequestRange, stream pb.Fabex_ExploreServer) error {
	log.Printf("Start stream from %d block", req.Startblock)
	// set blocks counter to latest saved in db block number value
	var blockCounter uint64 = uint64(req.Startblock)

	// insert missing blocks/txs into db
	for blockCounter <= uint64(req.Endblock) {
		customBlock, err := blockfetcher.GetBlock(s.Conf.LedgerClient, blockCounter)
		if err != nil {
			return errors.Wrap(err, "GetBlock error")
		}

		if customBlock != nil {
			for _, queryResult := range customBlock.Txs {
				stream.Send(&pb.Entry{Channelid: queryResult.ChannelId, Txid: queryResult.Txid, Hash: queryResult.Hash, Previoushash: queryResult.PreviousHash, Blocknum: queryResult.Blocknum, Payload: queryResult.Payload, Time: queryResult.Time})
			}
		}
		blockCounter++
	}

	return nil
}

func (s *fabexServer) GetByTxId(req *pb.Entry, stream pb.Fabex_GetByTxIdServer) error {

	QueryResults, err := s.Conf.Db.GetByTxId(req.Txid)
	if err != nil {
		return err
	}

	for _, queryResult := range QueryResults {
		stream.Send(&pb.Entry{Channelid: queryResult.ChannelId, Txid: queryResult.Txid, Hash: queryResult.Hash, Previoushash: queryResult.PreviousHash, Blocknum: queryResult.Blocknum, Payload: queryResult.Payload, Time: queryResult.Time})
	}

	return nil
}

func (s *fabexServer) GetByBlocknum(req *pb.Entry, stream pb.Fabex_GetByBlocknumServer) error {
	QueryResults, err := s.Conf.Db.GetByBlocknum(req.Blocknum)
	if err != nil {
		return err
	}

	for _, queryResult := range QueryResults {
		stream.Send(&pb.Entry{Channelid: queryResult.ChannelId, Txid: queryResult.Txid, Hash: queryResult.Hash, Previoushash: queryResult.PreviousHash, Blocknum: queryResult.Blocknum, Payload: queryResult.Payload, Time: queryResult.Time})
	}

	return nil
}

func (s *fabexServer) GetBlockInfoByPayload(req *pb.Entry, stream pb.Fabex_GetBlockInfoByPayloadServer) error {
	QueryResults, err := s.Conf.Db.GetBlockInfoByPayload(req.Payload)
	if err != nil {
		return err
	}

	for _, queryResult := range QueryResults {
		stream.Send(&pb.Entry{Channelid: queryResult.ChannelId, Txid: queryResult.Txid, Hash: queryResult.Hash, Previoushash: queryResult.PreviousHash, Blocknum: queryResult.Blocknum, Payload: queryResult.Payload, Time: queryResult.Time})
	}

	return nil
}

func StartGrpcServ(serv *fabexServer, fabex *models.Fabex) {

	grpcServer := grpc.NewServer()
	pb.RegisterFabexServer(grpcServer, serv)

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", serv.Address, serv.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("Listening on tcp://%s:%s", serv.Address, serv.Port)

	// start explorer
	go func() {
		for {
			err = helpers.Explore(fabex)
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	// start server
	grpcServer.Serve(l)
}
