/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/partha6369/pharmaceutical-chaincode-go/chaincode"
)

func main() {
	pharmaceuticalChaincode, err := contractapi.NewChaincode(
		&chaincode.SmartContract{},
	)
	if err != nil {
		log.Panicf(
			"Error creating pharmaceutical supply-chain chaincode: %v",
			err,
		)
	}

	if err := pharmaceuticalChaincode.Start(); err != nil {
		log.Panicf(
			"Error starting pharmaceutical supply-chain chaincode: %v",
			err,
		)
	}
}
