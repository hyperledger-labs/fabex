package ledgerclient

import (
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"unsafe"
)

type CustomLedgerClient struct {
	Client *ledger.Client
}

func (clc *CustomLedgerClient) QueryBlock(blockNumber uint64, options ...ledger.RequestOption) (*common.Block, error) {
	block, err := clc.Client.QueryBlock(blockNumber, options...)
	fabricBlock := (*common.Block)(unsafe.Pointer(block))
	return fabricBlock, err
}
