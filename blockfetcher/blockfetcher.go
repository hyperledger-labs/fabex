package blockfetcher

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/hyperledger/fabric/protos/ledger/rwset"
	"github.com/hyperledger/fabric/protos/utils"
)

func GetBlock(ledgerClient *ledger.Client, blocknum uint64) {
	block, err := ledgerClient.QueryBlock(blocknum)
	if err != nil {
		fmt.Printf("Failed to query block %d: %s", blocknum, err)
	}
	rawdata := block.GetData()
	for _, value := range rawdata.Data {

		envelope, err := utils.GetEnvelopeFromBlock(value)
		if err != nil {
			fmt.Printf("Can't extract envelope: ", err)
			return
		}

		action, err := utils.GetActionFromEnvelopeMsg(envelope)
		if err != nil {
			fmt.Printf("Can't extract cc action in peer.ChaincodeAction format: ", err)
			return
		}

		actionResults := action.GetResults()

		ReadWriteSet := &rwset.TxReadWriteSet{}

		err = proto.Unmarshal(actionResults, ReadWriteSet)
		if err != nil {
			fmt.Printf("Failed to unmarshal: %s", err)
		}

		txRWSet, err := rwsetutil.TxRwSetFromProtoMsg(ReadWriteSet)
		if err != nil {
			fmt.Printf("Failed to convert rwset.TxReadWriteSet to rwsetutil.TxRWSet: %s", err)
		}

		for _, nsRwSet := range txRWSet.NsRwSets {
			fmt.Println(nsRwSet.KvRwSet.Reads)
			fmt.Println(nsRwSet.KvRwSet.Writes)
		}

		// get tx id
		//bytesEnvelope, err := utils.GetBytesEnvelope(envelope)
		//if err != nil {
		//	fmt.Printf("Can't convert common.Envelope to bytes: ", err)
		//	return
		//}
		//bytesTxId, err := utils.GetOrComputeTxIDFromEnvelope(bytesEnvelope)
		//if err != nil {
		//	fmt.Printf("Can't extract tx id in bytes format: ", err)
		//	return
		//}
		//fmt.Println("Tx id:", string(bytesTxId))
	}
}
