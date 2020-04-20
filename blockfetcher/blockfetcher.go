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

package blockfetcher

import (
	"encoding/hex"
	"encoding/json"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/protos/ledger/rwset"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/vadiminshakov/fabex/db"
	"github.com/vadiminshakov/fabex/models"
	"strings"
)

type CustomBlock struct {
	Txs []db.Tx
}

func GetBlock(ledgerClient *ledger.Client, blocknum uint64) (*CustomBlock, error) {
	customBlock := &CustomBlock{}

	block, err := ledgerClient.QueryBlock(blocknum)
	if err != nil {
		// skip if it's just the end of the blockchain
		if strings.Contains(err.Error(), "Entry not found in index") {
			return nil, nil
		}
		return nil, err
	}

	// get block hash
	hash := hex.EncodeToString(block.Header.DataHash)

	// get hash of the previous block
	previoushash := hex.EncodeToString(block.Header.PreviousHash)

	rawdata := block.GetData()
	for _, value := range rawdata.Data {

		// get validation code (0 is valid)
		processedtx := &peer.ProcessedTransaction{}
		proto.Unmarshal(value, processedtx)
		if err != nil {
			return nil, err
		}
		validationCode := processedtx.GetValidationCode()

		envelope, err := utils.GetEnvelopeFromBlock(value)
		if err != nil {
			return nil, err
		}

		// get ChannelHeader
		channelHeader, err := utils.ChannelHeader(envelope)
		if err != nil {
			return nil, err
		}

		// get timestamp
		timeInBlock, err := ptypes.Timestamp(channelHeader.Timestamp)
		if err != nil {
			return nil, err
		}

		// get RW sets
		action, _ := utils.GetActionFromEnvelopeMsg(envelope)
		actionResults := action.GetResults()

		ReadWriteSet := &rwset.TxReadWriteSet{}

		err = proto.Unmarshal(actionResults, ReadWriteSet)
		if err != nil {
			//fmt.Printf("Failed to unmarshal: %s", err)
			return nil, err
		}

		txRWSet, err := rwsetutil.TxRwSetFromProtoMsg(ReadWriteSet)
		if err != nil {
			//fmt.Printf("Failed to convert rwset.TxReadWriteSet to rwsetutil.TxRWSet: %s", err)
			return nil, err
		}

		//get tx id
		bytesEnvelope, err := utils.GetBytesEnvelope(envelope)
		if err != nil {
			//fmt.Printf("Can't convert common.Envelope to bytes: ", err)
			return nil, err
		}
		TxId, err := utils.GetOrComputeTxIDFromEnvelope(bytesEnvelope)
		if err != nil {
			return nil, err
		}

		for _, nsRwSet := range txRWSet.NsRwSets {

			// get only those txs that changes state
			if len(nsRwSet.KvRwSet.Writes) != 0 {

				var stringedPayload []models.Chaincode
				for _, write := range nsRwSet.KvRwSet.Writes {
					stringedPayload = append(stringedPayload, models.Chaincode{Key: write.Key, Value: string(write.Value)})
				}

				jsonPayload, err := json.Marshal(stringedPayload)
				if err != nil {
					return nil, err
				}
				tx := db.Tx{
					channelHeader.ChannelId,
					TxId,
					hash,
					previoushash,
					blocknum,
					string(jsonPayload),
					validationCode,
					timeInBlock.Unix(),
				}
				customBlock.Txs = append(customBlock.Txs, tx)

			}
		}
	}
	return customBlock, nil
}
