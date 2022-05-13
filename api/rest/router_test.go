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

package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperledger-labs/fabex/models"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

type Response struct {
	Error string         `json:"error"`
	Msg   []models.Block `json:"msg"`
}

func ExecuteCMD(command string, args ...string) (*exec.Cmd, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = "../"
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

func TestEndpoints(t *testing.T) {
	t.Run("byblocknum", func(t *testing.T) {
		resp, err := http.Get("http://localhost:5252/byblocknum/1")
		if err != nil {
			t.Errorf(err.Error())
		}
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf(err.Error())
		}
		var txs Response
		err = json.Unmarshal(bodyBytes, &txs)
		if err != nil {
			t.Errorf(err.Error())
		}
		assert.Greater(t, len(txs.Msg), 0, "No transactions found")
		for _, tx := range txs.Msg {
			assert.EqualValuesf(t, tx.Blocknum, 1, "Not valid tx retrieved, got %d, want %d", tx.Blocknum, 1)
		}
	})

	t.Run("bytxid", func(t *testing.T) {
		var TXID string
		// get tx ID
		respbyblocknum, err := http.Get("http://localhost:5252/byblocknum/1")
		if err != nil {
			t.Errorf(err.Error())
		}
		bodyBytesByBlocknum, err := ioutil.ReadAll(respbyblocknum.Body)
		if err != nil {
			t.Errorf(err.Error())
		}
		var txsByBlocknum Response
		err = json.Unmarshal(bodyBytesByBlocknum, &txsByBlocknum)
		if err != nil {
			t.Errorf(err.Error())
		}
		assert.Greater(t, len(txsByBlocknum.Msg), 0, "No transactions found")
		for _, tx := range txsByBlocknum.Msg {
			assert.EqualValuesf(t, tx.Blocknum, 1, "Not valid tx retrieved, got %d, want %d", tx.Blocknum, 1)
			TXID = tx.Txs[0].Txid
		}

		// check tx with this ID
		resp, err := http.Get(fmt.Sprintf("http://localhost:5252/bytxid/%s", TXID))
		if err != nil {
			t.Errorf(err.Error())
		}
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf(err.Error())
		}
		var txs Response
		err = json.Unmarshal(bodyBytes, &txs)
		if err != nil {
			t.Errorf(err.Error())
		}
		assert.Greater(t, len(txs.Msg), 0, "No transactions found")
		for _, tx := range txs.Msg {
			assert.EqualValuesf(t, tx.Txs[0].Txid, TXID, "Not valid tx retrieved, got tx ID %d, want %d", tx.Txs[0].Txid, TXID)
		}
	})

	t.Run("InvalidBlockNumber", func(t *testing.T) {
		resp, err := http.Get("http://localhost:5252/byblocknum/999999999999")
		if err != nil {
			t.Errorf(err.Error())
		}
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf(err.Error())
		}

		var txs Response
		err = json.Unmarshal(bodyBytes, &txs)
		if err != nil {
			t.Errorf(err.Error())
		}
		assert.Equal(t, "no such data", txs.Error, "failed to handle invalid block number")
	})
}
