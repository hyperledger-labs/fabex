package main

import (
	"context"

	"github.com/hyperledger-labs/fabex/db"
	"github.com/hyperledger-labs/fabex/helpers"
	"github.com/hyperledger-labs/fabex/ledgerclient"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	fabctx "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/pkg/errors"
)

type Engine struct {
	db             db.Storage
	channelClient  *channel.Client
	ledgerClient   *ledgerclient.CustomLedgerClient
	channelContext fabctx.ChannelProvider
}

func engineCreator(sdk *fabsdk.FabricSDK, dbInstance db.Storage) func(ch, user, org string) (*Engine, error) {
	return func(ch, user, org string) (*Engine, error) {
		clientChannelContext := sdk.ChannelContext(ch, fabsdk.WithUser(user), fabsdk.WithOrg(org))
		ledgerClient, err := ledger.New(clientChannelContext)
		if err != nil {
			return nil, errors.WithStack(errors.Wrapf(err, "failed to create ledger client"))
		}

		channelclient, err := channel.New(clientChannelContext)
		if err != nil {
			return nil, errors.WithStack(errors.Wrapf(err, "failed to create channel cient"))
		}
		return &Engine{db: dbInstance, channelClient: channelclient, ledgerClient: &ledgerclient.CustomLedgerClient{Client: ledgerClient}, channelContext: clientChannelContext}, nil
	}
}

func (e *Engine) Run(ctx context.Context) error {
	ch, err := e.channelContext()
	if err != nil {
		return errors.WithStack(err)
	}

	if err := e.db.Init(ch.ChannelID()); err != nil {
		return err
	}

	return helpers.Explore(ctx, e.channelContext, e.db, e.ledgerClient)
}
