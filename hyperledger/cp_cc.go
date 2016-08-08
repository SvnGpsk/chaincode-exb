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
//					be done to the product and its business parts at points in its lifecycle
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


//==============================================================================================================================
//	Product 	- Defines the structure for a product passport object.
//	Contract	- Defines the structure for a sales contract, regarding the Product.
//	PPP		- Defines the structure for a Payment and Property Plan (PPP) regarding the Contract and the Product. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================

type Product struct {
	ProductID        string 	`json:pid`
	CheckID          string      	`json:checksum`
	Manufacturer     string       	`json:manufacturer`
	Owner            string       	`json:owner`
	Current_location string        	`json:current_location`
	State            int           	`json:state`
	Width            float32       	`json:width`
	Height           float32       	`json:height`
	Weight           float32 	`json:weight`
	//Contract
}

type Contract struct {
	Seller      string		`json:seller`
	Buyer       string              `json:buyer`
	Buyer_Bank  string              `json:buyerbank`
	Seller_Bank string              `json:sellerbank`
	Price       float32             `json:price`
	Currency    string              `json:currency`
	Origin      string              `json:origin`
	Destination string              `json:destination`
	Route       string              `json:route`
	//Product
	//PPP
}

type PPP struct {
	State         int             	`json:state`
	Property_Plan []string        	`json:sellerbank`
	Payment_Plan  []string        	`json:sellerbank`
}

type ProductId struct {
	Pid 	string	`json:pid`
}

//==============================================================================================================================
//	ProductID Holder - Defines the structure that holds all the ProductIDs for products that have been created.
//				Used as an index when querying all products.
//==============================================================================================================================
type ProductID_Holder struct {
	ProductIDs []string `json:productIds`
}

//==============================================================================================================================
//	ECertResponse - Struct for storing the JSON response of retrieving an ECert. JSON OK -> Struct OK
//==============================================================================================================================
type ECertResponse struct {
	OK    string `json:OK`
	Error string `json:Error`
}

func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	var ProductIds ProductID_Holder

	bytes, err := json.Marshal(ProductIds)

	if err != nil {
		return nil, errors.New("Error creating Product_Id_Holder record")
	}

	err = stub.PutState("productIds", bytes)

	err = stub.PutState("Peer_Address", []byte(args[0]))

	if err != nil {
		return nil, errors.New("Error storing peer address")
	}

	fmt.Println("EXB: Initialization complete")

	return nil, nil
}
//==============================================================================================================================
// createRandomId - Creates a random id for the product
//
//==============================================================================================================================

func (t *SimpleChaincode) createRandomId(stub *shim.ChaincodeStub) (string, error) {
	var randomId = 0
	var low = 100000000
	var high = 999999999
	for {
		randomId = rand.Intn(high - low) + low
		used, err :=t.isRandomIdUnused(stub, strconv.Itoa(randomId))
		if err != nil {
			fmt.Printf("isRandomIdUnused failed %s", err)
			return "-1", errors.New("isRandomIdUnused: Error retrieving vehicle with pid = ")

		}
		if (used) {
			break
		}
	}
	//TODO in createProduct() die ID zur ID-Liste hinzufügen

	return strconv.Itoa(randomId), nil
}

//==============================================================================================================================
// isRandomIdUnused - Checks if the randomly created id is already used by another product.
//
//==============================================================================================================================
func (t *SimpleChaincode) isRandomIdUnused(stub *shim.ChaincodeStub, randomId string) (bool, error) {
	usedIds := make([]string, 500000000)
	var err error
	usedIds, err = t.getAllUsedProductIds(stub)
	if err != nil {
		fmt.Printf("getAllUsedProductIds failed to return used ids", err)
		return true, errors.New("getAllUsedProductIds: Error retrieving product with pid = ")

	}
	for _, id := range usedIds {
		if (id == randomId) {
			return false, nil
		}
	}

	return true, nil
}

//==============================================================================================================================
//	 getProduct - Gets the state of the data at v5cID in the ledger then converts it from the stored
//					JSON into the Vehicle struct for use in the contract. Returns the Vehcile struct.
//					Returns empty v if it errors.
//==============================================================================================================================
func (t *SimpleChaincode) getProduct(stub *shim.ChaincodeStub, productId string) (Product, error) {

	var product Product

	bytes, err := stub.GetState(productId);

	if err != nil {
		fmt.Printf("RETRIEVE_PRODUCT: Failed to invoke chaincode: %s", err);
		return product, errors.New("getProduct: Error retrieving product with pid = "+productId)
	}

	err = json.Unmarshal(bytes, &product);

	if err != nil {
		fmt.Printf("RETRIEVE_PRODUCT: Corrupt product record " + string(bytes) + ": %s", err);
		return product, errors.New("RETRIEVE_PRODUCT: Corrupt product record" + string(bytes))
	}

	return product, nil
}

