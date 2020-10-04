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

// Package blockhandler provides functionality for fetching blocks from blockchain
package blockhandler

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger-labs/fabex/db"
	"github.com/hyperledger-labs/fabex/models"
	fabcommon "github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-protos-go/ledger/rwset"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/protoutil"
	"github.com/pkg/errors"
)

// LedgerClient interface used for dependency injection of Fabric ledger client
type LedgerClient interface {
	QueryBlock(blockNumber uint64, options ...ledger.RequestOption) (*fabcommon.Block, error)
}

// CustomBlock stores slice of transactions (with block data)
type CustomBlock struct {
	Txs []db.Tx
}

// GetBlock gets information about specified block with blocknum number
func HandleBlock(block *fabcommon.Block) (*CustomBlock, error) {
	customBlock := &CustomBlock{}

	// get block hash
	hash := hex.EncodeToString(block.Header.DataHash)

	// get hash of the previous block
	previoushash := hex.EncodeToString(block.Header.PreviousHash)

	rawdata := block.GetData()
	for _, value := range rawdata.Data {

		// get validation code (0 is valid)
		processedtx := &peer.ProcessedTransaction{}
		err := proto.Unmarshal(value, processedtx)
		if err != nil {
			return nil, err
		}
		validationCode := processedtx.GetValidationCode()

		envelope, err := protoutil.GetEnvelopeFromBlock(value)
		if err != nil {
			return nil, err
		}

		// get ChannelHeader
		channelHeader, err := protoutil.ChannelHeader(envelope)
		if err != nil {
			return nil, err
		}

		// get timestamp
		txtime, err := ptypes.Timestamp(channelHeader.Timestamp)
		if err != nil {
			return nil, err
		}

		// get RW sets
		action, _ := protoutil.GetActionFromEnvelopeMsg(envelope)
		actionResults := action.GetResults()

		ReadWriteSet := &rwset.TxReadWriteSet{}

		err = proto.Unmarshal(actionResults, ReadWriteSet)
		if err != nil {
			return nil, err
		}

		txRWSet, err := rwsetutil.TxRwSetFromProtoMsg(ReadWriteSet)
		if err != nil {
			//fmt.Printf("Failed to convert rwset.TxReadWriteSet to rwsetutil.TxRWSet: %s", err)
			return nil, err
		}

		//get tx id
		bytesEnvelope, err := protoutil.GetBytesEnvelope(envelope)
		if err != nil {
			//fmt.Printf("Can't convert common.Envelope to bytes: ", err)
			return nil, err
		}
		TxId, err := protoutil.GetOrComputeTxIDFromEnvelope(bytesEnvelope)
		if err != nil {
			return nil, err
		}

		if protoutil.IsConfigBlock(block) {
			// cast "github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/common".Block to
			// "github.com/hyperledger/fabric/fabric-protos-go/common".Block
			configEnvelope, blockType, err := ConfigEnvelopeFromBlock(block)
			if err != nil {
				return nil, err
			}

			var stringedPayload []models.Chaincode
			switch blockType {
			case "Config":
				configPayload, err := protoutil.UnmarshalPayload(configEnvelope.Payload)
				if err != nil {
					return nil, errors.Wrap(err, "failed to unmarshal config payload")
				}

				configEnv := &fabcommon.ConfigEnvelope{}
				err = proto.Unmarshal(configPayload.Data, configEnv)
				if err != nil {
					return nil, errors.Wrap(err, "failed to unmarshal config envelope")
				}

				config := configEnv.GetConfig()

				configGroup := config.GetChannelGroup()

				groups, err := json.Marshal(configGroup.Groups)
				if err != nil {
					return nil, errors.Wrap(err, "failed to marshal config groups")
				}

				values, err := json.Marshal(configGroup.Values)
				if err != nil {
					return nil, errors.Wrap(err, "failed to marshal config values")
				}

				policies, err := json.Marshal(configGroup.Policies)
				if err != nil {
					return nil, errors.Wrap(err, "failed to marshal config policies")
				}

				modpolicy, err := json.Marshal(configGroup.ModPolicy)
				if err != nil {
					return nil, errors.Wrap(err, "failed to marshal config ModPolicy")
				}

				stringedPayload = append(stringedPayload, models.Chaincode{Key: "Type", Value: blockType})
				stringedPayload = append(stringedPayload, models.Chaincode{Key: "Sequence", Value: fmt.Sprint(config.GetSequence())})
				stringedPayload = append(stringedPayload, models.Chaincode{Key: "Version", Value: fmt.Sprint(configGroup.Version)})
				stringedPayload = append(stringedPayload, models.Chaincode{Key: "Groups", Value: string(groups)})
				stringedPayload = append(stringedPayload, models.Chaincode{Key: "Values", Value: string(values)})
				stringedPayload = append(stringedPayload, models.Chaincode{Key: "Policies", Value: string(policies)})
				stringedPayload = append(stringedPayload, models.Chaincode{Key: "ModPolicy", Value: string(modpolicy)})

			// get config update
			case "ConfigUpdate":
				configUpdateEnvelope, err := protoutil.EnvelopeToConfigUpdate(configEnvelope)
				if err != nil {
					return nil, errors.Wrap(err, "failed read config update")
				}
				fmt.Println(configUpdateEnvelope)
				configUpdateBytes := configUpdateEnvelope.GetConfigUpdate()
				var configUpdate = &fabcommon.ConfigUpdate{}
				err = proto.Unmarshal(configUpdateBytes, configUpdate)
				if err != nil {
					return nil, errors.Wrap(err, "failed to unmarshal config update")
				}

				// extract config update data
				ReadSet, err := json.Marshal(configUpdate.GetReadSet())
				if err != nil {
					return nil, err
				}
				WriteSet, err := json.Marshal(configUpdate.GetWriteSet())
				if err != nil {
					return nil, err
				}

				stringedPayload = append(stringedPayload, models.Chaincode{Key: "ChannelId", Value: configUpdate.GetChannelId()})
				stringedPayload = append(stringedPayload, models.Chaincode{Key: "ReadSet", Value: string(ReadSet)})
				stringedPayload = append(stringedPayload, models.Chaincode{Key: "WriteSet", Value: string(WriteSet)})

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
				block.Header.Number,
				string(jsonPayload),
				validationCode,
				txtime.Unix(),
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
					block.Header.Number,
					string(jsonPayload),
					validationCode,
					txtime.Unix(),
				}
				customBlock.Txs = append(customBlock.Txs, tx)
			}
		}
	}

	return customBlock, nil
}

