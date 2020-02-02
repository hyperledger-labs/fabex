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
	"fmt"
	"github.com/ory/dockertest"
	"github.com/vadiminshakov/fabex/models"
	"log"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

var (
	fabex *models.Fabex
	wg    *sync.WaitGroup
)

func ExecuteCMD(command string, args ...string) {
	cmd := exec.Command(command, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n%s", err, string(out))
	}
	fmt.Printf("combined out:\n%s\n", string(out))
}

func TestMain(m *testing.M) {

	// start MongoDB
	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(fmt.Sprintf("Could not connect to docker: %s", err))
	}

	resource, err := pool.Run("mongo", "latest", []string{"name=mongodbtest", "p=27017:27017"})
	if err != nil {
		panic(fmt.Sprintf("Could not start resource: %s", err))
	}

	// start first network
	go ExecuteCMD("./fabex", "--task", "grpc", "--configpath", "./", "--configname", "config")
	time.Sleep(3 * time.Second)

	code := m.Run()

	// purge
	if err := pool.Purge(resource); err != nil {
		panic(fmt.Sprintf("Could not purge resource: %s", err))
	}

	os.Exit(code)
}

func TestExplore(t *testing.T) {
	go ExecuteCMD("./client/client")
}
