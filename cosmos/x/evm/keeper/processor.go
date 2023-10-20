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

package keeper

import (
	"context"
	"fmt"

	"pkg.berachain.dev/polaris/beacon/prysm"
	evmtypes "pkg.berachain.dev/polaris/cosmos/x/evm/types"
)

func (k *Keeper) ProcessPayloadEnvelope(
	ctx context.Context, msg *evmtypes.WrappedPayloadEnvelope,
) (*evmtypes.WrappedPayloadEnvelopeResponse, error) {
	// var (
	// 	// payloadStatus engine.PayloadStatusV1
	// 	envelope engine.ExecutionPayloadEnvelope
	// 	err      error
	// )

	// if err = envelope.UnmarshalJSON(msg.Data); err != nil {
	// 	return nil, fmt.Errorf("failed to unmarshal payload envelope: %w", err)
	// }

	builder := (&prysm.Builder{EngineCaller: k.executionClient})
	payloadID, _, err := builder.BlockValidation(ctx, msg.UnwrapPayload())
	if err != nil {
		fmt.Println("PANICCCCCCCCC", err)
	}

	k.Logger(ctx).Info("Processed Payload ID", "payloadID", payloadID, "TODO: FIGURE OUT WHY NIL")

	return &evmtypes.WrappedPayloadEnvelopeResponse{}, nil
}

// EthTransaction implements the MsgServer interface. It is intentionally a no-op, but is required
// for the cosmos-sdk to not freak out.
func (k *Keeper) EthTransaction(
	context.Context, *evmtypes.WrappedEthereumTransaction,
) (*evmtypes.WrappedEthereumTransactionResult, error) {
	panic("intentionally not implemented")
}
