# greenfield-driver

This is a Greenfield driver for Glacier to integrate Greenfield's feature! It implements Glacier's Standard Storage interface.

# Example

gnfd-cmd: https://github.com/bnb-chain/greenfield-cmd
Test BucketName: glc001-testnet-greenfield

- Create Bucket (public-read)

```
gnfd-cmd bucket create --visibility=public-read  gnfd://glc001-testnet-greenfield
```

- Upload Data by driver (inherit the bucket visibility)

```
go test -v
```

- List Objects

```
gnfd-cmd object ls  gnfd://glc001-testnet-greenfield
```