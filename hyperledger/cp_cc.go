package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"strconv"
	"bytes"
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
//	User		- Defines a user with his name and affiliation/role.
//	PPP		- Defines the structure for a Payment and Property Plan (PPP) regarding the Contract and the Product.
// 	ProductId	- Defines a struct for storing the ProductId
// 	JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================

type Product struct {
	ProductID        string        `json:pid`
	CheckID          string        `json:checksum`
	Manufacturer     string        `json:manufacturer`
	Owner            User        	`json:owner`
	Current_location string                `json:current_location`
	State            int                `json:state`
	Width            float32        `json:width`
	Height           float32        `json:height`
	Weight           float32        `json:weight`
	//state
	//Contract

}

type Contract struct {
	Seller      string              `json:seller`
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

type User struct {
	Role string        `json:role`
	Name string        `json:name`
}

type PPP struct {
	State         int                `json:state`
	Property_Plan []string                `json:sellerbank`
	Payment_Plan  []string                `json:sellerbank`
}

type ProductId struct {
	Pid string        `json:pid`
}

//==============================================================================================================================
//	ProductID Holder - Defines the structure that holds all the ProductIDs for products that have been created.
//				Used as an index when querying all products.
//==============================================================================================================================
type ProductID_Holder struct {
	ProductIDs []string `json:productIds`
}

//==============================================================================================================================
//	Init - Inits the blockchains and the peers.
//==============================================================================================================================
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
//	 Helping Functions
//==============================================================================================================================
// 	 createRandomId - Creates a random id for the product
//
//==============================================================================================================================

func (t *SimpleChaincode) createRandomId(stub *shim.ChaincodeStub) (string, error) {
	var randomId = 0
	var low = 100000000
	var high = 999999999
	for {
		randomId = rand.Intn(high - low) + low
		used, err := t.isRandomIdUnused(stub, strconv.Itoa(randomId))
		if err != nil {
			fmt.Printf("isRandomIdUnused failed %s", err)
			return "-1", errors.New("isRandomIdUnused: Error retrieving vehicle with pid = ")

		}
		if (used) {
			break
		}
	}

	return strconv.Itoa(randomId), nil
}

//==============================================================================================================================
// 	isRandomIdUnused - Checks if the randomly created id is already used by another product.
//
//==============================================================================================================================
func (t *SimpleChaincode) isRandomIdUnused(stub *shim.ChaincodeStub, randomId string) (bool, error) {
	usedIds := make([]string, 500000000)
	var err error
	usedIds, err = t.getAllUsedProductIds(stub)
	if err != nil {
		fmt.Println("getAllUsedProductIds failed to return used ids", err)
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
		fmt.Printf("getProduct: Failed to invoke chaincode: %s", err);
		return product, errors.New("getProduct: Error retrieving product with pid = " + productId)
	}

	err = json.Unmarshal(bytes, &product);

	if err != nil {
		fmt.Printf("RETRIEVE_PRODUCT: Corrupt product record " + string(bytes) + ": %s", err);
		return product, errors.New("RETRIEVE_PRODUCT: Corrupt product record" + string(bytes))
	}

	return product, nil
}

//==============================================================================================================================
// 	getAllUsedProductIds - Returns a list of all product IDs that are already in use
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
	if len(bytes) != 0 {
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
		usedIds[i] = product.ProductID
	}

	return usedIds, nil
}

// ============================================================================================================================
// 	Read - read a variable from chaincode state
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
	n := bytes.Index(productAsBytes, []byte{0})
	fmt.Println("PRODUCT",n)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for id\"}"
		return nil, errors.New(jsonResp)
	}
	return productAsBytes, nil                                                                                                        //send it onward
}

//============================================================================================================================
//	 ReadAll - read all products from the list inside chaincode state
//============================================================================================================================
func (t *SimpleChaincode) read_all(stub *shim.ChaincodeStub) ([]byte, error) {

	var jsonResp string
	var err error
	var productIdList ProductID_Holder

	productListAsBytes, err := stub.GetState("productIds")

	fmt.Println("productListAsBytes=", productListAsBytes)

	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for id\"}"
		return nil, errors.New(jsonResp)
	}

	fmt.Println("productList=", productIdList)

	return productListAsBytes, nil                                                                                                        //send it onward
}

