/*
Copyright IBM Corp. 2016 All Rights Reserved.

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

//WARNING - this chaincode's ID is hard-coded in chaincode_example04 to illustrate one way of
//calling chaincode from a chaincode. If this example is modified, chaincode_example04.go has
//to be modified as well with the new ID of chaincode_example02.
//chaincode_example05 show's how chaincode ID can be passed in as a parameter instead of
//hard-coding.

import (
	"errors"
	"fmt"
	//"strconv"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
)

//This a very easy chaincode to just test if something is being written in the local blockchain.
//It has no more features than cp_cc.go but is easier.

// Chaincode for Trade Finance on Blockchain Application
type SimpleChaincode struct {
}

type ProductIDHolder struct {
	ProductIDs []string `json:"productIDs"`
}

type Product struct {
	ProductName      string        `json:"product_name"`
	Width            float32       `json:"width"`
	Height           float32       `json:"height"`
	Weight           float32       `json:"weight"`
	OwnerRole        string        `json:"owner_role"`
	Owner            string        `json:"owner"`
	//Contract
}

//Saves all the changes to the blockchain
func (t *SimpleChaincode) save_changes(stub *shim.ChaincodeStub, p Product) (bool, error) {

	bytes, err := json.Marshal(p)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error converting product record: %s", err); return false, errors.New("Error converting product record") }

	err = stub.PutState(p.ProductName, bytes)

	if err != nil { fmt.Printf("SAVE_CHANGES: Error storing product record: %s", err); return false, errors.New("Error storing product record") }

	return true, nil
}

//Initializing the chaincode and initializing the ProductIDHolder
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	var ProductIDs ProductIDHolder
	bytes, err := json.Marshal(ProductIDs)

	if err != nil {
		return nil, errors.New("Error creating ProductIDHolder record")
	}

	err = stub.PutState("productIDs", bytes)

	fmt.Println("Initialization complete")
	return nil, nil
}

// Invoke Router Function: either creating or updating a product
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if function == "create_product" { return t.create_product(stub, args)
	} else {return nil, nil

		//if function == "update_make"  	    { return t.update_make(stub, args)
		//}

	}
}

func (t *SimpleChaincode) create_product(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var product Product
	var err error

	productName 	:= "\"pid\":\"" + args[0] + "\", "
	width 		:= "\"width\":\"" + args[1] + "\", "
	height 		:= "\"height\":\"" + args[2] + "\", "
	weight 		:= "\"weight\":\"" + args[3] + "\", "
	ownerrole 	:= "\"owner_role\":\"" + args[4] + "\", "
	owner 		:= "\"owner\":\"" + args[5] + "\", "

	product_json := "{"+productName+width+height+weight+ownerrole+owner+"}" 	// Concatenates the variables to create the total JSON object

	err = json.Unmarshal([]byte(product_json), &product)							// Convert the JSON defined above into a product object for go

	if err != nil { return nil, errors.New("Invalid JSON object") }

	record, err := stub.GetState(product.ProductName) 								// If not an error then a record exists so cant create a new product with this ProductName as it must be unique

	if record != nil { return nil, errors.New("Product already exists") }

	_, err  = t.save_changes(stub, product)

	if err != nil { fmt.Printf("CREATE_PRODUCT: Error saving changes: %s", err); return nil, errors.New("Error saving changes") }

	//Put the ProductName into the ProductIDs
	bytes, err := stub.GetState("productIDs")

	if err != nil { return nil, errors.New("Unable to get productIDs") }

	var allproductIDs ProductIDHolder

	err = json.Unmarshal(bytes, &allproductIDs)

	if err != nil {	return nil, errors.New("Corrupt ProductIDHolder record") }

	allproductIDs.ProductIDs = append(allproductIDs.ProductIDs, args[0])

	bytes, err = json.Marshal(allproductIDs)

	if err != nil { fmt.Print("Error creating ProductIDHolder record") }

	err = stub.PutState("productIDs", bytes)

	if err != nil { return nil, errors.New("Unable to put the state") }

	return nil, nil

}

// Query callback representing the query of a chaincode
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if function != "query" {
		return nil, errors.New("Invalid query function name. Expecting \"query\"")
	}
	var proName string // Entities
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the product to query")
	}

	proName = args[0]

	// Get the state from the ledger
	proName, err = stub.GetState(proName)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + proName + "\"}"
		return nil, errors.New(jsonResp)
	}

	if proName == nil {
		jsonResp := "{\"Error\":\"Nil amount for " + proName + "\"}"
		return nil, errors.New(jsonResp)
	}

	jsonResp := "{\"Name\":\"" + proName + "\",\"Product\":\"" + string(proName) + "\"}"
	fmt.Printf("Query Response:%s\n", jsonResp)
	return proName, nil
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}