// ConfigEnvelopeFromBlock extracts configuration envelope from the block based on the
// config type, i.e. HeaderType_ORDERER_TRANSACTION or HeaderType_CONFIG
func ConfigEnvelopeFromBlock(block *fabcommon.Block) (*fabcommon.Envelope, string, error) {
	if block == nil {
		return nil, "", errors.New("nil block")
	}

	envelope, err := protoutil.ExtractEnvelope(block, 0)
	if err != nil {
		return nil, "", errors.Wrapf(err, "failed to extract envelope from the block")
	}

	channelHeader, err := protoutil.ChannelHeader(envelope)
	if err != nil {
		return nil, "", errors.Wrap(err, "cannot extract channel header")
	}

	switch channelHeader.Type {
	case int32(fabcommon.HeaderType_ORDERER_TRANSACTION):
		payload, err := protoutil.UnmarshalPayload(envelope.Payload)
		if err != nil {
			return nil, "OrdererTx", errors.Wrap(err, "failed to unmarshal envelope to extract config payload for orderer transaction")
		}
		configEnvelop, err := protoutil.UnmarshalEnvelope(payload.Data)
		if err != nil {
			return nil, "OrdererTx", errors.Wrap(err, "failed to unmarshal config envelope for orderer type transaction")
		}

		return configEnvelop, "OrdererTx", nil
	case int32(fabcommon.HeaderType_CONFIG):
		return envelope, "Config", nil
	case int32(fabcommon.HeaderType_CONFIG_UPDATE):
		return envelope, "ConfigUpdate", nil
	default:
		return nil, "", errors.Errorf("unexpected header type: %v", channelHeader.Type)
	}
}
