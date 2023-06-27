package greenfield

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	"go.uber.org/zap"
)

type GreenfieldDriver struct {
	RpcAddr    string
	ChainID    string
	Bucket     string
	PrivateKey string
	gnfdclient client.Client
	logger     *zap.Logger
}

func NewGreenfieldDriver(rpcAddr, chainID, bucket, privateKey string) *GreenfieldDriver {
	account, err := types.NewAccountFromPrivateKey("Bundler", privateKey)
	if err != nil {
		panic(err)
	}
	fmt.Println(rpcAddr, account.GetAddress().String())

	gnfdclient, err := client.New(chainID, rpcAddr, client.Option{DefaultAccount: account})
	if err != nil {
		panic(err)
	}

	logger, _ := zap.NewProduction()

	return &GreenfieldDriver{
		RpcAddr:    rpcAddr,
		ChainID:    chainID,
		Bucket:     bucket,
		PrivateKey: privateKey,
		gnfdclient: gnfdclient,
		logger:     logger.With(zap.Any("bucket", bucket)),
	}
}

func (gd *GreenfieldDriver) Put(ctx context.Context, key string, data []byte) (txHash string, err error) {
	objectInfo, err := gd.gnfdclient.HeadObject(ctx, gd.Bucket, key)
	gd.logger.Info("HeadObject", zap.Any("key", key), zap.Any("objectInfo", objectInfo), zap.Any("err", err))

	if err != nil {
		txHash, err = gd.gnfdclient.CreateObject(ctx, gd.Bucket, key, bytes.NewReader(data), types.CreateObjectOptions{})
		if err != nil {
			gd.logger.Error("CreateObject", zap.Any("key", key), zap.Any("err", err.Error()), zap.Any("txHash", txHash))
			return
		}
	}

	if objectInfo.GetObjectStatus().String() == "OBJECT_STATUS_SEALED" {
		err = fmt.Errorf("key already exists")
		return
	}

	gd.logger.Info("CreateObject", zap.Any("key", key), zap.Any("txHash", txHash), zap.Any("size", len(data)))

	err = gd.gnfdclient.PutObject(ctx, gd.Bucket, key, int64(len(data)), bytes.NewReader(data), types.PutObjectOptions{
		TxnHash: txHash,
	})
	if err != nil {
		gd.logger.Error("PubOject", zap.Any("key", key), zap.Any("err", err.Error()))
		return
	}

	// Check if object is sealed
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-timeout:
			err = fmt.Errorf("reach to the wait limit: %s", key)
			return
		case <-ticker.C:
			headObjOutput, err := gd.gnfdclient.HeadObject(ctx, gd.Bucket, key)
			gd.logger.Info("HeadObject", zap.Any("key", key), zap.Any("objectInfo", headObjOutput), zap.Any("err", err))
			if err != nil {
				return txHash, err
			}

			if headObjOutput.GetObjectStatus().String() == "OBJECT_STATUS_SEALED" {
				ticker.Stop()
				gd.logger.Info("PutObject successfully", zap.Any("key", key))
				return txHash, nil
			}
		}
	}
}

func (gd *GreenfieldDriver) Get(ctx context.Context, key string) (data []byte, err error) {
	objectDataReader, _, err := gd.gnfdclient.GetObject(ctx, gd.Bucket, key, types.GetObjectOption{})
	if err != nil {
		return
	}

	data, err = ioutil.ReadAll(objectDataReader)
	return
}