//==============================================================================================================================
// 	save_changes - Writes to the ledger the Product struct passed in a JSON format. Uses the shim file's
//				  method 'PutState'.
//==============================================================================================================================
func (t *SimpleChaincode) save_changes(stub *shim.ChaincodeStub, product Product) (bool, error) {

	bytes, err := json.Marshal(product)

	if err != nil {
		fmt.Printf("SAVE_CHANGES: Error converting vehicle record: %s", err); return false, errors.New("Error converting product record")
	}

	err = stub.PutState(product.ProductID, bytes)

	if err != nil {
		fmt.Printf("SAVE_CHANGES: Error storing vehicle record: %s", err); return false, errors.New("Error storing product record")
	}

	return true, nil
}

//==============================================================================================================================
//	 Router Functions
//=================================================================================================================================
//	Query - Called on chaincode query. Takes a function name passed and calls that function. Passes the
//  		initial arguments passed are passed on to the called function.
//=================================================================================================================================

func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	//need one arg

	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read_id" {
		//read a variable
		return t.read_id(stub, args)
	} else if function == "read_all" {
		return t.read_all(stub)
	}
	fmt.Println("query did not find func: " + function)                                                //error

	return nil, errors.New("Received unknown function query")
}

func (t *SimpleChaincode) Run(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("run is running " + function)
	return t.Invoke(stub, function, args)
}

//==============================================================================================================================
//	Invoke - Called on chaincode invoke. Takes a function name passed and calls that function.
//==============================================================================================================================

func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	if function == "create_product" {
		fmt.Println("Writing in Product Blockchain")
		//Create an asset with some value
		return t.create_product(stub, args)
	} else if function == "init" {
		fmt.Println("Firing init")
		return t.Init(stub, "init", args)
	} else {
		fmt.Println(args)
		product, err := t.getProduct(stub, args[1]) //TODO args?
		if err != nil {
			fmt.Printf("getProduct: Error getting product: %s", err);
			return nil, errors.New("Error getting product")
		}
		fmt.Println("GetProduct result: ", product)

		//var caller User
		//var recipient User

		if function == "seller_to_buyer" {
			return nil, nil
			//return t.seller_to_buyer(product)
		} else if function == "seller_to_buyersbank" {
			//return t.seller_to_buyersbank(stub, product, caller, recipient)
		} else if function == "buyersbank_to_buyer" {
			//return t.buyersbank_to_buyer(product)
		}
	}

	return nil, errors.New("Received unknown function invocation")
}
//=================================================================================================================================
//	 Create Functions
//==============================================================================================================================
//	 create_product - Creates a product in the blockchain with arguments.
//=================================================================================================================================

func (t *SimpleChaincode) create_product(stub *shim.ChaincodeStub, args []string) ([]byte, error) {

	var product Product
	var user User
	fmt.Println("EXB:", args[0])
	var err error
	err = json.Unmarshal([]byte(args[0]), &user)
	if err != nil {
		fmt.Println("EXB: error unmarshaling product")
		return nil, errors.New("EXB: error unmarshaling product")
	}
	fmt.Println("EXB USER OBJECT: ", user)
	if user.Role == "2" {
		fmt.Println("EXB:", product)
		product.Owner = user;
		product.Manufacturer = user.Name;
		product.ProductID, err = t.createRandomId(stub)
		product.State = 0
		str, err := json.Marshal(&product)
		fmt.Println("EXB PRODUCT FOR PUT: ", product)
		fmt.Println(str)
		err = stub.PutState(product.ProductID, []byte(str))

		fmt.Println(product.Owner)
		if err != nil {
			fmt.Println("EXB: Error writing product")
			return nil, errors.New("EXB: Error writing the test back")
		}

		bytes, err := stub.GetState("productIds")

		if err != nil {
			return nil, errors.New("Unable to get productIds")
		}
		var productIds ProductID_Holder
		if len(bytes) > 0 {
			err = json.Unmarshal(bytes, &productIds)
		}
		if err != nil {
			return nil, errors.New("Corrupt ProductID_Holder record")
		}

		productIds.ProductIDs = append(productIds.ProductIDs, product.ProductID)
		bytes, err = json.Marshal(productIds)

		fmt.Println("json marshal:", bytes)

		if err != nil {
			fmt.Print("Error creating ProductID_Holder record")
		}

		err = stub.PutState("productIds", bytes)

		if err != nil {
			return nil, errors.New("Unable to put the state")
		}
	}
	return nil, nil
}

