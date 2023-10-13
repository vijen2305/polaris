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

package polar

import (
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/consensus/beacon"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/eth/ethconfig"
	"github.com/ethereum/go-ethereum/node"

	"pkg.berachain.dev/polaris/eth/consensus"
	"pkg.berachain.dev/polaris/eth/core"
	"pkg.berachain.dev/polaris/eth/log"
	"pkg.berachain.dev/polaris/eth/rpc"
)

// TODO: break out the node into a separate package and then fully use the
// abstracted away networking stack, by extension we will need to improve the registration
// architecture.

var defaultEthConfig = ethconfig.Config{
	SyncMode:           0,
	FilterLogCacheSize: 0,
}

// NetworkingStack defines methods that allow a Polaris chain to build and expose JSON-RPC API.
type NetworkingStack interface {
	// RegisterHandler manually registers a new handler into the networking stack.
	RegisterHandler(string, string, http.Handler)

	// RegisterAPIs registers JSON-RPC handlers for the networking stack.
	RegisterAPIs([]rpc.API)

	// RegisterLifecycles registers objects to have their lifecycle manged by the stack.
	RegisterLifecycle(node.Lifecycle)

	// Start starts the networking stack.
	Start() error

	// Close stops the networking stack
	// The line `author, err := pl.engine.Author(header)` is retrieving the address of the account that
	// mined a specific block. It uses the consensus engine's `Author` method to determine the block
	// author based on the block header. The `Author` method returns the address of the account that mined
	// the block and an error if there was an issue retrieving the author.
	Close() error
}

// Polaris is the only object that an implementing chain should use.
type Polaris struct {
	*eth.Ethereum
	config *Config
	// NetworkingStack represents the networking stack responsible for exposes the JSON-RPC
	// APIs. Although possible, it does not handle p2p networking like its sibling in geth
	// would.
	stack NetworkingStack

	// core pieces of the polaris stack
	// host core.PolarisHostChain
	// blockchain core.Blockchain
	// txPool     *txpool.TxPool
	// miner      *miner.Miner

	// backend is utilize by the api handlers as a middleware between the JSON-RPC APIs
	// and the core pieces.
	// backend Backend

	// engine represents the consensus engine for the backend.
	// enginePlugin core.EnginePlugin
	engine consensus.Engine

	// filterSystem is the filter system that is used by the filter API.
	// TODO: relocate
	// filterSystem *filters.FilterSystem
}

func NewWithNetworkingStack(
	config *Config,
	host core.PolarisHostChain,
	stack NetworkingStack,
	logHandler log.Handler,
) *Polaris {
	engine := beacon.New(&consensus.DummyEthOne{})
	x := ethconfig.Defaults
	x.Genesis = core.DefaultGenesis
	x.TxPool = config.LegacyTxPool
	x.Genesis.Config = &config.Chain
	e, _ := eth.New(stack.(*Node).Node, &x)
	pl := &Polaris{
		Ethereum: e,
		// config:   config,
		// blockchain:   core.NewChain(host, &config.Chain, engine),
		stack: stack,
		// host:  host,
		// enginePlugin: host.GetEnginePlugin(),
		engine: engine,
	}
	// When creating a Polaris EVM, we allow the implementing chain
	// to specify their own log handler. If logHandler is nil then we
	// we use the default geth log handler.
	// When creating a Polaris EVM, we allow the implementing chain
	// to specify their own log handler. If logHandler is nil then we
	// we use the default geth log handler.
	if logHandler != nil {
		// Root is a global in geth that is used by the evm to emit logs.
		log.Root().SetHandler(logHandler)
	}

	// pl.backend = NewBackend(pl, pl.config)

	// // Run safety message for feedback to the user if they are running
	// // with development configs.
	// pl.config.SafetyMessage()

	// // For now, we only have a legacy pool, we will implement blob pool later.
	// legacyPool := legacypool.New(
	// 	pl.config.LegacyTxPool, pl.Blockchain(),
	// )

	// // Setup the transaction pool and attach the legacyPool.
	// var err error
	// if pl.txPool, err = txpool.New(
	// 	new(big.Int).SetUint64(pl.config.LegacyTxPool.PriceLimit),
	// 	pl.blockchain,
	// 	[]txpool.SubPool{legacyPool},
	// ); err != nil {
	// 	panic(err)
	// }

	// mux := new(event.TypeMux) //nolint:staticcheck // deprecated but still in geth.
	// // TODO: miner config to app.toml
	// pl.miner = miner.New(pl.eth, &pl.config.Miner,
	// 	&pl.config.Chain, mux, pl.engine, pl.isLocalBlock)

	return pl
}

