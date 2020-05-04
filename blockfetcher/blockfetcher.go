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
	"github.com/hyperledger/fabric/common/configtx"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/ledger/rwset"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/pkg/errors"
	"github.com/vadiminshakov/fabex/db"
	"github.com/vadiminshakov/fabex/models"
	"strings"
	"unsafe"
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

		if len(txRWSet.NsRwSets) == 0 {
			// cast "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common".Block to
			// "github.com/hyperledger/fabric/protos/common".Block
			var fabricBlock = (*common.Block)(unsafe.Pointer(block))
			configEnvelope, err := ConfigEnvelopeFromBlock(fabricBlock)
			if err != nil {
				return nil, err
			}

			payload, err := utils.ExtractPayload(configEnvelope)
			if err != nil {
				return nil, errors.Wrap(err, "failed to extract payload from config envelope")
			}

			// get config update
			configUpdate, err := configtx.UnmarshalConfigUpdateFromPayload(payload)
			if err != nil {
				return nil, errors.Wrap(err, "could not read config update")
			}

			var stringedPayload []models.Chaincode
			ReadSet, err := json.Marshal(configUpdate.ReadSet)
			if err != nil {
				return nil, err
			}
			WriteSet, err := json.Marshal(configUpdate.WriteSet)
			if err != nil {
				return nil, err
			}
			stringedPayload = append(stringedPayload, models.Chaincode{Key: "ChannelId", Value: configUpdate.ChannelId})
			stringedPayload = append(stringedPayload, models.Chaincode{Key: "ReadSet", Value: string(ReadSet)})
			stringedPayload = append(stringedPayload, models.Chaincode{Key: "WriteSet", Value: string(WriteSet)})

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

// ConfigEnvelopeFromBlock extracts configuration envelope from the block based on the
// config type, i.e. HeaderType_ORDERER_TRANSACTION or HeaderType_CONFIG
func ConfigEnvelopeFromBlock(block *common.Block) (*common.Envelope, error) {
	if block == nil {
		return nil, errors.New("nil block")
	}

	envelope, err := utils.ExtractEnvelope(block, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to extract envelope from the block")
	}

	channelHeader, err := utils.ChannelHeader(envelope)
	if err != nil {
		return nil, errors.Wrap(err, "cannot extract channel header")
	}

	switch channelHeader.Type {
	case int32(common.HeaderType_ORDERER_TRANSACTION):
		payload, err := utils.UnmarshalPayload(envelope.Payload)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal envelope to extract config payload for orderer transaction")
		}
		configEnvelop, err := utils.UnmarshalEnvelope(payload.Data)
		if err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal config envelope for orderer type transaction")
		}

		return configEnvelop, nil
	case int32(common.HeaderType_CONFIG):
		return envelope, nil
	default:
		return nil, errors.Errorf("unexpected header type: %v", channelHeader.Type)
	}
}
