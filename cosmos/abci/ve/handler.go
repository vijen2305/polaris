package ve

import (
	"context"
	"fmt"
	"time"

	"cosmossdk.io/log"
	cometabci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"pkg.berachain.dev/polaris/cosmos/runtime/miner"
)

// VoteExtensionHandler is a handler that extends a vote with the oracle's
// current price feed. In the case where oracle data is unable to be fetched
// or correctly marshalled, the handler will return an empty vote extension to
// ensure liveliness.
type VoteExtensionHandler struct {
	logger log.Logger

	// timeout is the maximum amount of time to wait for the oracle to respond
	// to a price request.
	timeout time.Duration

	miner *miner.Miner
}

// NewVoteExtensionHandler returns a new VoteExtensionHandler.
func NewVoteExtensionHandler(logger log.Logger, timeout time.Duration, miner *miner.Miner) *VoteExtensionHandler {
	return &VoteExtensionHandler{
		logger:  logger,
		timeout: timeout,
		miner:   miner,
	}
}

// ExtendVoteHandler returns a handler that extends a vote with the oracle's
// current price feed. In the case where oracle data is unable to be fetched
// or correctly marshalled, the handler will return an empty vote extension to
// ensure liveliness.
func (h *VoteExtensionHandler) ExtendVoteHandler() sdk.ExtendVoteHandler {
	return func(ctx sdk.Context, req *cometabci.RequestExtendVote) (resp *cometabci.ResponseExtendVote, err error) {
		// Catch any panic that occurs in the oracle request.
		defer func() {
			if r := recover(); r != nil {
				h.logger.Error(
					"recovered from panic in ExtendVoteHandler",
					"err", r,
				)

				resp, err = &cometabci.ResponseExtendVote{VoteExtension: []byte{}}, nil
			}
		}()

		// Create a context with a timeout to ensure we do not wait forever for the oracle
		// to respond.
		_, cancel := context.WithTimeout(ctx.Context(), h.timeout)
		defer cancel()

		// voteExt := vetypes.ForkChoiceExtension{
		// 	Height: req.Height,
		// }

		// voteExt :=

		bz, err := h.miner.BuildVoteExtension(ctx, req.Height)
		if err != nil {
			h.logger.Error(
				"failed to marshal vote extension; returning empty vote extension",
				"height", req.Height,
				"err", err,
			)

			return &cometabci.ResponseExtendVote{VoteExtension: []byte{}}, nil
		}

		h.logger.Info(
			"extending vote with oracle prices",
			"vote_extension_height", req.Height,
			"req_height", req.Height,
		)

		return &cometabci.ResponseExtendVote{VoteExtension: bz}, nil
	}
}

// VerifyVoteExtensionHandler returns a handler that verifies the vote extension provided by
// a validator is valid. In the case when the vote extension is empty, we return ACCEPT. This means
// that the validator may have been unable to fetch prices firom the oracle and s voting an empty vote extension.
// We reject any vote extensions that are not empty and fail to unmarshal or contain invalid prices.
func (h *VoteExtensionHandler) VerifyVoteExtensionHandler() sdk.VerifyVoteExtensionHandler {
	return func(ctx sdk.Context, req *cometabci.RequestVerifyVoteExtension) (*cometabci.ResponseVerifyVoteExtension, error) {

		fmt.Println("VERIFYING VOTE EXTENSION")
		// The line `// voteExtension := req.VoteExtension` is a comment in Go code. It is not actually doing
		// anything in the code. It is just a comment to provide information or explanation about the code.
		// voteExtension := req.VoteExtension

		// if err := ValidateOracleVoteExtension(voteExtension, req.Height); err != nil {
		// 	h.logger.Error(
		// 		"failed to validate vote extension",
		// 		"height", req.Height,
		// 		"err", err,
		// 	)

		// 	return &cometabci.ResponseVerifyVoteExtension{Status: cometabci.ResponseVerifyVoteExtension_REJECT}, err
		// }

		h.logger.Info(
			"validated vote extension",
			"height", req.Height,
		)

		return &cometabci.ResponseVerifyVoteExtension{Status: cometabci.ResponseVerifyVoteExtension_ACCEPT}, nil
	}
}
