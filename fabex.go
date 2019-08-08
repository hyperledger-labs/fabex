package main

import (
	"fabex/blockfetcher"
	"fabex/db"
	"fabex/helpers"
	"flag"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
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
	Table    string `yaml:table`
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
	c := config.FromFile("./connection-profile.yaml")
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

	dbInstance := db.CreateDBConf(globalConfig.DB.Host, globalConfig.DB.Port, globalConfig.DB.Dbuser, globalConfig.DB.Dbsecret, globalConfig.DB.Table)
	err = dbInstance.Connect()
	if err != nil {
		log.Fatalln("DB connection failed:", err.Error())
	}
	log.Println("Connected to Postgres successfully")
	fabex := &Fabex{dbInstance, channelclient, ledgerClient}

	switch *task {
	case "initdb":
		err = fabex.db.Init()
		if err != nil {
			fmt.Printf("Failed to create table: %s", err)
			return
		}
		log.Println("Table created successfully")
	case "invoke":
		helpers.InvokeCC(fabex.channelClient, "a", "b", "30")
	case "query":
		helpers.QueryCC(fabex.channelClient, []byte("b"), *cc)
	case "channelinfo":
		helpers.QueryChannelInfo(fabex.ledgerClient)
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
		var blockCounter uint64 = 0

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
	case "getall":
		txs, err := fabex.db.QueryAll("txs")
		if err != nil {
			log.Fatalf("Can't query data: ", err)
		}
		for _, tx := range txs {
			fmt.Printf("\nBlock number: %d\nBlock hash: %s\nTx id: %s\na=%d\nb=%d\n", tx.Blocknum, tx.Blockhash, tx.Txid, tx.A, tx.B)
		}
	}
}
