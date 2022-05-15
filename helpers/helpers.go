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
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"unicode/utf8"

	"go.uber.org/zap"

	fabctx "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"

	"github.com/hyperledger-labs/fabex/blockhandler"
	"github.com/hyperledger-labs/fabex/db"
	"github.com/hyperledger-labs/fabex/models"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/events/deliverclient/seek"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const NOT_FOUND_ERR = "not found"

func Explore(ctx context.Context, chprovider fabctx.ChannelProvider, database db.Storage, lClient blockhandler.LedgerClient) error {
	log, ok := ctx.Value("log").(*zap.Logger)
	if !ok {
		return errors.WithStack(errors.New("failed to get logger from context"))
	}

	// check we have up-to-date db or not
	// get last block hash
	resp, err := QueryChannelInfo(lClient)
	if err != nil {
		return err
	}
	currentHash := hex.EncodeToString(resp.BCI.CurrentBlockHash)

	//find txs from this block in db
	chclient, err := chprovider()
	if err != nil {
		return errors.WithStack(err)
	}

	txs, err := database.QueryBlockByHash(chclient.ChannelID(), currentHash)
	if err != nil {
		if err.Error() != "sql: no rows in result set" && err.Error() != "mongo: no documents in result" {
			return err
		}
	}

	if txs == nil {
		// find latest tx in db
		lastTx, err := database.GetLastEntry(chclient.ChannelID())
		if err != nil && err.Error() != NOT_FOUND_ERR {
			return errors.Wrap(err, "Can't to get last block")
		}

		// set blocks listener from latest saved in db blockchain height+1
		var blockNumber uint64
		if lastTx.Hash != "" {
			blockNumber = lastTx.Blocknum + 1
		}
		eventClient, err := event.New(
			chprovider,
			event.WithBlockEvents(),
			event.WithSeekType(seek.FromBlock),
			event.WithBlockNum(blockNumber), // increment for fetching next (after last added to DB) block from ledger
			event.WithEventConsumerTimeout(0),
		)
		if err != nil {
			return errors.WithStack(errors.Wrap(err, "event service error"))
		}
		reg, notifier, err := eventClient.RegisterBlockEvent()
		if err != nil {
			return errors.WithStack(errors.Wrap(err, "event service registration error"))
		}
		defer func() {
			go func() {
				for range notifier {
				}
			}()
			eventClient.Unregister(reg)
		}()

		// insert missing blocks/txs into db
		for ctx.Err() == nil {
			blockEvent, ok := <-notifier
			if !ok {
				break
			}

			customBlock, err := blockhandler.HandleBlock(blockEvent.Block)
			if err != nil {
				return errors.Wrap(err, "GetBlock error")
			}

			if customBlock == nil {
				break
			}

			for _, tx := range customBlock.Txs {
				err = database.Insert(chclient.ChannelID(), tx)
				if err != nil {
					return err
				}
				log.Debug("add tx", zap.String("channel", chclient.ChannelID()), zap.Uint64("block number", blockEvent.Block.Header.Number), zap.String("tx ID", tx.Txid))
			}
		}
		log.Info("stop expoler", zap.String("channel", chclient.ChannelID()))
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

func QueryChannelInfo(ledgerClient blockhandler.LedgerClient) (*fab.BlockchainInfoResponse, error) {
	resp, err := ledgerClient.QueryInfo()
	return resp, errors.WithStack(err)
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
