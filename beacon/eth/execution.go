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

package eth

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/rpc"

	"github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/miner"
	"github.com/prysmaticlabs/prysm/v4/network"
	prsymnetwork "github.com/prysmaticlabs/prysm/v4/network"
	"github.com/prysmaticlabs/prysm/v4/network/authorization"
	"pkg.berachain.dev/polaris/beacon/log"
	prsym "pkg.berachain.dev/polaris/beacon/prysm"
	"pkg.berachain.dev/polaris/eth/common"
	"pkg.berachain.dev/polaris/eth/core"
	coretypes "pkg.berachain.dev/polaris/eth/core/types"
)

type (
	// BuilderAPI represents the `Miner` that exists on the backend of the execution layer.
	BuilderAPI interface {
		BuildBlock(context.Context, *miner.BuildPayloadArgs) (*engine.ExecutionPayloadEnvelope, error)
		Etherbase(context.Context) (common.Address, error)
		BlockByNumber(uint64) (*coretypes.Block, error)
		CurrentBlock(ctx context.Context) (*coretypes.Block, error)
	}

	// TxPool represents the `TxPool` that exists on the backend of the execution layer.
	TxPoolAPI interface {
		Add([]*coretypes.Transaction, bool, bool) []error
		Stats() (int, int)
		SubscribeNewTxsEvent(chan<- core.NewTxsEvent) event.Subscription
	}

	ConsensusAPI interface {
		prsym.EngineCaller
		BlockByNumber(context.Context, *big.Int) (*coretypes.Block, error)
		BlockNumber(context.Context) (uint64, error)
	}
)

// ExecutionClient represents the execution layer client.
type ExecutionClient struct {
	BlockBuilder BuilderAPI
	TxPool       TxPoolAPI
	Consensus    ConsensusAPI
}

// NewRemoteExecutionClient creates a new remote execution client.
func NewRemoteExecutionClient(dialURL, jwtSecret string, logger log.Logger) (*ExecutionClient, error) {
	ctx := context.Background()
	var (
		client  *rpc.Client
		chainID *big.Int
		err     error
	)

	jwtSecretR := common.FromHex(strings.TrimSpace(string(jwtSecret)))

	endpoint := NewPrysmEndpoint(dialURL, jwtSecretR)
	client, err = newRPCClientWithAuth(ctx, nil, endpoint)
	if err != nil {
		return nil, err
	}

	var ethClient *ethclient.Client
	for i := 0; i < 100; func() { i++; time.Sleep(time.Second) }() {
		logger.Info("waiting for connection to execution layer", "dial-url", dialURL)
		ethClient = ethclient.NewClient(client)
		chainID, err = ethClient.ChainID(ctx)
		fmt.Println(err)
		if err != nil {
			continue
		}
		logger.Info("Successfully connected to execution layer", "ChainID", chainID)
		break
	}
	if client == nil || err != nil {
		return nil, fmt.Errorf("failed to establish connection to execution layer: %w", err)
	}

	prsymClient := prsym.NewEngineClientService(ethClient)

	return &ExecutionClient{
		// BlockBuilder: &builderAPI{Client: client},
		TxPool:    &txPoolAPI{Client: ethClient},
		Consensus: prsymClient,
	}, nil
}

// Initializes an RPC connection with authentication headers.
func newRPCClientWithAuth(ctx context.Context, headersMap map[string]string, endpoint network.Endpoint) (*rpc.Client, error) {
	headers := http.Header{}
	if endpoint.Auth.Method != authorization.None {
		header, err := endpoint.Auth.ToHeaderValue()
		if err != nil {
			return nil, err
		}
		headers.Set("Authorization", header)
	}
	for _, h := range headersMap {
		if h == "" {
			continue
		}
		keyValue := strings.Split(h, "=")
		if len(keyValue) < 2 {
			// log.LoggerWarn("Incorrect HTTP header flag format. Skipping %v", keyValue[0])
			continue
		}
		headers.Set(keyValue[0], strings.Join(keyValue[1:], "="))
	}

	fmt.Println("HEADERS", headers)
	return network.NewExecutionRPCClient(ctx, endpoint, headers)
}

func NewPrysmEndpoint(endpointString string, secret []byte) prsymnetwork.Endpoint {
	if len(secret) == 0 {
		return network.HttpEndpoint(endpointString)
	}
	// Overwrite authorization type for all endpoints to be of a bearer type.
	hEndpoint := network.HttpEndpoint(endpointString)
	hEndpoint.Auth.Method = authorization.Bearer
	hEndpoint.Auth.Value = string(secret)

	return hEndpoint

}
