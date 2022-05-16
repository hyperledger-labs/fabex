/*
   Copyright 2019 Vadim Inshakov

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/hyperledger-labs/fabex/log"

	"github.com/hyperledger-labs/fabex/api/rest"

	"github.com/hyperledger-labs/fabex/api/grpc"

	"go.uber.org/zap"

	fabconfig "github.com/hyperledger/fabric-sdk-go/pkg/core/config"

	"github.com/hyperledger-labs/fabex/config"
	"github.com/hyperledger-labs/fabex/db"
	"github.com/hyperledger-labs/fabex/helpers"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func main() {
	// init configs
	bootConf, err := config.GetBootConfig()
	if err != nil {
		panic(err)
	}
	l, err := log.GetLogger(bootConf)
	if err != nil {
		panic(err)
	}
	defer l.Sync()

	ctx := context.WithValue(context.Background(), "log", l)
	ctx, cancel := context.WithCancel(ctx)

	conf, err := config.GetMainConfig(bootConf)
	if err != nil {
		l.Panic(err.Error())
	}

	// create sdk instance
	sdk, err := fabsdk.New(fabconfig.FromFile(conf.Fabric.ConnectionProfile))
	if err != nil {
		l.Error("failed to create new SDK", zap.Error(err))
	}
	defer sdk.Close()

	if bootConf.Enrolluser {
		err = helpers.EnrollUser(sdk, conf.Fabric.User, conf.Fabric.Secret)
		if err != nil {
			l.Panic("failed to enroll user", zap.Error(err))
		}
	}

	// choose database
	var dbInstance db.Storage
	switch bootConf.Database {
	case "mongo":
		dbInstance = db.CreateDBConfMongo(conf.Mongo.Host, conf.Mongo.Port, conf.Mongo.Dbuser, conf.Mongo.Dbsecret, conf.Mongo.Dbname, conf.Mongo.Collection)
	case "cassandra":
		dbInstance = db.NewCassandraClient(conf.Cassandra.Host, conf.Cassandra.Dbuser, conf.Cassandra.Dbsecret, conf.Cassandra.Keyspace, conf.Cassandra.Columnfamily)
	}

	err = dbInstance.Connect()
	if err != nil {
		l.Panic("DB connection failed", zap.Error(err))
	}
	l.Info("Connected to database successfully")

	// engines for channels
	ecr := engineCreator(sdk, dbInstance)
	var wg sync.WaitGroup
	for _, ch := range conf.Fabric.Channels {
		if err := dbInstance.Init(ch); err != nil {
			l.Error("engine error", zap.Error(err), zap.String("channel", ch))
			continue
		}

		engine, err := ecr(ch, conf.Fabric.User, conf.Fabric.Org)
		if err != nil {
			l.Error(err.Error())
		}

		wg.Add(1)
		go func(ch string, wg *sync.WaitGroup) {
			defer wg.Done()
			if err := engine.Run(ctx); err != nil {
				l.Error("engine error", zap.Error(err), zap.String("channel", ch))
			}
		}(ch, &wg)
	}

	l.Info("start REST server")
	go func() {
		l.Panic("REST server error", zap.Error(rest.Run(dbInstance, conf.UI.Host, conf.UI.Port, bootConf.UI)))
	}()
	l.Info(fmt.Sprintf("REST server started on %s", net.JoinHostPort(conf.UI.Host, conf.UI.Port)))

	// grpc server
	l.Info("start GRPC server")
	go func() {
		serv := grpc.NewFabexServer(conf.GRPCServer.Host, conf.GRPCServer.Port, dbInstance)
		l.Panic("GRPC server error", zap.Error(grpc.StartGrpcServ(ctx, serv)))
	}()
	l.Info(fmt.Sprintf("GRPC server started on %s", net.JoinHostPort(conf.GRPCServer.Host, conf.GRPCServer.Port)))

	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, os.Interrupt, syscall.SIGTERM, os.Interrupt)
	s := <-interruptCh
	l.Info("os signal received, shutdown", zap.String("signal", s.String()))
	cancel()
	wg.Wait()
}
