fabric-test:
	@cd tests/chaincode/fabcar/go && tar -zxvf vendor.tar.gz
#	@mkdir ./tests/tmp/ && cd ./tests/tmp/ && curl -sSL https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh | bash -s 2.2.1
	@cd tests/fabcar && ./startFabric.sh

stop-fabric-test:
	@sudo rm -rf /tmp/fabric-samples
	@cd tests/fabcar && ./stopFabric.sh \
	&& rm -rf credstore \
    && rm -rf cryptostore

fabex-test:
	@go run fabex.go -task=grpc -configpath=tests -configname=config -enrolluser=true -db=mongo

fabex-test-integration:
	@go run fabex.go -task=grpc -configpath=tests -configname=config-integration-tests -enrolluser=false -db=mongo


fabex-mongo:
	@go run fabex.go -task=grpc -configpath=configs -configname=config -enrolluser=true -db=mongo

fabex-cassandra:
	@go run fabex.go -task=grpc -configpath=configs -configname=config -enrolluser=true -db=cassandra

cassandra:
	@docker run -v /var/lib/cassandra:/var/lib/cassandra --name cassandra --net=host -d cassandra

mongo:
	@cd db/mongo-compose && docker-compose -f docker-compose.yaml up -d

mongo-test:
	@cd tests/db/mongo-compose && docker-compose -f docker-compose.yaml up -d

stop-mongo-test:
	@docker rm -f fabexmongo

unit-tests:
	@cd blockhandler && go test -v

integration-tests:
	@cd tests/chaincode/fabcar/go && tar -zxf vendor.tar.gz
	@sleep 10
	@cd client/ && go test -v
	@sleep 20
	@cd ./rest && go test -v