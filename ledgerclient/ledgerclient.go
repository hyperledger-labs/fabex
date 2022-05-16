package ledgerclient

import (
	"unsafe"

	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/pkg/errors"
)

type CustomLedgerClient struct {
	Client *ledger.Client
}

func (clc *CustomLedgerClient) QueryBlock(blockNumber uint64, options ...ledger.RequestOption) (*common.Block, error) {
	block, err := clc.Client.QueryBlock(blockNumber, options...)
	fabricBlock := (*common.Block)(unsafe.Pointer(block))
	return fabricBlock, err
}

func (clc *CustomLedgerClient) QueryInfo(options ...ledger.RequestOption) (*fab.BlockchainInfoResponse, error) {
	resp, err := clc.Client.QueryInfo(options...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp, nil
}