//=================================================================================================================================
//	 Transfer Functions
//=================================================================================================================================
//	 seller to buyersbank
//=================================================================================================================================
//func (t *SimpleChaincode) seller_to_buyersbank(stub *shim.ChaincodeStub, product Product, caller User, recipient User) ([]byte, error) {
//
//	//if product.Make == "UNDEFINED" ||
//	//	product.Model == "UNDEFINED" ||
//	//	product.Reg == "UNDEFINED" ||
//	//	product.Colour == "UNDEFINED" ||
//	//	product.VIN == 0 {
//	//	//If any part of the car is undefined it has not bene fully manufactured so cannot be sent
//	//	fmt.Println("MANUFACTURER_TO_PRIVATE: Car not fully defined")
//	//	return nil, errors.New("Car not fully defined")
//	//}
//
//	if product.State == STATE_PAYMENTANDPROPERTYPLANADDED       &&
//		product.Owner == caller.Name                                &&
//		caller.Role == SELLER                        &&
//		recipient.Role == BUYER_BANK{
//
//		product.Owner = recipient.Name
//		product.State = STATE_PRODUCTBEINGSHIPPED
//
//	} else {
//		return nil, errors.New("Permission denied")
//	}
//
//	_, err := t.save_changes(stub, product)
//
//	if err != nil {
//		fmt.Printf("MANUFACTURER_TO_PRIVATE: Error saving changes: %s", err); return nil, errors.New("Error saving changes")
//	}
//
//	return nil, nil
//
//}

////=================================================================================================================================
////	 private_to_private
////=================================================================================================================================
//func (t *SimpleChaincode) seller_to_buyer(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {
//
//	if v.Status == STATE_PRIVATE_OWNERSHIP        &&
//		v.Owner == caller                                        &&
//		caller_affiliation == PRIVATE_ENTITY                        &&
//		recipient_affiliation == PRIVATE_ENTITY                        &&
//		v.Scrapped == false {
//
//		v.Owner = recipient_name
//
//	} else {
//
//		return nil, errors.New("Permission denied")
//
//	}
//
//	_, err := t.save_changes(stub, v)
//
//	if err != nil {
//		fmt.Printf("PRIVATE_TO_PRIVATE: Error saving changes: %s", err); return nil, errors.New("Error saving changes")
//	}
//
//	return nil, nil
//
//}

////=================================================================================================================================
////	 private_to_lease_company
////=================================================================================================================================
//func (t *SimpleChaincode) buyersbank_to_buyer(stub *shim.ChaincodeStub, v Vehicle, caller string, caller_affiliation int, recipient_name string, recipient_affiliation int) ([]byte, error) {
//
//	if v.Status == STATE_PRIVATE_OWNERSHIP        &&
//		v.Owner == caller                                        &&
//		caller_affiliation == PRIVATE_ENTITY                        &&
//		recipient_affiliation == LEASE_COMPANY                        &&
//		v.Scrapped == false {
//
//		v.Owner = recipient_name
//
//	} else {
//		return nil, errors.New("Permission denied")
//	}
//
//	_, err := t.save_changes(stub, v)
//	if err != nil {
//		fmt.Printf("PRIVATE_TO_LEASE_COMPANY: Error saving changes: %s", err); return nil, errors.New("Error saving changes")
//	}
//
//	return nil, nil
//
//}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Println("Error starting Simple chaincode:", err)
	}
}
