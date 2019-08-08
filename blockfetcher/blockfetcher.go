package blockfetcher

import (
	"encoding/hex"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/protos/ledger/rwset"
	"github.com/hyperledger/fabric/protos/peer"
	"github.com/hyperledger/fabric/protos/utils"
)

type CustomBlock struct {
	Txs []Tx
}

type Tx struct {
	Hash     string
	Blocknum uint64
	Txid     string
	A        int64
	B        int64
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
				return nil, nil
			}

			action, err := utils.GetActionFromEnvelopeMsg(envelope)
			if err != nil {
				//fmt.Printf("Can't extract cc action in peer.ChaincodeAction format: ", err)
				return nil, nil
			}

			actionResults := action.GetResults()

			ReadWriteSet := &rwset.TxReadWriteSet{}

			err = proto.Unmarshal(actionResults, ReadWriteSet)
			if err != nil {
				//fmt.Printf("Failed to unmarshal: %s", err)
				return nil, nil
			}

			txRWSet, err := rwsetutil.TxRwSetFromProtoMsg(ReadWriteSet)
			if err != nil {
				//fmt.Printf("Failed to convert rwset.TxReadWriteSet to rwsetutil.TxRWSet: %s", err)
				return nil, nil
			}

			//get tx id
			bytesEnvelope, err := utils.GetBytesEnvelope(envelope)
			if err != nil {
				//fmt.Printf("Can't convert common.Envelope to bytes: ", err)
				return nil, nil
			}
			bytesTxId, err := utils.GetOrComputeTxIDFromEnvelope(bytesEnvelope)
			if err != nil {
				//fmt.Printf("Can't extract tx id in bytes format: ", err)
				return nil, nil
			}

			for _, nsRwSet := range txRWSet.NsRwSets {
				//fmt.Println(nsRwSet.KvRwSet.Writes)
				if len(nsRwSet.KvRwSet.Writes) != 0 && (nsRwSet.KvRwSet.Writes[0].Key == "a" || nsRwSet.KvRwSet.Writes[0].Key == "b") {
					//fmt.Println(nsRwSet.KvRwSet.Writes)
					tx := Tx{
						hash,
						blocknum,
						string(bytesTxId),
						1,
						1,
					}
					customBlock.Txs = append(customBlock.Txs, tx)
				}
				//fmt.Println(nsRwSet.KvRwSet.Reads)
				//fmt.Println(nsRwSet.KvRwSet.Writes)
			}
		}
	}
	return customBlock, nil
}