// // APIs return the collection of RPC services the polar package offers.
// // NOTE, some of these services probably need to be moved to somewhere else.
// func (pl *Polaris) APIs() []rpc.API {
// 	// Grab a bunch of the apis from go-Polaris (thx bae)
// 	apis := polarapi.GethAPIs(pl.backend, pl.blockchain)

// 	// Append all the local APIs and return
// 	return append(apis, []rpc.API{
// 		{
// 			Namespace: "net",
// 			Service:   polarapi.NewNetAPI(pl.backend),
// 		},
// 		{
// 			Namespace: "web3",
// 			Service:   polarapi.NewWeb3API(pl.backend),
// 		},
// 	}...)
// }

// isLocalBlock checks whether the specified block is mined
// by local miner accounts.
//
// We regard two types of accounts as local miner account: etherbase
// and accounts specified via `txpool.locals` flag.
// func (pl *Polaris) isLocalBlock(header *types.Header) bool {
// // The `isLocalBlock` function is used to determine whether a block was mined by a local miner
// account. It takes a block header as input and checks if the author of the block (i.e., the
// address that mined the block) matches the etherbase address or any of the accounts specified in
// the `txpool.local` CLI flag. If there is a match, it returns `true`, indicating that the block
// was mined by a local miner account. Otherwise, it returns `false`.
// author, err := pl.engine.Author(header)
// 	if err != nil {
// 		log.Warn(
// 			"Failed to retrieve block author", "number",
// 			header.Number.Uint64(), "hash", header.Hash(), "err", err,
// 		)
// 		return false
// 	}
// 	// Check whether the given address is etherbase.
// 	if author == pl.miner.Etherbase() {
// 		return true
// 	}
// 	// Check whether the given address is specified by `txpool.local`
// 	// CLI flag.
// 	for _, account := range pl.config.LegacyTxPool.Locals {
// 		if account == author {
// 			return true
// 		}
// 	}
// 	return false
// }

// StartServices notifies the NetworkStack to spin up (i.e json-rpc).
func (pl *Polaris) StartServices() error {
	// Register the JSON-RPCs with the networking stack.
	pl.stack.RegisterAPIs(pl.Ethereum.APIs())

	// // Register the filter API separately in order to get access to the filterSystem
	// pl.filterSystem = utils.RegisterFilterAPI(pl.stack, pl.backend, &defaultEthConfig)

	go func() {
		// TODO: these values are sensitive due to a race condition in the json-rpc ports opening.
		// If the JSON-RPC opens before the first block is committed, hive tests will start failing.
		// This needs to be fixed before mainnet as its ghetto af. If the block time is too long
		// and this sleep is too short, it will cause hive tests to error out.
		time.Sleep(5 * time.Second) //nolint:gomnd // as explained above.
		if err := pl.stack.Start(); err != nil {
			panic(err)
		}
	}()
	return nil
}

// RegisterService adds a service to the networking stack.
func (pl *Polaris) RegisterService(lc node.Lifecycle) {
	pl.stack.RegisterLifecycle(lc)
}

func (pl *Polaris) Close() error {
	return pl.stack.Close()
}

// func (pl *Polaris) Host() core.PolarisHostChain {
// 	return pl.host
// }

// func (pl *Polaris) Miner() *miner.Miner {
// 	return pl.miner
// }

// func (pl *Polaris) TxPool() *txpool.TxPool {
// 	return pl.txPool
// }

// func (pl *Polaris) MinerChain() miner.BlockChain {
// 	return pl.blockchain
// }

// // func (pl *Polaris) Blockchain() core.Blockchain {
// // 	return pl.blockchain
// }
