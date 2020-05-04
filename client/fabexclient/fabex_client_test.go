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

package fabexclient

import (
	"github.com/vadiminshakov/fabex/models"
	pb "github.com/vadiminshakov/fabex/proto"
	"log"
	"bytes"
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

func ExecuteCMD(command string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = "../../"
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	go func(cmd *exec.Cmd) {
		err := cmd.Run()
		if err != nil {
			log.Fatalf(err.Error() + ": " + stderr.String())
		}
	}(cmd)
	return cmd, nil
}

func TestMain(m *testing.M) {

	// start MongoDB
	_, err := ExecuteCMD("make", "mongo")
	if err != nil {
		log.Fatal(err)
	}

	// start test network
	log.Println("Test setup")
	_, err = ExecuteCMD("make", "start-fabric")
	if err != nil {
		log.Fatal(err)
	}
	// wait until containers start, channel and chaincode created
	time.Sleep(105 * time.Second)

	log.Println("Start Fabex")
	cancelCh := make(chan bool)

	go func(cancelCh chan bool) {

		cmd, err := ExecuteCMD("make", "start-fabex")
		if err != nil {
			log.Fatal(err)
		}
		<-cancelCh

		if cmd != nil && cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil {
				log.Fatal("failed to kill Fabex process: ", err)
			}
		}
	}(cancelCh)
	time.Sleep(15 * time.Second)

	log.Println("Run tests")
	code := m.Run()
	cancelCh <- true
	log.Println("Clean test artifacts")

	// purge
	_, err = ExecuteCMD("make", "stop-fabric")
	_, err = ExecuteCMD("make", "stop-mongo")
	if err != nil {
		log.Fatal(err)
	}

	//wait for containers to be removed
	time.Sleep(10 * time.Second)

	os.Exit(code)
}

func TestNew(t *testing.T) {
	_, err := New("localhost", "6000")
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestExplore(t *testing.T) {
	fabcli, err := New("localhost", "6000")
	if err != nil {
		t.Errorf(err.Error())
	}
	err = fabcli.Explore(0, 3)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestGetByBlocknum(t *testing.T) {
	fabcli, err := New("localhost", "6000")
	if err != nil {
		t.Errorf(err.Error())
	}
	txs, err := fabcli.GetByBlocknum(&pb.Entry{Blocknum: 1})
	if err != nil {
		t.Errorf(err.Error())
	}
	if len(txs) == 0 {
		t.Errorf("no transactions found")
	}
}
