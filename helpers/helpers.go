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

package helpers

import (
	"encoding/hex"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/hyperledger-labs/fabex/blockfetcher"
	"github.com/hyperledger-labs/fabex/models"
)

const NOT_FOUND_ERR = "not found"

func Explore(fab *models.Fabex) error {
	// check we have up-to-date db or not
	// get last block hash
	resp, err := QueryChannelInfo(fab.LedgerClient.Client)
	if err != nil {
		return err
	}
	currentHash := hex.EncodeToString(resp.BCI.CurrentBlockHash)

	//find txs from this block in db
	txs, err := fab.Db.QueryBlockByHash(currentHash)
	if err != nil {
		if err.Error() != "sql: no rows in result set" && err.Error() != "mongo: no documents in result" {
			return err
		}
	}

	// update db if block with current hash not finded
	var blockCounter uint64

	if txs == nil {

		// find latest tx in db
		lastTx, err := fab.Db.GetLastEntry()
		if err != nil && err.Error() != NOT_FOUND_ERR {
			return errors.Wrap(err, "Can't to get last block")
		}

		if err != nil && err.Error() == NOT_FOUND_ERR {
			lastTx.Blocknum = 0
		}

		// set blocks counter to latest saved in db block number value
		blockCounter = lastTx.Blocknum

		// insert missing blocks/txs into db
		for {
			// increment for fetching next (after last added to DB) block from ledger
			blockCounter++

			customBlock, err := blockfetcher.GetBlock(fab.LedgerClient, blockCounter)
			if err != nil {
				return errors.Wrap(err, "GetBlock error")
			}

			if customBlock != nil {
				for _, tx := range customBlock.Txs {
					err = fab.Db.Insert(tx)
					if err != nil {
						return err
					}
				}
			} else {
				break
			}
		}
	}
	return nil
}

func EnrollUser(sdk *fabsdk.FabricSDK, user, secret string) error {
	ctx := sdk.Context()
	mspClient, err := msp.New(ctx)
	if err != nil {
		return errors.Wrap(err, "Failed to create msp client")
	}

	_, err = mspClient.GetSigningIdentity(user)
	if err == msp.ErrUserNotFound {
		log.Println("Going to enroll user")
		err = mspClient.Enroll(user, msp.WithSecret(secret))

		if err != nil {
			return errors.Wrap(err, "Failed to enroll user")
		}
		log.Printf("Success enroll user: %s\n", user)
	} else if err != nil {
		return errors.Wrap(err, "Failed to get user")
	}

	log.Printf("User %s already enrolled, skip enrollment.\n", user)
	return nil
}

func QueryChannelConfig(ledgerClient *ledger.Client) error {
	resp, err := ledgerClient.QueryConfig()
	if err != nil {
		return errors.Wrap(err, "Failed to queryConfig")
	}
	log.Printf("ChannelID: %v\nChannel Orderers: %v\nChannel Versions: %v\n", resp.ID(), resp.Orderers(), resp.Versions())

	return nil
}

func QueryChannelInfo(ledgerClient *ledger.Client) (*fab.BlockchainInfoResponse, error) {
	resp, err := ledgerClient.QueryInfo()
	return resp, err
}

func SetupLogLevel(lvl logging.Level) {
	logging.SetLevel("fabsdk", lvl)
	logging.SetLevel("fabsdk/common", lvl)
	logging.SetLevel("fabsdk/fab", lvl)
	logging.SetLevel("fabsdk/client", lvl)
}
