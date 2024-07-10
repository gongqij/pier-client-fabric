package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/core/chaincode/shim"
)

func getUint64(stub shim.ChaincodeStubInterface, key string) (uint64, error) {
	value, err := stub.GetState(key)
	if err != nil {
		return 0, fmt.Errorf("amount must be an interger %w", err)
	}

	ret, err := strconv.ParseUint(string(value), 10, 64)
	if err != nil {
		return 0, err
	}

	return ret, nil
}

func getAmountArg(arg string) (uint64, error) {
	amount, err := strconv.ParseUint(arg, 10, 64)
	if err != nil {
		shim.Error(fmt.Errorf("amount must be an interger %w", err).Error())
		return 0, err
	}

	if amount < 0 {
		return 0, fmt.Errorf("amount must be a positive integer, got %s", arg)
	}

	return amount, nil
}

func (t *Transfer) getLadingBillCrossParamsMap(stub shim.ChaincodeStubInterface) (map[string]ladingBillCrossParams, error) {
	ladingBillCrossMapBytes, err := stub.GetState(ladingBillsCrossParamsMapKey)
	if err != nil {
		return nil, err
	}
	ladingBillsMap := make(map[string]ladingBillCrossParams)
	if ladingBillCrossMapBytes == nil {
		return ladingBillsMap, nil
	}
	if err := json.Unmarshal(ladingBillCrossMapBytes, &ladingBillsMap); err != nil {
		return nil, err
	}
	return ladingBillsMap, nil
}

func (t *Transfer) putLadingBillCrossParamsMap(stub shim.ChaincodeStubInterface, ladingBillsMap map[string]ladingBillCrossParams) error {
	ladingBillsMapBytes, err := json.Marshal(ladingBillsMap)
	if err != nil {
		return err
	}
	return stub.PutState(ladingBillsCrossParamsMapKey, ladingBillsMapBytes)
}

func (t *Transfer) getCrossChainStatusMap(stub shim.ChaincodeStubInterface) (map[string]CrossChainStatus, error) {
	crossChainStatusMapBytes, err := stub.GetState(crossChainStatusKey)
	if err != nil {
		return nil, err
	}
	crossChainStatusMap := make(map[string]CrossChainStatus)
	if crossChainStatusMapBytes == nil {
		return crossChainStatusMap, nil
	}
	if err := json.Unmarshal(crossChainStatusMapBytes, &crossChainStatusMap); err != nil {
		return nil, err
	}
	return crossChainStatusMap, nil
}

func (t *Transfer) putCrossChainStatusMap(stub shim.ChaincodeStubInterface, crossChainStatusMap map[string]CrossChainStatus) error {
	crossChainStatusMapBytes, err := json.Marshal(crossChainStatusMap)
	if err != nil {
		return err
	}
	return stub.PutState(crossChainStatusKey, crossChainStatusMapBytes)
}
