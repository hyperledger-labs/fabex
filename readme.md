**Block explorer for Hyperledger Fabric**

Usage scenario example:  
(_you can skip any step)_

configure `config.yaml` and `connection-profile.yaml`

start [fist network](https://github.com/hyperledger/fabric-samples/tree/release-1.4/first-network)

install and start Postgres

build:  
`go build`

enroll user:  
`./fabex --enrolluser true`

create database and table for data saving:  
`./fabex --task initdb --config ./config.yaml`

get transactions of specific block (chain operation):  
`./fabex --task getblock --blocknum 14 --config ./config.yaml`

get all transactions (db operation):  
`./fabex --task getblock --blocknum 15 --config ./config.yaml`

---

Or you can use GRPC client with example RPC call in client/client.go:

`go run client.go`
