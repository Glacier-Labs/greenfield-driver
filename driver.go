package greenfield

import (
	"context"
)

type IDriver interface {
	Put(ctx context.Context, key string, data []byte) (txHash string, err error)
	Get(ctx context.Context, key string) (data []byte, txHash string, err error)
	Type() string
	DaID(dataHash string, txHash string) string
}

func GetGreenfieldDriver(rpcAddr, chainID, bucket, privateKey string) IDriver {
	return NewGreenfieldDriver(rpcAddr, chainID, bucket, privateKey)
}
