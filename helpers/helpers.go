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
	"encoding/json"
	"fmt"
	"github.com/hyperledger-labs/fabex/blockhandler"
	"github.com/hyperledger-labs/fabex/db"
	"github.com/hyperledger-labs/fabex/models"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"unicode/utf8"
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

	if txs == nil {
		// find latest tx in db
		lastTx, err := fab.Db.GetLastEntry()
		if err != nil && err.Error() != NOT_FOUND_ERR {
			return errors.Wrap(err, "Can't to get last block")
		}

		// set blocks listener from latest saved in db blockchain height+1
		var blockNumber uint64
		if lastTx.Hash != "" {
			blockNumber = lastTx.Blocknum + 1
		}
		eventClient, err := event.New(
			fab.ChannelContext,
			event.WithBlockEvents(),
			event.WithSeekType(seek.FromBlock),
			event.WithBlockNum(blockNumber), // increment for fetching next (after last added to DB) block from ledger
		)
		if err != nil {
			return errors.Wrap(err, "event service error")
		}
		_, notifier, err := eventClient.RegisterBlockEvent()
		if err != nil {
			return errors.Wrap(err, "event service registration error")
		}

		// insert missing blocks/txs into db
		for {
			blockEvent := <-notifier

			customBlock, err := blockhandler.HandleBlock(blockEvent.Block)
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

func QueryChannelConfig(ledgerClient *ledger.Client) (fab.ChannelCfg, error) {
	return ledgerClient.QueryConfig()
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

func PackTxsToBlocks(blocks []db.Tx) ([]models.Block, error) {
	var blockAlreadyRead = make(map[uint64]bool)

	var Blocks []models.Block
	for _, in := range blocks {
		var (
			block models.Block
			tx    models.Tx
		)

		if _, ok := blockAlreadyRead[in.Blocknum]; !ok {
			block = models.Block{ChannelId: in.ChannelId, Blocknum: in.Blocknum, BlockHash: in.Hash, PreviousHash: in.PreviousHash}
		}

		tx.Txid = in.Txid
		tx.ValidationCode = in.ValidationCode

		var ccData []models.WriteKV

		err := json.Unmarshal(in.Payload, &ccData)
		if err != nil {
			return nil, err
		}

		for _, item := range ccData {
			tx.KV = append(tx.KV, models.WriteKV{item.Key, item.Value})
		}

		block.Txs = append(block.Txs, tx)
		Blocks = append(Blocks, block)
		blockAlreadyRead[in.Blocknum] = true
	}

	return Blocks, nil
}

const (
	minUnicodeRuneValue   = 0            //U+0000
	maxUnicodeRuneValue   = utf8.MaxRune //U+10FFFF - maximum (and unallocated) code point
	compositeKeyNamespace = "\x00"
)

func CreateCompositeKey(objectType string, attributes []string) (string, error) {
	if err := validateCompositeKeyAttribute(objectType); err != nil {
		return "", err
	}
	ck := compositeKeyNamespace + objectType + string(minUnicodeRuneValue)
	for _, att := range attributes {
		if err := validateCompositeKeyAttribute(att); err != nil {
			return "", err
		}
		ck += att + string(minUnicodeRuneValue)
	}
	return ck, nil
}

func validateCompositeKeyAttribute(str string) error {
	if !utf8.ValidString(str) {
		return fmt.Errorf("not a valid utf8 string: [%x]", str)
	}
	for index, runeValue := range str {
		if runeValue == minUnicodeRuneValue || runeValue == maxUnicodeRuneValue {
			return fmt.Errorf(`input contains unicode %#U starting at position [%d]. %#U and %#U are not allowed in the input attribute of a composite key`,
				runeValue, index, minUnicodeRuneValue, maxUnicodeRuneValue)
		}
	}
	return nil
}
