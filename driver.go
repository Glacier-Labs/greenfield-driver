package greenfield

import (
	"context"

	"github.com/bnb-chain/greenfield-go-sdk/types"
)

type IDriver interface {
	Put(ctx context.Context, key string, data []byte) (txHash string, err error)
	Get(ctx context.Context, key string) (data []byte, txHash string, err error)
	Account() *types.Account
}

func GetGreenfieldDriver(rpcAddr, chainID, bucket, privateKey string) IDriver {
	return NewGreenfieldDriver(rpcAddr, chainID, bucket, privateKey)
}
