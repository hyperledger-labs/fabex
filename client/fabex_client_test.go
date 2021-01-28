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

package client

import (
	"bytes"
	"github.com/hyperledger-labs/fabex/models"
	pb "github.com/hyperledger-labs/fabex/proto"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"os/exec"
	"path"
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
	dir, _ := os.Getwd()
	cmd.Dir = path.Join(dir, "../")
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
	_, err := ExecuteCMD("make", "mongo-test")
	if err != nil {
		log.Fatal(err)
	}

	// start test network
	log.Println("Test setup")
	_, err = ExecuteCMD("make", "fabric-test")
	if err != nil {
		log.Fatal(err)
	}
	// wait until containers start, channel and chaincode created
	time.Sleep(135 * time.Second)

	log.Println("Start Fabex")
	cancelCh := make(chan bool)

	go func(cancelCh chan bool) {
		cmd, err := ExecuteCMD("make", "fabex-test-integration")
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
	time.Sleep(55 * time.Second)

	log.Println("Run tests")
	code := m.Run()
	cancelCh <- true
	log.Println("Clean test artifacts")

	// purge
	_, err = ExecuteCMD("make", "stop-fabric-test")
	_, err = ExecuteCMD("make", "stop-mongo-test")
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

func TestGetRange(t *testing.T) {
	fabcli, err := New("localhost", "6000")
	if err != nil {
		t.Errorf(err.Error())
	}
	txs, err := fabcli.GetRange(0, 3)
	if err != nil {
		t.Errorf(err.Error())
	}
	assert.Greater(t, len(txs), 0, "No transactions found")
}

func TestGet(t *testing.T) {
	fabcli, err := New("localhost", "6000")
	if err != nil {
		t.Errorf(err.Error())
	}

	txs, err := fabcli.Get(&pb.Entry{Blocknum: 1})
	if err != nil {
		t.Errorf(err.Error())
	}
	assert.Greater(t, len(txs), 0, "No transactions found")
	for _, tx := range txs {
		assert.EqualValuesf(t, tx.Blocknum, 1, "Not valid tx retrieved, got %d, want %d", tx.Blocknum, 1)
	}

	txs, err = fabcli.Get(&pb.Entry{Txid: txs[0].Txid})
	if err != nil {
		t.Errorf(err.Error())
	}
	assert.Greater(t, len(txs), 0, "No transactions found")
}

func TestGetAllAndCheckValidationCode(t *testing.T) {
	fabcli, err := New("localhost", "6000")
	if err != nil {
		t.Errorf(err.Error())
	}
	txs, err := fabcli.Get(nil)
	if err != nil {
		t.Errorf(err.Error())
	}

	for _, tx := range txs {
		assert.Equal(t, tx.ValidationCode, int32(0), "validation code of tx %s is %d (invalid)", tx.Txid, tx.ValidationCode)
	}
}
