package helpers

import (
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/ledger"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/msp"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/logging"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func QueryInstalledCC(sdk *fabsdk.FabricSDK, user string) {
	userContext := sdk.Context(fabsdk.WithUser(user))

	resClient, err := resmgmt.New(userContext)
	if err != nil {
		fmt.Println("Failed to create resmgmt: ", err)
	}

	resp2, err := resClient.QueryInstalledChaincodes()
	if err != nil {
		fmt.Println("Failed to query installed cc: ", err)
	}
	fmt.Println("Installed cc: ", resp2.GetChaincodes())
}

func QueryCC(client *channel.Client, name []byte, chaincodeId string) string {
	var queryArgs = [][]byte{name}
	response, err := client.Query(channel.Request{
		ChaincodeID: chaincodeId,
		Fcn:         "query",
		Args:        queryArgs,
	}, channel.WithTargetEndpoints("grpcs://localhost:7051"))

	if err != nil {
		fmt.Println("Failed to query: ", err)
	}

	ret := string(response.Payload)
	fmt.Println("Chaincode status: ", response.ChaincodeStatus)
	fmt.Println("Payload: ", ret)
	return ret
}

func InvokeCC(client *channel.Client, chaincodeId string, args ...string) {
	fmt.Println("Invoke cc with new value:", args)
	invokeArgs := [][]byte{[]byte(args[0]), []byte(args[1]), []byte(args[2])}

	_, err := client.Execute(channel.Request{
		ChaincodeID: chaincodeId,
		Fcn:         "invoke",
		Args:        invokeArgs,
	})

	if err != nil {
		fmt.Printf("Failed to invoke: %+v\n", err)
	}
}

func EnrollUser(sdk *fabsdk.FabricSDK, user, secret string) {
	ctx := sdk.Context()
	mspClient, err := msp.New(ctx)
	if err != nil {
		fmt.Printf("Failed to create msp client: %s\n", err)
	}

	_, err = mspClient.GetSigningIdentity(user)
	if err == msp.ErrUserNotFound {
		fmt.Println("Going to enroll user")
		err = mspClient.Enroll(user, msp.WithSecret(secret))

		if err != nil {
			fmt.Printf("Failed to enroll user: %s\n", err)
		} else {
			fmt.Printf("Success enroll user: %s\n", user)
		}

	} else if err != nil {
		fmt.Printf("Failed to get user: %s\n", err)
	} else {
		fmt.Printf("User %s already enrolled, skip enrollment.\n", user)
	}
}

func QueryChannelConfig(ledgerClient *ledger.Client) {
	resp1, err := ledgerClient.QueryConfig()
	if err != nil {
		fmt.Printf("Failed to queryConfig: %s", err)
	}
	fmt.Println("ChannelID: ", resp1.ID())
	fmt.Println("Channel Orderers: ", resp1.Orderers())
	fmt.Println("Channel Versions: ", resp1.Versions())
}

func QueryChannelInfo(ledgerClient *ledger.Client) (*fab.BlockchainInfoResponse, error) {
	resp, err := ledgerClient.QueryInfo()
	if err != nil {
		fmt.Printf("Failed to queryInfo: %s", err)
		return nil, err
	}
	return resp, nil
}

func SetupLogLevel(lvl logging.Level) {
	logging.SetLevel("fabsdk", lvl)
	logging.SetLevel("fabsdk/common", lvl)
	logging.SetLevel("fabsdk/fab", lvl)
	logging.SetLevel("fabsdk/client", lvl)
}
