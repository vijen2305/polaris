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

// Package miner implements the Ethereum miner.
package miner

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/ethereum/go-ethereum/beacon/engine"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/miner"
	payloadattribute "github.com/prysmaticlabs/prysm/v4/consensus-types/payload-attribute"
	pb "github.com/prysmaticlabs/prysm/v4/proto/engine/v1"
	"pkg.berachain.dev/polaris/beacon/eth"
	"pkg.berachain.dev/polaris/eth/core/types"
)

// emptyHash is a common.Hash initialized to all zeros.
var emptyHash = common.Hash{}

// EnvelopeSerializer is used to convert an envelope into a byte slice that represents
// a cosmos sdk.Tx.
type EnvelopeSerializer interface {
	ToSdkTxBytes(*engine.ExecutionPayloadEnvelope, uint64) ([]byte, error)
}

// Miner implements the baseapp.TxSelector interface.
type Miner struct {
	eth.ConsensusAPI
	serializer         EnvelopeSerializer
	etherbase          common.Address
	curForkchoiceState *pb.ForkchoiceState
	lastBlockTime      uint64
}

// New produces a cosmos miner from a geth miner.
func New(gm eth.ConsensusAPI) *Miner {
	return &Miner{
		ConsensusAPI:       gm,
		curForkchoiceState: &pb.ForkchoiceState{},
	}
}

// Init sets the transaction serializer.
func (m *Miner) Init(serializer EnvelopeSerializer) {
	m.serializer = serializer
}

// PrepareProposal implements baseapp.PrepareProposal.
func (m *Miner) PrepareProposal(
	ctx sdk.Context, _ *abci.RequestPrepareProposal,
) (*abci.ResponsePrepareProposal, error) {
	var _ []byte
	var err error
	if _, err = m.buildBlock(ctx); err != nil {
		return nil, err
	}
	return &abci.ResponsePrepareProposal{Txs: [][]byte{}}, err
}

// finalizedBlockHash returns the block hash of the finalized block corresponding to the given number
// or nil if doesn't exist in the chain.
func (c *Miner) finalizedBlockHash(number uint64) *common.Hash {
	var finalizedNumber = number
	// if number%devEpochLength == 0 {
	// } else {
	// 	finalizedNumber = (number - 1) / devEpochLength * devEpochLength
	// }

	fmt.Println(finalizedNumber)
	if finalizedBlock, err := c.ConsensusAPI.BlockByNumber(context.Background(), big.NewInt(int64(finalizedNumber))); finalizedBlock != nil && err == nil {
		fh := finalizedBlock.Hash()
		fmt.Println("FH", fh)
		return &fh
	}
	return nil
}

// buildBlock builds and submits a payload, it also waits for the txs
// to resolve from the underying worker.
func (m *Miner) buildBlock(ctx sdk.Context) ([]byte, error) {
	var (
		err error
		// envelope *engine.ExecutionPayloadEnvelope
		sCtx = sdk.UnwrapSDKContext(ctx)
	)

	// Reset to CurrentBlock in case of the chain was rewound
	number, err := m.ConsensusAPI.BlockNumber(ctx)
	if err != nil {
		fmt.Println(err)
	}

	if header, err := m.ConsensusAPI.BlockByNumber(ctx, big.NewInt(int64(number))); err != nil {
		fmt.Println(err)
	} else if !bytes.Equal(m.curForkchoiceState.HeadBlockHash, header.Hash().Bytes()) {

		finalizedHash := m.finalizedBlockHash(header.Number().Uint64())

		m.setCurrentState(header.Hash().Bytes(), finalizedHash.Bytes())
	}

	tstamp := sCtx.BlockTime()
	var random [32]byte
	rand.Read(random[:])

	attrs, err := payloadattribute.New(&pb.PayloadAttributesV2{
		Timestamp:             uint64(tstamp.Unix()),
		SuggestedFeeRecipient: m.etherbase.Bytes(),
		Withdrawals:           nil,
		PrevRandao:            append([]byte{}, random[:]...),
	})
	if err != nil {
		return nil, err
	}

	fcResponse, latestValidHash, err := m.ConsensusAPI.ForkchoiceUpdated(ctx, m.curForkchoiceState, attrs)
	if err != nil {
		fmt.Println(err)
	}
	time.Sleep(400 * time.Millisecond)
	// // Build Payload
	// parent := m.CurrentBlock(ctx)
	// if envelope, err = m.BuildBlock(ctx, m.constructPayloadArgs(sCtx, parent)); err != nil {
	// 	sCtx.Logger().Error("failed to build payload", "err", err)
	// 	return nil, err
	// }

	data, _, _, err := m.ConsensusAPI.GetPayload(ctx, *fcResponse, 100000000000)
	if err != nil {
		fmt.Println(err)
	}
	time.Sleep(400 * time.Millisecond)

	// // interfaces.ExecutionData, *pb.BlobsBundle, bool, error
	// if data. == engine.STATUS_SYNCING {
	// 	return errors.New("chain rewind prevented invocation of payload creation")
	// }
	// Mark the payload as canon
	if _, err = m.ConsensusAPI.NewPayload(ctx, data, nil, nil); err != nil {
		if err != nil {
			fmt.Println(err)
		}
		return nil, err
	}
	time.Sleep(400 * time.Millisecond)
	m.setCurrentState(data.BlockHash(), latestValidHash)

	// // m.ConsensusAPI.NewPayload(ctx)

	// bz, err := m.serializer.ToSdkTxBytes(data.En, envelope.ExecutionPayload.GasLimit)
	// if err != nil {
	// 	return nil, err
	// }
	time.Sleep(400 * time.Millisecond)
	attrs2, err := payloadattribute.New(&pb.PayloadAttributesV2{})
	if err != nil {
		return nil, err
	}

	fcResponse, latestValidHash, err = m.ConsensusAPI.ForkchoiceUpdated(ctx, m.curForkchoiceState, attrs2)
	if err != nil {
		fmt.Println(err)
	}
	return []byte{}, nil
}

// setCurrentState sets the current forkchoice state
func (c *Miner) setCurrentState(headHash, finalizedHash []byte) {
	c.curForkchoiceState = &pb.ForkchoiceState{
		HeadBlockHash:      headHash,
		SafeBlockHash:      headHash,
		FinalizedBlockHash: finalizedHash,
	}
}

// constructPayloadArgs builds a payload to submit to the miner.
func (m *Miner) constructPayloadArgs(
	ctx sdk.Context, parent *types.Block) *miner.BuildPayloadArgs {
	// etherbase, err := m.Etherbase(ctx)
	// if err != nil {
	// 	ctx.Logger().Error("failed to get etherbase", "err", err)
	// 	return nil
	// }

	return &miner.BuildPayloadArgs{
		Timestamp: parent.Header().Time + 2, //nolint:gomnd // todo fix this arbitrary number.
		// FeeRecipient: etherbase,
		Random:      common.Hash{}, /* todo: generated random */
		Withdrawals: make(types.Withdrawals, 0),
		BeaconRoot:  &emptyHash,
		Parent:      parent.Hash(),
	}
}
