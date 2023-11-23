package greenfield

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

func TestGreenfieldDriver(t *testing.T) {
	godotenv.Load()

	privateKey := os.Getenv("GREEN_PRIVATE_KEY")

	bucket := "glc001-testnet-greenfield"
	rpcAddr := "https://gnfd-testnet-fullnode-tendermint-us.bnbchain.org:443"
	chainID := "greenfield_5600-1"

	key := "hello.txt"
	data := []byte("hello greenfield!")

	driver := GetGreenfieldDriver(rpcAddr, chainID, bucket, privateKey)

	fmt.Println("account:", driver.Account().GetAddress().String())

	ctx := context.Background()

	txHash, err := driver.Put(ctx, key, data)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("txHash", txHash)

	data0, txHash0, err := driver.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(data, data0) {
		t.Fatal("data not match")
	}
	if txHash != txHash0 {
		t.Fatal("txHash not match")
	}
}
