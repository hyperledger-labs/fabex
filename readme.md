**Block explorer for Hyperledger Fabric**

Usage scenario example:  
(_you can skip any step)_

configure `config.yaml` and `connection-profile.yaml`

start [fist network](https://github.com/hyperledger/fabric-samples/tree/release-1.4/first-network)

install and start PostgresSQL or MongoDB

build:  
`go build`

enroll user:  
`./fabex --enrolluser true`

create database and table for data saving:  
`./fabex --task initdb --config ./config.yaml`

save blocks data to db:
`./fabex --task explore --config ./config.yaml`

get transactions of specific block (chain operation):  
`./fabex --task getblock --blocknum 14 --config ./config.yaml`

get all transactions (db operation):  
`./fabex --task getblock --blocknum 15 --config ./config.yaml`

---

You can choose PostgresSQL or MongoDB for data saving and retrieving with `--db flag`:       

`./fabex --task getall --config ./config.yaml --db postgres`

`./fabex --task explore --config ./config.yaml --db mongo`

---

Or you can use GRPC client with example RPC call in client/client.go:

`go run client.go`

---

Start service in daemon mode:  
`go run fabex.go --task explore --config ./config.yaml --forever true --duration 1s`
