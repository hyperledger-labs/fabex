package helpers

import (
	"encoding/hex"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/vadiminshakov/fabex/blockfetcher"
	"github.com/vadiminshakov/fabex/models"
	"log"
	"sync"
)

func Explore(wg *sync.WaitGroup, fab *models.Fabex) error {
	// check we have up-to-date db or not
	// get last block hash
	resp, err := QueryChannelInfo(fab.LedgerClient)
	if err != nil {
		return err
	}
	currentHash := hex.EncodeToString(resp.BCI.CurrentBlockHash)

	//find txs from this block in db
	tx, err := fab.Db.QueryBlockByHash(currentHash)

	if err != nil {
		if err.Error() != "sql: no rows in result set" && err.Error() != "mongo: no documents in result" {
			return err
		}
	}

	// update db if block with current hash not finded
	var blockCounter uint64
	if tx == nil {
		log.Println("Explore new blocks")
		// find latest block in db
		txs, err := fab.Db.QueryAll()

		if len(txs) != 0 {
			if err != nil {
				return err
			}
			var max uint64 = txs[0].Blocknum
			for _, tx := range txs {
				if tx.Blocknum > max {
					max = tx.Blocknum
				}
			}

			// set blocks counter to latest saved in db block number value
			blockCounter = max
		} else {
			blockCounter = 0
		}

		// insert missing blocks/txs into db
		for {
			customBlock, err := blockfetcher.GetBlock(fab.LedgerClient, blockCounter)
			if err != nil {
				break
			}

			if customBlock != nil {
				for _, block := range customBlock.Txs {
					//log.Printf("\nBlock finded\nBlock number: %d\nBlock hash: %s\nTx id: %s\nPayload:=%s\n", block.Blocknum, block.Hash, block.Txid, block.Payload)
					fab.Db.Insert(block.Txid, block.Hash, block.Blocknum, block.Payload)
				}
			}
			blockCounter++
		}
	}
	wg.Done()
	return nil
}

func EnrollUser(sdk *fabsdk.FabricSDK, user, secret string) {
	ctx := sdk.Context()
	mspClient, err := msp.New(ctx)
	if err != nil {
		fmt.Printf("Failed to create msp client: %s\n", err)
	}

	_, err = mspClient.GetSigningIdentity(user)
	if err == msp.ErrUserNotFound {
		fmt.Println("Going to enroll user")
		err = mspClient.Enroll(user, msp.WithSecret(secret))

		if err != nil {
			fmt.Printf("Failed to enroll user: %s\n", err)
		} else {
			fmt.Printf("Success enroll user: %s\n", user)
		}

	} else if err != nil {
		fmt.Printf("Failed to get user: %s\n", err)
	} else {
		fmt.Printf("User %s already enrolled, skip enrollment.\n", user)
	}
}

func QueryChannelConfig(ledgerClient *ledger.Client) {
	resp1, err := ledgerClient.QueryConfig()
	if err != nil {
		fmt.Printf("Failed to queryConfig: %s", err)
	}
	fmt.Println("ChannelID: ", resp1.ID())
	fmt.Println("Channel Orderers: ", resp1.Orderers())
	fmt.Println("Channel Versions: ", resp1.Versions())
}

func QueryChannelInfo(ledgerClient *ledger.Client) (*fab.BlockchainInfoResponse, error) {
	resp, err := ledgerClient.QueryInfo()
	if err != nil {
		fmt.Printf("Failed to queryInfo: %s", err)
		return nil, err
	}
	return resp, nil
}

func SetupLogLevel(lvl logging.Level) {
	logging.SetLevel("fabsdk", lvl)
	logging.SetLevel("fabsdk/common", lvl)
	logging.SetLevel("fabsdk/fab", lvl)
	logging.SetLevel("fabsdk/client", lvl)
}
