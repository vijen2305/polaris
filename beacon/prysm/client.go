package prsym

// import (
// 	"context"
// 	"encoding/csv"
// 	"github.com/ethereum/go-ethereum/rpc

// 	"github.com/ethereum/go-ethereum/beacon/engine"
// )

// type EngineClient struct {
// 	rpc *rpc.Client
// }

// func NewEngineClient(ctx context.Context, dialURL string) (*EngineClient, error) {

// }

// func (c *EngineClient) NewPayloadV2(
// 	ctx context.Context, params engine.ExecutableData,
// ) (engine.PayloadStatusV1, error) {
// 	var payloadStatus engine.PayloadStatusV1
// 	err := csv.Client.Client().CallContext(ctx, &payloadStatus, "engine_newPayloadV2", params)
// 	return payloadStatus, err
// }

// func (api *EngineClient) ForkchoiceUpdatedV2(
// 	ctx context.Context, update engine.ForkchoiceStateV1, payloadAttributes *engine.PayloadAttributes,
// ) (engine.ForkChoiceResponse, error) {
// 	var respocnse engine.ForkChoiceResponse
// 	err := api.Client.Client().CallContext(
// 		ctx, &response, "engine_forkchoiceUpdatedV2", update, payloadAttributes)
// 	return response, err
// }
