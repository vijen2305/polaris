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

package prysm

import (
	"context"
	"fmt"

	"github.com/prysmaticlabs/prysm/v4/consensus-types/interfaces"
	payloadattribute "github.com/prysmaticlabs/prysm/v4/consensus-types/payload-attribute"
	enginev1 "github.com/prysmaticlabs/prysm/v4/proto/engine/v1"
)

type Builder struct {
	EngineCaller
}

func (b *Builder) BlockProposal(
	ctx context.Context, payload interfaces.ExecutionData, attrs payloadattribute.Attributer,
) (interfaces.ExecutionData, error) {
	payloadID, latestValidHash, err := b.BlockValidation(ctx, payload)
	if err != nil {
		return nil, err
	}
	if latestValidHash == nil {
		return nil, err
	}

	builtPayload, _, _, err := b.GetPayload(ctx, *payloadID, 100000000000)
	// todo: wait for slot tick or equivalent

	return builtPayload, err
}

// BlockValidation builds a payload from the provided execution data, and submits it to
// the execution client. It then submits a forkchoice update with the updated
// valid hash returned by the execution client.
// This should be called by a node when it receives Execution Data as part of
// beacon chain consensus.
// receives payload -> get latestValidHash from our execution client -> forkchoice locally.
func (b *Builder) BlockValidation(
	ctx context.Context, payload interfaces.ExecutionData,
) (*enginev1.PayloadIDBytes, []byte, error) {
	// new Payload
	latestValidHash, err := b.NewPayload(ctx, payload, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	// TODO: wait for potential reorg?? on new payload.
	// time.Sleep(800 * time.Millisecond) //nolint:gomnd // temp.
	var payloadID *enginev1.PayloadIDBytes
	payloadID, _, err = b.ForkchoiceUpdated(ctx, &enginev1.ForkchoiceState{
		HeadBlockHash: latestValidHash,
		// The two below are technically incorrect? These should be set later imo.
		SafeBlockHash:      latestValidHash,
		FinalizedBlockHash: latestValidHash,
	}, payloadattribute.EmptyWithVersion(3))

	fmt.Println("PAYLOAD ID", payloadID)

	return payloadID, latestValidHash, err
}
