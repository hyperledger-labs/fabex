package ledgerclient

import (
	"io/ioutil"
	"unsafe"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric/protoutil"
)

type FakeLedgerClient struct {
}

func FakeBlock() *common.Block {
	blockBytes, err := ioutil.ReadFile("../tests/custom.block")
	if err != nil {
		panic(err)
	}
	block, err := protoutil.UnmarshalBlock(blockBytes)
	if err != nil {
		panic(err)
	}

	fabricBlock := (*common.Block)(unsafe.Pointer(block))

	return fabricBlock
}

func (_ *FakeLedgerClient) QueryBlock(_ uint64, _ ...ledger.RequestOption) (*common.Block, error) {
	block := FakeBlock()
	return block, nil
}

func (_ *FakeLedgerClient) QueryInfo(_ ...ledger.RequestOption) (*fab.BlockchainInfoResponse, error) {
	return &fab.BlockchainInfoResponse{
		BCI: &common.BlockchainInfo{
			Height:           1,
			CurrentBlockHash: []byte("000"),
		},
	}, nil
}
