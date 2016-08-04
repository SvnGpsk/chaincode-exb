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
    	fmt.Println("Initialization complete")
	return nil, nil
}

func (t *SimpleChaincode) GetRandomId() int {
	var id = 0
	id = rand.Intn(100)
	id=10
	return id
}

func (t *SimpleChaincode) write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	if len(args) != 2 {
		fmt.Println("error invalid arguments")
		return nil, errors.New("Incorrect number of arguments. Expecting Test object")
	}

	var err error
	var id = args[1]
	str := `{"name": "` + args[0] + `", "id": "` + id + `"}`
	fmt.Println("Unmarshalling Test")

	err = stub.PutState(id, []byte(str))
	if err != nil {
		fmt.Println("Error writting test back")
		return nil, errors.New("Error writing the test back")
	}
	return nil, nil
}


func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	//need one arg
	if len(args) < 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting ......")
	}

	if args[0] == "GetTest" {

		fmt.Println("Getting particular test")
		test, err := GetTest(args[1], stub)
		if err != nil {
			fmt.Println("Error Getting particular test")
			return nil, err
		} else {
			testBytes, err1 := json.Marshal(&test)
			if err1 != nil {
				fmt.Println("Error marshalling the test")
				return nil, err1
			}
			fmt.Println("All success, returning the test")
			return testBytes, nil
		}

	}
	return nil, nil
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
		return t.write(stub, args)
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
