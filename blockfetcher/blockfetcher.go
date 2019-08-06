package blockfetcher

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	pbrwset "github.com/hyperledger/fabric/protos/ledger/rwset"
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

		rw := &pbrwset.NsReadWriteSet{}
		err = proto.Unmarshal(actionResults, rw)
		if err != nil {
			fmt.Printf("Failed to unmarshal rwset: %s", err)
		}

		fmt.Println("RWSet unmarshaled:", string(rw.Rwset))

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
