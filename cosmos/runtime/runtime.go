// SPDX-License-Identifier: BUSL-1.1
//
// Copyright (C) 2023, Berachain Foundation. All rights reserved.
// Use of this software is govered by the Business Source License included
// in the LICENSE file of this repository and at www.mariadb.com/bsl11.
//
// ANY USE OF THE LICENSED WORK IN VIOLATION OF THIS LICENSE WILL AUTOMATICALLY
// TERMINATE YOUR RIGHTS UNDER THIS LICENSE FOR THE CURRENT AND ALL OTHER
// VERSIONS OF THE LICENSED WORK.
//
// THIS LICENSE DOES NOT GRANT YOU ANY RIGHT IN ANY TRADEMARK OR LOGO OF
// LICENSOR OR ITS AFFILIATES (PROVIDED THAT YOU MAY USE A TRADEMARK OR LOGO OF
// LICENSOR AS EXPRESSLY REQUIRED BY THIS LICENSE).
//
// TO THE EXTENT PERMITTED BY APPLICABLE LAW, THE LICENSED WORK IS PROVIDED ON
// AN “AS IS” BASIS. LICENSOR HEREBY DISCLAIMS ALL WARRANTIES AND CONDITIONS,
// EXPRESS OR IMPLIED, INCLUDING (WITHOUT LIMITATION) WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, NON-INFRINGEMENT, AND
// TITLE.

package runtime

import (
	"github.com/prysmaticlabs/prysm/v4/consensus-types/interfaces"

	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/client"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"

	"pkg.berachain.dev/polaris/beacon/eth"
	"pkg.berachain.dev/polaris/cosmos/config"
	antelib "pkg.berachain.dev/polaris/cosmos/lib/ante"
	libtx "pkg.berachain.dev/polaris/cosmos/lib/tx"
	"pkg.berachain.dev/polaris/cosmos/runtime/miner"
	"pkg.berachain.dev/polaris/cosmos/runtime/txpool"
	evmtypes "pkg.berachain.dev/polaris/cosmos/x/evm/types"
	coretypes "pkg.berachain.dev/polaris/eth/core/types"
)

// EVMKeeper is an interface that defines the methods needed for the EVM setup.
type EVMKeeper interface {
	// Setup initializes the EVM keeper.
	Setup(*eth.ExecutionClient) error
}

// CosmosApp is an interface that defines the methods needed for the Cosmos setup.
type CosmosApp interface {
	SetPrepareProposal(sdk.PrepareProposalHandler)
	SetMempool(mempool.Mempool)
	SetAnteHandler(sdk.AnteHandler)
}

// Polaris is a struct that wraps the Polaris struct from the polar package.
// It also includes wrapped versions of the Geth Miner and TxPool.
type Polaris struct {
	*eth.ExecutionClient

	// WrappedMiner is a wrapped version of the Miner component.
	WrappedMiner *miner.Miner
	// WrappedTxPool is a wrapped version of the Mempool component.
	WrappedTxPool *txpool.Mempool
	// logger is the underlying logger supplied by the sdk.
	logger log.Logger
}

// New creates a new Polaris runtime from the provided
// dependencies.
func New(
	appOpts servertypes.AppOptions,
	logger log.Logger,
) (*Polaris, error) {
	var err error
	p := &Polaris{
		logger: logger,
	}

	// Read the configuration from the cosmos app options
	cfg, err := config.ReadConfigFromAppOpts(appOpts)
	if err != nil {
		return nil, err
	}
	// Connect to the execution client.
	p.ExecutionClient, err = eth.NewRemoteExecutionClient(
		cfg.ExecutionClient.RPCDialURL, cfg.ExecutionClient.JWTSecretPath, logger,
	)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// New creates a new Polaris runtime from the provided
// dependencies, panics on error.
func MustNew(appOpts servertypes.AppOptions, logger log.Logger) *Polaris {
	p, err := New(appOpts, logger)
	if err != nil {
		panic(err)
	}
	return p
}

// Build is a function that sets up the Polaris struct.
// It takes a BaseApp and an EVMKeeper as arguments.
// It returns an error if the setup fails.
func (p *Polaris) Build(app CosmosApp) error {
	p.WrappedTxPool = txpool.New(p.TxPoolAPI)
	app.SetMempool(p.WrappedTxPool)

	p.WrappedMiner = miner.New(p.ExecutionClient.EngineAPI, p.logger)
	app.SetPrepareProposal(p.WrappedMiner.PrepareProposal)

	// Set the ante handler to nil, since it is not needed.
	app.SetAnteHandler(antelib.NewAnteHandler())

	return nil
}

// SetupServices initializes and registers the services with Polaris.
// It takes a client context as an argument and returns an error if the setup fails.
func (p *Polaris) SetupServices(clientCtx client.Context) error {
	// Initialize the miner with a new execution payload serializer.
	p.WrappedMiner.Init(libtx.NewSerializer[interfaces.ExecutionData](
		clientCtx.TxConfig, evmtypes.WrapPayload))

	// Initialize the txpool with a new transaction serializer.
	p.WrappedTxPool.Init(p.logger, clientCtx, libtx.NewSerializer[*coretypes.Transaction](
		clientCtx.TxConfig, evmtypes.WrapTx))
	return nil
}
