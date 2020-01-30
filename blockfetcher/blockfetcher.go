package blockfetcher

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/protos/ledger/rwset"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
	"github.com/vadiminshakov/fabex/db"
)

type CustomBlock struct {
	Txs []db.Tx
}

func GetBlock(ledgerClient *ledger.Client, blocknum uint64) (*CustomBlock, error) {
	customBlock := &CustomBlock{}

	block, err := ledgerClient.QueryBlock(blocknum)
	if err != nil {
		//fmt.Printf("Failed to query block %d: %s", blocknum, err)
		return nil, err
	}

	// get block hash
	hash := hex.EncodeToString(block.Header.DataHash)

	rawdata := block.GetData()
	for _, value := range rawdata.Data {

		// get validation code (0 is valid)
		processedtx := &peer.ProcessedTransaction{}
		proto.Unmarshal(value, processedtx)
		if err != nil {
			fmt.Printf("Failed to unmarshal: %s", err)
		}
		validationCode := processedtx.GetValidationCode()

		if validationCode == 0 {
			envelope, err := utils.GetEnvelopeFromBlock(value)
			if err != nil {
				//fmt.Printf("Can't extract envelope: ", err)
				return nil, err
			}

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
			bytesTxId, err := utils.GetOrComputeTxIDFromEnvelope(bytesEnvelope)
			if err != nil {
				return nil, err
			}

			for _, nsRwSet := range txRWSet.NsRwSets {

				// get only those txs that changes state
				if len(nsRwSet.KvRwSet.Writes) != 0 {

					var stringedPayload []Chaincode
					for _, write := range nsRwSet.KvRwSet.Writes {
						stringedPayload = append(stringedPayload, Chaincode{Key: write.Key, Value: string(write.Value)})
					}

					jsonPayload, err := json.Marshal(stringedPayload)
					if err != nil {
						return nil, err
					}
					tx := db.Tx{
						bytesTxId,
						hash,
						blocknum,
						string(jsonPayload),
					}
					customBlock.Txs = append(customBlock.Txs, tx)

				}
			}
		}
	}
	return customBlock, nil
}
