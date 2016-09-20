/*
Copyright IBM Corp 2016 All Rights Reserved.

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

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

// SimpleChaincode example simple Chaincode implementation
type SimpleChaincode struct {
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return nil, err
	}

	return nil, nil
}

// Invoke isur entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub *shim.ChaincodeStub, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
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

// write - invoke function to write key/value pair
func (t *SimpleChaincode) write(stub *shim.ChaincodeStub, args []string) ([]byte, error) {
	var key, value string
	var err error
	fmt.Println("running write()")

	if len(args) != 2 {
		return nil, errors.New("Incorrect number of arguments. Expecting 2. name of the key and value to set")
	}

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

	if len(args) != 3 {
		return nil, errors.New("Incorrect number of arguments. Expecting from & to usernames and number of points")
	}

	//	from balance
	fromState, err = stub.GetState(args[0])
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get current balance for " + args[0] + "\"}"
		return nil, errors.New(jsonResp)
	}
	buf1 := bytes.NewBuffer(fromState)
	fromBal, err2 := binary.ReadVarint(buf1)
	if err2 != nil {
		jsonResp = "{\"Error\":\"Failed to get current balance for " + args[0] + "\"}"
		return nil, errors.New(jsonResp)
	}

	//	transfer points
	points, err3 := strconv.Atoi(args[2])
	if err3 != nil {
		jsonResp = "{\"Error\":\"Failed to convert points to integer}"
		return nil, errors.New(jsonResp)
	}
	if fromBal < int64(points) {
		jsonResp = "{\"Error\":\"Failed to get current balance for " + args[0] + "\"}"
		return nil, errors.New(jsonResp)
	}

	//	to balance
	toState, err4 := stub.GetState(args[1])
	if err4 != nil {
		jsonResp = "{\"Error\":\"Failed to get current balance for " + args[1] + "\"}"
		return nil, errors.New(jsonResp)
	}
	buf2 := bytes.NewBuffer(toState)
	toBal, err5 := binary.ReadVarint(buf2)
	if err5 != nil {
		jsonResp = "{\"Error\":\"Failed to get current balance for " + args[1] + "\"}"
		return nil, errors.New(jsonResp)
	}

	//	apply transfer
	toBal = toBal + int64(points)
	fromBal = fromBal - int64(points)
	err = stub.PutState(args[0], []byte(strconv.Itoa(int(fromBal)))) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}
	err = stub.PutState(args[1], []byte(strconv.Itoa(int(toBal)))) //write the variable into the chaincode state
	if err != nil {
		return nil, err
	}

	return nil, nil
}
