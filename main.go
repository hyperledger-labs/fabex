package main

import (
	"fabex/blockfetcher"
	"fabex/helpers"
	"flag"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"os"
)

var (
	cc          *string
	user        *string
	secret      *string
	channelName *string
	lvl         = logging.INFO
)

func main() {
	cc = flag.String("cc", "mycc", "chaincode name")
	user = flag.String("user", "admin", "user name")
	secret = flag.String("secret", "adminpw", "user secret")
	channelName = flag.String("channel", "mychannel", "channel name")
	init := flag.Bool("init", false, "enroll user (true) or not (false)")
	task := flag.String("task", "query", "choose the task to execute")
	blocknum := flag.Uint64("blocknum", 0, "block number")
	flag.Parse()

	fmt.Println("Reading connection profile..")
	c := config.FromFile("./connection-profile.yaml")
	sdk, err := fabsdk.New(c)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		os.Exit(1)
	}
	defer sdk.Close()

	helpers.SetupLogLevel(lvl)

	if *init {
		helpers.EnrollUser(sdk, *user, *secret)
	}

	clientChannelContext := sdk.ChannelContext(*channelName, fabsdk.WithUser(*user), fabsdk.WithOrg("Org1"))
	ledgerClient, err := ledger.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to create channel [%s] client: %#v", *channelName, err)
		os.Exit(1)
	}

	client, err := channel.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to create channel [%s]:", *channelName, err)
	}

	switch *task {
	case "invoke":
		helpers.InvokeCC(client, "a", "b", "30")
	case "query":
		helpers.QueryCC(client, []byte("b"), *cc)
	case "channelinfo":
		helpers.QueryChannelInfo(ledgerClient)
	case "channelconfig":
		helpers.QueryChannelConfig(ledgerClient)
	case "getblock":
		blockfetcher.GetBlock(ledgerClient, *blocknum)
	}

	fmt.Println("Done.")
}
