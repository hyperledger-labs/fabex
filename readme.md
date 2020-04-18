## **Block explorer for Hyperledger Fabric**

[Prerequisites](#prerequisites)

[CLI usage](#cli)

[Microservice mode](#microservice)


### <a name="prerequisites">**Prerequisites**</a>

1. Configure `config.yaml` (it's main config of the Fabex) and `connection-profile.yaml` (Hyperledger Fabric connection profile)

2. Start your Fabric blockchain network or sample [fist network](https://github.com/hyperledger/fabric-samples/tree/release-1.4/first-network)

3. Install and start PostgresSQL or MongoDB

    example for MongoDB: 
    
    1. set initial user name and password in `db/mongo-compose/docker-compose.yaml`
    2. start container:
    
    ```
        cd db/mongo-compose
        docker-compose -f docker-compose.yaml up -d
    ```
### <a name="cli">**CLI**</a>
1. Build Fabex executable binary file:  

    `go build`

2. Enroll admin user:  

    `./fabex --enrolluser true`

3. [For Postgres] Create database and table for data saving:  

    `./fabex --task initdb --configpath ./ --configname config`

4. Save blocks data to db:

    `./fabex --task explore --configpath ./ --configname config`
    

Also you can start service for fetching blocks in daemon mode: 
 
    `./fabex --task explore --configpath ./ --configname config --forever true --duration 1s` 
    
    
5. Get transactions of specific block (chain operation):  

    `./fabex --task getblock --blocknum 14 --configpath ./ --configname config`

6. Get all transactions (db operation):  

    `./fabex --task getall --configpath ./ --configname config`

---

You can choose PostgresSQL or MongoDB for data saving and retrieving with `--db flag`:

    `./fabex --task getall ---configpath ./ --configname config --db postgres`

    `./fabex --task explore --configpath ./ --configname config --db mongo`

---



### <a name="microservice">**Microservice mode**</a>

You can start Fabex as standalone microservice:

    `./fabex --task grpc --configpath ./ --configname config`

You can use [fabex.proto](https://github.com/VadimInshakov/fabex/blob/master/proto/fabex.proto) as service contract.

Sample client implementation [is here](https://github.com/VadimInshakov/fabex/blob/master/client/client.go). 
You can run example client with this command:

    `go run ./client/client.go`

