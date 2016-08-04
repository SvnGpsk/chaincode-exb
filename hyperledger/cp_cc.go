/*
Copyright 2016 IBM

Licensed under the Apache License, Version 2.0 (the "License")
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

Licensed Materials - Property of IBM
© Copyright IBM Corp. 2016
*/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"github.com/hyperledger/fabric/core/chaincode/shim"

)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type Test struct {
	Name    string	`json:"name"`
	Id 	int	`json:"id"`
}

func GetTest(incomingtest string, stub *shim.ChaincodeStub) (Test, error){
	var test Test

	testBytes, err := stub.GetState(incomingtest)
	if err != nil {
		fmt.Println("Error retrieving test " + incomingtest)
		return test, errors.New("Error retrieving test " + incomingtest)
	}

	err = json.Unmarshal(testBytes, &test)
	if err != nil {
		fmt.Println("Error unmarshalling test " + incomingtest)
		return test, errors.New("Error unmarshalling test " + incomingtest)
	}

	return test, nil
}


func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
    	fmt.Println("EXB: Initialization complete")
	return nil, nil
}

func (t *SimpleChaincode) GetRandomId() int {
	var id = 0
	id = rand.Intn(100)
	id=10
	return id
}

func (t *SimpleChaincode) initProduct(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	if len(args) <= 1 {
		fmt.Println("EXB: error invalid arguments")
		return nil, errors.New("EXB: Incorrect number of arguments. Expecting Test object")
	}

	var err error
	str := `{"name": "` + args[1] + `", "id": "` + args[0] + `"}`
	fmt.Println("EXB: Unmarshalling Test")

	err = stub.PutState(args[0], []byte(str))
	if err != nil {
		fmt.Println("EXB: Error writing test")
		return nil, errors.New("EXB: Error writing the test back")
	}
	return nil, nil
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var id, jsonResp string
	var err error

	if len(args) < 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the var to query")
	}

	id = args[0]
	valAsbytes, err := stub.GetState(id)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + id + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil													//send it onward
}

func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	//need one arg
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" {													//read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query")
}

func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	if function == "write" {
		fmt.Println("Writing in Blockchain")
		//Create an asset with some value
		return t.initProduct(stub, args)
	} else if function == "init" {
		fmt.Println("Firing init")
		return t.Init(stub, "init", args)
	}

	return nil, errors.New("Received unknown function invocation")
}

//noinspection ALL
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Println("Error starting Simple chaincode: %s", err)
	}
}
