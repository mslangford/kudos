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

type Kudos struct {
	Name    string `json:"name"`
	Balance int    `json:"balance"`
}

type KudosTab struct {
	Balances []Kudos `json:"balances"`
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	var err error
	var bal Kudos
	var balTab KudosTab
	var j int

	for i := 0; i < len(args); i = i + 2 {
		fmt.Println("setting balance for " + args[i] + " to " + args[i+1])
		err = stub.PutState(args[i], []byte(args[i+1]))
		if err != nil {
			return nil, err
		}
		bal.Name = args[i]
		bal.Balance, err = strconv.Atoi(args[i+1])
		if err != nil {
			return nil, err
		}
		balTab.Balances[j] = bal
		j++
	}

	tab := KudosTab{}
	tab.Balances = balTab.Balances //change the user

	jsonAsBytes, _ := json.Marshal(tab)
	err = stub.PutState("kudos", jsonAsBytes) //rewrite the marble with id as key
	if err != nil {
		return nil, err
	}
	fmt.Println("kudos: " + string(jsonAsBytes))

	return nil, nil
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {

	// Handle different functions
	fmt.Println("invoke is running " + function)
	if function == "init" {
		return t.Init(stub, "init", args)
	} else if function == "write" {
		return t.write(stub, args)
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
	var jsonResp string
	var err error
	var fromState, toState []byte
	var fromBal, toBal, points int

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting from & to usernames and number of points")
	}

	//	from balance
	fmt.Println("from balance: " + args[0])
	fromState, err = stub.GetState(args[0])
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get current balance for " + args[0] + "\"}"
		return nil, errors.New(jsonResp)
	}
	fromBal, err = strconv.Atoi(string(fromState))
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to convert current balance for " + args[0] + "\"}"
		return nil, errors.New(jsonResp)
	}

	//	transfer points
	fmt.Println("convert points")
	points, err = strconv.Atoi(args[2])
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to convert points to integer}"
		return nil, errors.New(jsonResp)
	}
	fmt.Println("points: " + strconv.Itoa(points))
	if fromBal < points {
		jsonResp = "{\"Error\":\"Point balance does not cover transfer amount for " + args[0] + "\"}"
		return nil, errors.New(jsonResp)
	}

	//	to balance
	fmt.Println("to balance: " + args[1])
	toState, err = stub.GetState(args[1])
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get current balance for " + args[1] + "\"}"
		return nil, errors.New(jsonResp)
	}
	toBal, err = strconv.Atoi(string(toState))
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to convert current balance for " + args[1] + "\"}"
		return nil, errors.New(jsonResp)
	}

	//	apply transfer
	fromBal = fromBal - points
	toBal = toBal + points
	fmt.Println("apply transfer - new from points " + strconv.Itoa(fromBal) + " new to points " + strconv.Itoa(toBal))
	err = stub.PutState(args[0], []byte(strconv.Itoa(fromBal))) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	err = stub.PutState(args[1], []byte(strconv.Itoa(toBal))) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}

	return nil, nil
}