//==============================================================================================================================
// isRandomIdUnused - Checks if the randomly created id is already used by another product. TODO Check comment
//
//==============================================================================================================================
func (t *SimpleChaincode) getAllUsedProductIds(stub *shim.ChaincodeStub) ([]string, error) {

	usedIds := make([]string, 500000000)

	bytes, err := stub.GetState("productIds")
	fmt.Println("EXB: Bytes of productIdList contain: ", bytes)
	if err != nil {
		return nil, errors.New("Unable to get productIdList")
	}

	var productIds ProductID_Holder
	if len(bytes) != 0{
		err = json.Unmarshal(bytes, &productIds)

		if err != nil {
			fmt.Println(err)
			return nil, errors.New("Invalid JSON for productIdList")
		}
	}

	var product Product

	for i, pid := range productIds.ProductIDs {

		product, err = t.getProduct(stub, pid)

		if err != nil {
			return nil, errors.New("Failed to retrieve pid")
		}
		//TODO prüfung productID != nil und nicht leer
		usedIds[i] = product.ProductID

	}

	return usedIds, nil
}

func (t *SimpleChaincode) init_product(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	var product Product

	fmt.Println("EXB:", args)

	var err error
	err = json.Unmarshal([]byte(args[0]), &product)
	if err != nil {
		fmt.Println("EXB: error unmarshaling test")
		return nil, errors.New("EXB: error unmarshaling test")
	}
	fmt.Println("EXB:", product)

	product.ProductID, err = t.createRandomId(stub)
	product.State = 0
	str, err := json.Marshal(&product)
	fmt.Println("EXB: ", product.ProductID)
	fmt.Println("DEBUG EXB:", []byte(str))
	fmt.Println("DEBUG EXB:", product.ProductID)
	fmt.Println("EXB: ", product.Manufacturer)

	err = stub.PutState(product.ProductID, []byte(str))

	if err != nil {
		fmt.Println("EXB: Error writing product")
		return nil, errors.New("EXB: Error writing the test back")
	}

	bytes, err := stub.GetState("productIds")

	fmt.Println("productIds nach GetState:", bytes)

	if err != nil {
		return nil, errors.New("Unable to get productIds")
	}
	fmt.Println("ZEILE 295")
	var productIds ProductID_Holder
	fmt.Println("ZEILE 297")
	if len(bytes)>0{
		err = json.Unmarshal(bytes, &productIds)
	}
	fmt.Println("ZEILE 299")
	if err != nil {
		return nil, errors.New("Corrupt ProductID_Holder record")
	}
	fmt.Println("ZEILE 303")
	productIds.ProductIDs = append(productIds.ProductIDs, product.ProductID)
	fmt.Println("ZEILE 305")
	fmt.Println("NONONONONO:",productIds.ProductIDs)
	fmt.Println("ZEILE 307")
	bytes, err = json.Marshal(productIds)

	fmt.Println("json marshal:", bytes)

	if err != nil {
		fmt.Print("Error creating ProductID_Holder record")
	}

	err = stub.PutState("productIds", bytes)

	if err != nil {
		return nil, errors.New("Unable to put the state")
	}

	return nil, nil
}

// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read_id(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	var jsonResp string
	var err error
	var productId ProductId
	fmt.Println(args)
	err = json.Unmarshal([]byte(args[0]), &productId)

	fmt.Println(productId.Pid)

	productAsBytes, err := stub.GetState(productId.Pid)                                                                       //get the var from chaincode state
	fmt.Println("productAsBytes=", productAsBytes)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for id\"}"
		return nil, errors.New(jsonResp)
	}
	return productAsBytes, nil                                                                                                        //send it onward
}

// ============================================================================================================================
// ReadAll - read all products from the list inside chaincode state
// ============================================================================================================================
//func (t *SimpleChaincode) read_all(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
//
//	var jsonResp string
//	var err error
//	var productIdList ProductID_Holder
//	fmt.Println(args)
//	err = json.Unmarshal([]byte(args[0]), &productIdList)
//
//	fmt.Println(productId.Pid)
//
//	productAsBytes, err := stub.GetState(productId.Pid)                                                                       //get the var from chaincode state
//	fmt.Println("productAsBytes=", productAsBytes)
//	if err != nil {
//		jsonResp = "{\"Error\":\"Failed to get state for id\"}"
//		return nil, errors.New(jsonResp)
//	}
//	return productAsBytes, nil                                                                                                        //send it onward
//}

func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	//need one arg

	//TODO Produkt IMMER holen und in read_id mit rein geben!
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read_id" {
		//read a variable
		return t.read_id(stub, args)
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
