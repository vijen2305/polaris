package main

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"pkg.berachain.dev/polaris/eth/common"
)

func main() {
	cosmosAddr := "cosmos10fqsgcm6k9q7xd7cmt0jdzy3ffmkezqaa765zh"
	accAddress := sdk.MustAccAddressFromBech32(cosmosAddr)
	fmt.Println(common.BytesToAddress(accAddress.Bytes()))
}
