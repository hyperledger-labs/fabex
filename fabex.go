package main

import (
	"encoding/hex"
	"fabex/blockfetcher"
	"fabex/db"
	"fabex/helpers"
	pb "fabex/proto"
	"flag"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"google.golang.org/grpc"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net"
	"os"
)

var (
	cc          *string
	user        *string
	secret      *string
	channelName *string
	lvl         = logging.INFO
)

type Fabex struct {
	db            *db.DB
	channelClient *channel.Client
	ledgerClient  *ledger.Client
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

func main() {

	// parse flags
	cc = flag.String("cc", "mycc", "chaincode name")
	user = flag.String("user", "admin", "user name")
	secret = flag.String("secret", "adminpw", "user secret")
	channelName = flag.String("channel", "mychannel", "channel name")
	enrolluser := flag.Bool("enrolluser", false, "enroll user (true) or not (false)")
	task := flag.String("task", "query", "choose the task to execute")
	blocknum := flag.Uint64("blocknum", 0, "block number")
	confpath := flag.String("config", "./config.yaml", "path to YAML config")
	profile := flag.String("profile", "./connection-profile.yaml", "path to connection profile")
	grpcAddr := flag.String("grpcaddr", "0.0.0.0", "grpc server address")
	grpcPort := flag.String("grpcport", "6000", "grpc server port")

	flag.Parse()

	// read config
	data, err := ioutil.ReadFile(*confpath)
	if err != nil {
		log.Println("Reading file error: ")
		return
	}

	var globalConfig Config
	err = yaml.Unmarshal([]byte(data), &globalConfig)
	if err != nil {
		log.Println("Unmarshalling error: ")
		return
	}

	fmt.Println("Reading connection profile..")
	c := config.FromFile(*profile)
	sdk, err := fabsdk.New(c)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		os.Exit(1)
	}
	defer sdk.Close()

	helpers.SetupLogLevel(lvl)
	if *enrolluser {
		helpers.EnrollUser(sdk, *user, *secret)
	}

	clientChannelContext := sdk.ChannelContext(*channelName, fabsdk.WithUser(*user), fabsdk.WithOrg("Org1"))
	ledgerClient, err := ledger.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to create channel [%s] client: %#v", *channelName, err)
		os.Exit(1)
	}

	channelclient, err := channel.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to create channel [%s]:", *channelName, err)
	}

	dbInstance := db.CreateDBConf(globalConfig.DB.Host, globalConfig.DB.Port, globalConfig.DB.Dbuser, globalConfig.DB.Dbsecret, globalConfig.DB.Dbname)
	var fabex *Fabex
	if *task != "initdb" {
		err = dbInstance.Connect()
		if err != nil {
			log.Fatalln("DB connection failed:", err.Error())
		}
		log.Println("Connected to Postgres successfully")
		fabex = &Fabex{dbInstance, channelclient, ledgerClient}
	} else {
		fabex = &Fabex{dbInstance, channelclient, ledgerClient}
	}
	switch *task {
	case "initdb":
		err = fabex.db.Init()
		if err != nil {
			fmt.Printf("Failed to create table: %s", err)
			return
		}
		log.Println("Database and table created successfully")
	case "invoke":
		helpers.InvokeCC(fabex.channelClient, "a", "b", "30")
	case "query":
		helpers.QueryCC(fabex.channelClient, []byte("b"), *cc)
	case "channelinfo":
		resp, err := helpers.QueryChannelInfo(fabex.ledgerClient)
		if err != nil {
			log.Fatalf("Can't query blockchain info: %s", err)
		}
		fmt.Println("BlockChainInfo:", resp.BCI)
		fmt.Println("Endorser:", resp.Endorser)
		fmt.Println("Status:", resp.Status)
	case "channelconfig":
		helpers.QueryChannelConfig(fabex.ledgerClient)

	case "getblock":
		customBlock, err := blockfetcher.GetBlock(fabex.ledgerClient, *blocknum)
		if err != nil {
			break
		}

		if customBlock != nil {
			for _, block := range customBlock.Txs {
				fmt.Printf("\nBlock number: %d\nBlock hash: %s\na=%d\nb=%d\n", block.Blocknum, block.Hash, block.A, block.B)
			}
		}

	case "explore":

		// check we have up-to-date db or not
		// get last block hash
		resp, err := helpers.QueryChannelInfo(fabex.ledgerClient)
		if err != nil {
			log.Fatalf("Can't query blockchain info: %s", err)
		}
		currentHash := hex.EncodeToString(resp.BCI.CurrentBlockHash)

		//find txs from this block in db
		tx, err := fabex.db.QueryBlockByHash(currentHash)
		if err != nil {
			log.Printf("Can't find data with hash %s: %s", currentHash, err)
		}

		// update db if block with current hash not finded
		if tx == nil {
			log.Println("Explore new blocks")
			// find latest block in db
			txs, err := fabex.db.QueryAll()
			if err != nil {
				log.Fatalf("Can't query data: ", err)
			}
			var max uint64 = txs[0].Blocknum
			for _, tx := range txs {
				if tx.Blocknum > max {
					max = tx.Blocknum
				}
			}

			// set blocks counter to latest saved in db block number value
			var blockCounter uint64 = max

			// insert missing blocks/txs into db
			for {
				customBlock, err := blockfetcher.GetBlock(fabex.ledgerClient, blockCounter)
				if err != nil {
					break
				}

				if customBlock != nil {
					for _, block := range customBlock.Txs {
						fabex.db.Insert(block.Txid, block.Hash, block.Blocknum, block.A, block.B)
					}
				}
				blockCounter++
			}
			fmt.Println("All blocks saved")
		}
	case "getall":
		txs, err := fabex.db.QueryAll()
		if err != nil {
			log.Fatalf("Can't query data: ", err)
		}
		for _, tx := range txs {
			fmt.Printf("\nBlock number: %d\nBlock hash: %s\nTx id: %s\na=%d\nb=%d\n", tx.Blocknum, tx.Blockhash, tx.Txid, tx.A, tx.B)
		}

	case "grpc":
		serv := NewFabexServer(*grpcAddr, *grpcPort, fabex)
		StartGrpcServ(serv)
	}
}

type fabexServer struct {
	Address string
	Port    string
	Conf    *Fabex
}

func NewFabexServer(addr string, port string, conf *Fabex) *fabexServer {
	return &fabexServer{addr, port, conf}
}

func (s *fabexServer) Explore(req *pb.Request, stream pb.Fabex_ExploreServer) error {
	log.Printf("Strat stream from %d block", req.Startblock)
	// set blocks counter to latest saved in db block number value
	var blockCounter uint64 = uint64(req.Startblock)

	// insert missing blocks/txs into db
	for blockCounter <= uint64(req.Endblock) {
		customBlock, err := blockfetcher.GetBlock(s.Conf.ledgerClient, blockCounter)
		if err != nil {
			break
		}

		if customBlock != nil {
			for _, block := range customBlock.Txs {
				stream.Send(&pb.Reply{Txid: block.Txid, Hash: block.Hash, Blocknum: int64(block.Blocknum), A: block.A, B: block.B})
			}
		}
		blockCounter++
	}
	log.Println("Stream closed")
	return nil
}

func StartGrpcServ(serv *fabexServer) {
	grpcServer := grpc.NewServer()
	pb.RegisterFabexServer(grpcServer, serv)

	l, err := net.Listen("tcp", fmt.Sprintf("%s:%s", serv.Address, serv.Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("\nListening on tcp://%s:%s", serv.Address, serv.Port)
	grpcServer.Serve(l)
}
