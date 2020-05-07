package blockfetcher

import (
	"github.com/stretchr/testify/assert"
	"github.com/hyperledger-labs/fabex/db"
	"github.com/hyperledger-labs/fabex/ledgerclient"
	"testing"
)

var fakeTxs = &CustomBlock{
	Txs: []db.Tx{
		db.Tx{Blocknum: 1},
	},
}

func TestGetBlock(t *testing.T) {
	fakeLedgerClient := new(ledgerclient.FakeLedgerClient)
	block, err := GetBlock(fakeLedgerClient, 1)
	assert.Equal(t, nil, err, "GetBlock err not nil")
	assert.Greater(t, len(block.Txs), 0, "GetBlock result empty")
}
