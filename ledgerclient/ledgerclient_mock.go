package ledgerclient

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric/protos/common"
	"github.com/hyperledger/fabric/protos/utils"
	"io/ioutil"
	"unsafe"
)

type FakeLedgerClient struct {
}

func FakeBlock() *common.Block {
	blockBytes, err := ioutil.ReadFile("../tests/custom.block")
	if err != nil {
		panic(err)
	}
	block, err := utils.UnmarshalBlock(blockBytes)
	if err != nil {
		panic(err)
	}

	fabricBlock := (*common.Block)(unsafe.Pointer(block))

	return fabricBlock
}

func (m *FakeLedgerClient) QueryBlock(blockNumber uint64, options ...ledger.RequestOption) (*common.Block, error) {
	block := FakeBlock()
	return block, nil
}
