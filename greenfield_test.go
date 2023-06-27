package greenfield

import (
	"bytes"
	"context"
	"fmt"
	"testing"
)

func TestGreenfieldDriver(t *testing.T) {
	privateKey := "<YOUR-PRVIATE-KEY>"

	bucket := "glc001-testnet-greenfield"
	rpcAddr := "https://gnfd-testnet-fullnode-tendermint-us.bnbchain.org:443"
	chainID := "greenfield_5600-1"

	key := "hello.txt"
	data := []byte("hello greenfield!")

	driver := GetGreenfieldDriver(rpcAddr, chainID, bucket, privateKey)

	ctx := context.Background()
	txHash, err := driver.Put(ctx, key, data)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("txHash", txHash)

	data0, err := driver.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, data0) {
		t.Fatal("data not match")
	}
}
