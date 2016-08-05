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
	"strconv"
)

//==============================================================================================================================
//	 Participant types - Each participant type is mapped to an integer which we use to compare to the value stored in a
//						 user's eCert
//==============================================================================================================================
const GOVERNMENT = 1
const SELLER = 2
const BUYER = 3
const SELLER_BANK = 4
const BUYER_BANK = 5
const SHIPPER = 6
const PRODUCT = 7


//==============================================================================================================================
//	 Status types - Asset lifecycle is broken down into 8 statuses, this is part of the business logic to determine what can
//					be done to the product and its busines parts at points in its lifecycle
//==============================================================================================================================
const STATE_PRODUCTPASSPORTADDED = 0
const STATE_CONTRACTADDED = 1
const STATE_PAYMENTANDPROPERTYPLANADDED = 2
const STATE_LETTEROFCREDITACCEPTED = 3
const STATE_PRODUCTPASSPORTCOMPLETE = 4
const STATE_PRODUCTBEINGSHIPPED = 5
const STATE_PRODUCTINUSE = 6
const STATE_MAINTENANCENEEDED = 7


// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type Tid struct {
	Tid  string        `json:tid`
}

type Test struct {
	Name string        `json:name`
	Tid  string        `json:tid`
}
//==============================================================================================================================
//	Product 	- Defines the structure for a product passport object.
//	Contract	- Defines the structure for a sales contract, regarding the Product.
//	PPP		- Defines the structure for a Payment and Property Plan (PPP) regarding the Contract and the Product. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
//noinspection GoStructTag
type Product struct {
	ProductID        string 	`json:pid`
	CheckID          string 	`json:checksum`
	Manufacturer     string 	`json:manufacturer`
	Owner            string 	`json:owner`
	Current_location string 	`json:current_location`
	State            int 		`json:state`
	Width            float32 	`json:width`
	Height           float32 	`json:height`
	Weight           float32 	`json:weight`
	//Contract
}

type Contract struct {
	Seller      string 		`json:seller`
	Buyer       string 		`json:buyer`
	Buyer_Bank  string 		`json:buyerbank`
	Seller_Bank string		`json:sellerbank`
	Price       float32 		`json:price`
	Currency    string 		`json:currency`
	Origin      string 		`json:origin`
	Destination string 		`json:destination`
	Route       string 		`json:route`
	//Product
	//PPP
}

type PPP struct {
	State		int 		`json:state`
	Property_Plan	[]string 	`json:sellerbank`
	Payment_Plan	[]string 	`json:sellerbank`
}

func GetTest(incomingtest string, stub *shim.ChaincodeStub) (Test, error) {
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
	id = rand.Intn(10000)
	return id
}

func (t *SimpleChaincode) init_product(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	var test Test

	fmt.Println("EXB:", args)

	var err error
	err = json.Unmarshal([]byte(args[0]), &test)
	if err != nil {
		fmt.Println("EXB: error unmarshaling test")
		return nil, errors.New("EXB: error unmarshaling test")
	}
	fmt.Println("EXB:", test)


	//str := `{"name": "` + test.Name + `", "id": "` + test.Tid + `"}`
	test.Tid = strconv.Itoa(t.GetRandomId())
	str, err := json.Marshal(&test)
	fmt.Println("EXB: ", test.Tid)
	err = stub.PutState(test.Tid, str)
	if err != nil {
		fmt.Println("EXB: Error writing test")
		return nil, errors.New("EXB: Error writing the test back")
	}
	return nil, nil
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read_all(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	var jsonResp string
	var err error
	var tid Tid

	err = json.Unmarshal([]byte(args[0]), &tid)
	//var test Test

	fmt.Println(args)
	testObjAsbytes, err := stub.GetState(tid.Tid)                                                                       //get the var from chaincode state
	fmt.Println("testObjAsbytes=",testObjAsbytes)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for id\"}"
		return nil, errors.New(jsonResp)
	}
	//err = json.Unmarshal(testObjAsbytes, &test);
	return testObjAsbytes, nil                                                                                                        //send it onward
}

func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	//need one arg
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read_all" {
		//read a variable
		return t.read_all(stub, args)
	}
	fmt.Println("query did not find func: " + function)                                                //error

	return nil, errors.New("Received unknown function query")
}

func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	if function == "init_product" {
		fmt.Println("Writing in Blockchain")
		//Create an asset with some value
		return t.init_product(stub, args)
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
