package proposals

import (
	"fmt"

	cometabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/prysmaticlabs/prysm/v4/consensus-types/blocks"
	enginev1 "github.com/prysmaticlabs/prysm/v4/proto/engine/v1"
	"pkg.berachain.dev/polaris/beacon/prysm"
)

// ValidateExtendedCommitInfo validates the extended commit info for a block. It first
// ensures that the vote extensions compose a supermajority of the signatures and
// voting power for the block. Then, it ensures that oracle vote extensions are correctly
// marshalled and contain valid prices.
func (h *ProposalHandler) ValidateExtendedCommitInfo(
	ctx sdk.Context,
	height int64,
	extendedCommitInfo cometabci.ExtendedCommitInfo,
) error {
	fmt.Println("VALIDATE VOTE EXTENSION")
	if err := h.validateVoteExtensionsFn(ctx, height, extendedCommitInfo); err != nil {
		h.logger.Error(
			"failed to validate vote extensions; vote extensions may not comprise a supermajority",
			"height", height,
			"err", err,
		)

		return err
	}

	// Validate all oracle vote extensions.
	fmt.Println("VALIDATE VOTE EXTENSION", len(extendedCommitInfo.Votes))
	for _, vote := range extendedCommitInfo.Votes {
		address := sdk.ConsAddress{}
		if err := address.Unmarshal(vote.Validator.Address); err != nil {
			h.logger.Error(
				"failed to unmarshal validator address",
				"height", height,
			)

			return err
		}

		// The vote extension are from the previous block.
		if err := h.ValidateOracleVoteExtension(ctx, vote.VoteExtension, uint64(height-1)); err != nil {
			h.logger.Error(
				"failed to validate oracle vote extension",
				"height", height,
				"validator", address.String(),
				"err", err,
			)

			return err
		}
	}

	return nil
}

func (h *ProposalHandler) ValidateOracleVoteExtension(ctx sdk.Context, voteExtension []byte, height uint64) error {
	fmt.Println("VALIDATE ORACLE VOTE EXTENSION")
	fmt.Println(len(voteExtension))

	builder := (&prysm.Builder{EngineCaller: h.miner})

	payload := new(enginev1.ExecutionPayloadCapellaWithValue)
	payload.Payload = new(enginev1.ExecutionPayloadCapella)
	if err := payload.Payload.UnmarshalSSZ(voteExtension); err != nil {
		h.logger.Error(
			"failed to unmarshal vote extension",
			"height", height,
			"err", err,
		)

		return err
	}
	// todo handle hardforks without needing codechange.
	data, err := blocks.WrappedExecutionPayloadCapella(
		payload.Payload, blocks.PayloadValueToGwei(payload.Value),
	)
	if err != nil {
		h.logger.Error(
			"failed to wrap vote extension",
			"height", height,
			"err", err,
		)
		return err
	}

	// TODO: switch evmtypes.WrappedPayloadEnvelope to just use the proto type
	// from prysm.
	_, err = builder.BlockValidation(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to validate payload: %w", err)
	}

	return nil
}
