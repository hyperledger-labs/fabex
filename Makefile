start-fabric-test:
	@cd tests/fabcar && ./startFabric.sh

stop-fabric-test:
	@cd tests/fabcar && ./stopFabric.sh

start-fabex-test:
	@go run fabex.go -task=grpc -configpath=tests -configname=config -enrolluser=true -db=mongo

start-fabex:
	@go run fabex.go -task=grpc -configpath=configs -configname=config -enrolluser=true -db=mongo

mongo-test:
	@cd tests/db/mongo-compose && docker-compose -f docker-compose.yaml up -d

stop-mongo-test:
	@docker rm -f fabexmongo \
    && rm -rf credstore \
    && rm -rf cryptostore

unit-tests:
	@cd blockfetcher && go test -v

integration-tests:
	@cd client/fabexclient && go test -v