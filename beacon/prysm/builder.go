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
	"time"

	"github.com/prysmaticlabs/prysm/v4/consensus-types/interfaces"
	payloadattribute "github.com/prysmaticlabs/prysm/v4/consensus-types/payload-attribute"
	enginev1 "github.com/prysmaticlabs/prysm/v4/proto/engine/v1"
)

type Builder struct {
	*Service
}

func (b *Builder) BlockProposal(ctx context.Context) error {
	return nil
}

func (b *Builder) BlockValidation(ctx context.Context, payload interfaces.ExecutionData) error {
	// new Payload
	latestValidHash, err := b.Service.NewPayload(ctx, payload, nil, nil)
	if err != nil {
		return err
	}

	// TODO: wait for potential reorg?? on new payload.
	time.Sleep(800 * time.Millisecond) //nolint:gomnd // temp.
	_, _, err = b.Service.ForkchoiceUpdated(ctx, &enginev1.ForkchoiceState{
		HeadBlockHash:      latestValidHash,
		SafeBlockHash:      latestValidHash,
		FinalizedBlockHash: latestValidHash,
	}, payloadattribute.EmptyWithVersion(2))

	return err
}
