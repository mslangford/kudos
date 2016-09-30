/*
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

type Account struct {
	ID           string `json:"id"`
	PointBalance int    `json:"pointBalance"`
	Out          int    `json:"given"`
	In           int    `json:"received"`
}

type Transaction struct {
	FromID string `json:"fromID"`
	ToID   string `json:"toID"`
	Points int    `json:"points"`
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	var account Account
	var accounts []Account

	for i := 0; i < len(args); i = i + 2 {
		fmt.Println("setting balance for " + args[i] + " to " + args[i+1])
		points, err := strconv.Atoi(args[i+1])
		if err != nil {
			return nil, err
		}
		account = Account{ID: args[i], PointBalance: points}
		accounts = append(accounts, account)
	}
	accountsBytes, _ := json.Marshal(&accounts)
	err := stub.PutState("accounts", accountsBytes)
	if err != nil {
		fmt.Println("error storing accounts" + account.ID)
		return nil, errors.New("Error storing account " + account.ID)
	}

	return nil, nil
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	// Handle different functions
	fmt.Println("invoke is running " + function)
	if function == "init" {
		return t.Init(stub, "init", args)
		//	} else if function == "write" {
		//		return t.write(stub, args)
	} else if function == "transfer" {
		return t.transfer(stub, args)
	}
	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation")
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "read" { //read a variable
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)

	return nil, errors.New("Received unknown function query")
}

/*
func (t *SimpleChaincode) write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var key, value string
	var err error
	fmt.Println("running write()")
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}
	// write - invoke function to write key/value pair
	key = args[0] //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value)) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	return nil, nil
}
*/

// read - query function to read key/value pair
func (t *SimpleChaincode) read(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var key, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting name of the key to query")
	}

	key = args[0]
	valAsbytes, err := stub.GetState(key)
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + key + "\"}"
		return nil, errors.New(jsonResp)
	}

	return valAsbytes, nil
}

// Transfer points between users
func (t *SimpleChaincode) transfer(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var err error
	//	var fromState, toState []byte
	var fromBal, toBal, points, fromIndex, toIndex int
	var accounts []Account
	var transaction Transaction
	var transactions []Transaction

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting from & to usernames and number of points")
	}

	// get balances from accounts array
	fmt.Println("Retrieving accounts")
	accountsBytes, err := stub.GetState("accounts")
	if err != nil {
		fmt.Println("Error retrieving accounts")
		return nil, errors.New("Error retrieving accounts")
	}
	fmt.Println("Converting accounts")
	err = json.Unmarshal(accountsBytes, &accounts)
	if err != nil {
		fmt.Println("Error converting accounts")
		return nil, errors.New("Error converting accounts")
	}
	fromIndex = -1
	toIndex = -1
	for i := 0; i < len(accounts); i++ {
		if accounts[i].ID == args[0] {
			fromBal = accounts[i].PointBalance
			fromIndex = i
		} else if accounts[i].ID == args[1] {
			toBal = accounts[i].PointBalance
			toIndex = i
		}
	}
	if fromIndex == -1 {
		fmt.Println("Could not retrieve balance for " + args[0])
		return nil, errors.New("Could not retrieve balance for " + args[0])
	} else if toIndex == -1 {
		fmt.Println("Could not retrieve balance for " + args[1])
		return nil, errors.New("Could not retrieve balance for " + args[0])
	}

	//	transfer points
	fmt.Println("convert points")
	points, err = strconv.Atoi(args[2])
	if err != nil {
		fmt.Println("Failed to convert points to integer")
		return nil, errors.New("Failed to convert points to integer")
	}
	if fromBal < points {
		fmt.Println("Point balance does not cover transfer amount for " + args[0])
		return nil, errors.New("Point balance does not cover transfer amount for " + args[0])
	}

	//	apply transfer
	fromBal = fromBal - points
	toBal = toBal + points
	fmt.Println("apply transfer from " + args[0] + " to " + args[1] + " for " + args[3] + " points - new points " + strconv.Itoa(fromBal) + "/" + strconv.Itoa(toBal))
	err = stub.PutState(args[0], []byte(strconv.Itoa(fromBal))) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	err = stub.PutState(args[1], []byte(strconv.Itoa(toBal))) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}

	// update accounts
	accounts[fromIndex].PointBalance = fromBal
	accounts[fromIndex].Out = accounts[fromIndex].Out + points
	accounts[toIndex].PointBalance = toBal
	accounts[toIndex].In = accounts[toIndex].In + points
	accountsBytes, _ = json.Marshal(&accounts)
	err = stub.PutState("accounts", accountsBytes)
	if err != nil {
		return nil, err
	}

	// add transaction
	transaction.FromID = args[0]
	transaction.ToID = args[1]
	transaction.Points = points
	transactions = append(transactions, transaction)
	transBytes, _ := json.Marshal(&transactions)
	err = stub.PutState("transactions", transBytes)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
