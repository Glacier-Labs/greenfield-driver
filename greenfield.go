package greenfield

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/bnb-chain/greenfield-go-sdk/client"
	"github.com/bnb-chain/greenfield-go-sdk/types"
	"go.uber.org/zap"
)

type GreenfieldDriver struct {
	RpcAddr    string
	ChainID    string
	Bucket     string
	privateKey string
	account    *types.Account
	gnfdclient client.IClient
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
		privateKey: privateKey,
		account:    account,
		gnfdclient: gnfdclient,
		logger:     logger.With(zap.Any("bucket", bucket)),
	}
}

func (gd *GreenfieldDriver) Put(ctx context.Context, key string, data []byte) (txHash string, err error) {
	objectDetail, err := gd.gnfdclient.HeadObject(ctx, gd.Bucket, key)
	if err != nil && !strings.Contains(err.Error(), "No such object") {
		return
	}

	if objectDetail != nil {
		objectInfo := objectDetail.ObjectInfo
		gd.logger.Info("HeadObject", zap.Any("key", key), zap.Any("objectInfo", objectInfo), zap.Any("err", err))

		if objectInfo.GetObjectStatus().String() == "OBJECT_STATUS_SEALED" {
			err = fmt.Errorf("key already exists")
			return
		}
	} else {
		txHash, err = gd.gnfdclient.CreateObject(ctx, gd.Bucket, key, bytes.NewReader(data), types.CreateObjectOptions{})
		if err != nil {
			gd.logger.Error("CreateObject", zap.Any("key", key), zap.Any("err", err.Error()), zap.Any("txHash", txHash))
			return
		}
	
		gd.logger.Info("CreateObject", zap.Any("key", key), zap.Any("txHash", txHash), zap.Any("size", len(data)))	
	}


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
			headObjOutputDetail, err := gd.gnfdclient.HeadObject(ctx, gd.Bucket, key)
			if err != nil {
				return txHash, err
			}

			headObjOutput := headObjOutputDetail.ObjectInfo
			gd.logger.Info("HeadObject", zap.Any("key", key), zap.Any("objectInfo", headObjOutput), zap.Any("err", err))
			if headObjOutput.GetObjectStatus().String() == "OBJECT_STATUS_SEALED" {
				ticker.Stop()
				gd.logger.Info("PutObject successfully", zap.Any("key", key))
				return txHash, nil
			}
		}
	}
}

func (gd *GreenfieldDriver) Get(ctx context.Context, key string) (data []byte, err error) {
	objectDataReader, _, err := gd.gnfdclient.GetObject(ctx, gd.Bucket, key, types.GetObjectOptions{})
	if err != nil {
		return
	}

	data, err = ioutil.ReadAll(objectDataReader)
	return
}

func (gd *GreenfieldDriver) Account() *types.Account {
	return gd.account
}
