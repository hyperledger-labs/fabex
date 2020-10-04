package blockhandler

import (
	"github.com/hyperledger/fabric/protoutil"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

func TestGetBlock(t *testing.T) {
	blockBytes, err := ioutil.ReadFile("../tests/custom.block")
	if err != nil {
		panic(err)
	}
	rawBlock, err := protoutil.UnmarshalBlock(blockBytes)
	if err != nil {
		panic(err)
	}
	block, err := HandleBlock(rawBlock)
	assert.Equal(t, nil, err, "GetBlock err not nil")
	assert.Greater(t, len(block.Txs), 0, "GetBlock result empty")
}
