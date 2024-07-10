package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric/common/util"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
)

const (
	channelID               = "mychannel"
	brokerContractName      = "broker"
	emitInterchainEventFunc = "EmitInterchainEvent"

	ladingBillsCrossParamsMapKey = "ladingBillsOrigin"

	crossChainStatusKey = "crossChainStatusKey"
)

type Transfer struct{}

func (t *Transfer) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (t *Transfer) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	fmt.Printf("invoke: %s\n", function)
	switch function {
	case "register":
		return t.register(stub, args)
	case "issueLadingBillCrossParams":
		return t.issueLadingBillCrossParams(stub, args)
	case "queryLadingBillCrossParams":
		return t.queryLadingBillCrossParams(stub, args)
	case "queryCrossChainStatus":
		return t.queryCrossChainStatus(stub, args)
	case "transferLadingBillCrossParams":
		return t.transferLadingBillCrossParams(stub, args)
	case "transferLadingBillCrossParamsRollback":
		return t.transferLadingBillCrossParamsRollback(stub, args)
	case "transferLadingBillCrossParamsCallBack":
		return t.transferLadingBillCrossParamsCallBack(stub, args)
	case "ladingBillCrossChainCall":
		return t.ladingBillCrossChainCall(stub, args)
	default:
		return shim.Error("invalid function: " + function + ", args: " + strings.Join(args, ","))
	}
}

func (t *Transfer) register(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		shim.Error("incorrect number of arguments, expecting 1")
	}
	invokeArgs := util.ToChaincodeArgs("register", args[0])
	response := stub.InvokeChaincode(brokerContractName, invokeArgs, channelID)
	if response.Status != shim.OK {
		return shim.Error(fmt.Sprintf("invoke chaincode '%s' err: %s", brokerContractName, response.Message))
	}
	return response
}

// issueLadingBillCrossParams issue lading bill cross params
func (t *Transfer) issueLadingBillCrossParams(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("call issueLadingBillCrossParams meet error: incorrect number of arguments")
	}
	ladingBillCrossParamsJsonStr := args[0]

	var ladingBillCrossParamsObject ladingBillCrossParams
	if err := json.Unmarshal([]byte(ladingBillCrossParamsJsonStr), &ladingBillCrossParamsObject); err != nil {
		return shim.Error("call issueLadingBillCrossParams meet error:" + err.Error())
	}

	if ladingBillCrossParamsObject.CrossChainID == "" {
		return shim.Error("call issueLadingBillCrossParams meet error: crossChainID can not be empty")
	}

	ladingBillCrossParamsMap, err := t.getLadingBillCrossParamsMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call issueLadingBillCrossParams meet error: %s", err.Error()))
	}

	crossChainStatusMap, err := t.getCrossChainStatusMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call issueLadingBillCrossParams meet error: %s", err.Error()))
	}

	//以下情况不允许更新提单
	//1、跨链转发成功的提单不允许更新
	_, ok := ladingBillCrossParamsMap[ladingBillCrossParamsObject.CrossChainID]
	if ok && crossChainStatusMap[ladingBillCrossParamsObject.CrossChainID] == CrossChainReceiptReceived {
		return shim.Error(fmt.Sprintf("this ladingBillCrossParams has been forward and update is not allowed,crossChainID:%s", ladingBillCrossParamsObject.CrossChainID))
	}
	//2、跨链接收到的提单不允许更新
	if ok && crossChainStatusMap[ladingBillCrossParamsObject.CrossChainID] == CrossChainReceiptSent {
		return shim.Error(fmt.Sprintf("this ladingBillCrossParams received from partner and update is not allowed,crossChainID:%s", ladingBillCrossParamsObject.CrossChainID))
	}

	//change cross chain status
	crossChainStatusMap[ladingBillCrossParamsObject.CrossChainID] = CrossChainOnChain

	err = t.putCrossChainStatusMap(stub, crossChainStatusMap)
	if err != nil {
		return shim.Error(err.Error())
	}

	ladingBillCrossParamsMap[ladingBillCrossParamsObject.CrossChainID] = ladingBillCrossParamsObject
	err = t.putLadingBillCrossParamsMap(stub, ladingBillCrossParamsMap)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(ladingBillCrossParamsObject.CrossChainID))
}

// queryLadingBillCrossParams query lading bill cross params
func (t *Transfer) queryLadingBillCrossParams(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("call queryLadingBillCrossParams meet error: incorrect number of arguments")
	}
	crossChainID := args[0]

	ladingBillCrossParamsMap, err := t.getLadingBillCrossParamsMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call queryLadingBillCrossParams meet error: %s", err.Error()))
	}

	ladingBillCrossParamsObject, ok := ladingBillCrossParamsMap[crossChainID]
	if !ok {
		return shim.Success([]byte("crossChainID:" + crossChainID + " not found"))
	}
	ladingBillCrossParamsBytes, err := json.Marshal(ladingBillCrossParamsObject)
	return shim.Success(ladingBillCrossParamsBytes)
}

// queryCrossChainStatus query cross chain status
func (t *Transfer) queryCrossChainStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("call queryCrossChainStatus meet error: incorrect number of arguments")
	}
	crossChainID := args[0]

	crossChainStatusMap, err := t.getCrossChainStatusMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call queryCrossChainStatus meet error: %s", err.Error()))
	}
	status := crossChainStatusMap[crossChainID]

	return shim.Success([]byte(status.String()))
}

// begin cross-chain call
func (t *Transfer) transferLadingBillCrossParams(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("call transferLadingBillCrossParams meet error: incorrect number of arguments")
	}
	dstServiceID := args[0]
	crossChainID := args[1]

	ladingBillCrossParamsMap, err := t.getLadingBillCrossParamsMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call transferLadingBillCrossParams meet error: %s", err.Error()))
	}

	ladingBillCrossParamsObject, ok := ladingBillCrossParamsMap[crossChainID]
	if !ok {
		return shim.Success([]byte(crossChainID + " not found"))
	}

	ladingBillCrossParamsBytes, err := json.Marshal(ladingBillCrossParamsObject)

	var callArgs []string
	callArgs = append(callArgs, string(ladingBillCrossParamsBytes))
	callArgsBytes, err := json.Marshal(callArgs)
	if err != nil {
		return shim.Error(err.Error())
	}
	var typAndArgs [][]byte
	//目的链hvm合约，需加上参数类型java.util.List<java.lang.String>
	typAndArgs = append(typAndArgs, []byte("java.util.List<java.lang.String>"), callArgsBytes)
	typAndArgsBytes, err := json.Marshal(typAndArgs)
	if err != nil {
		return shim.Error(err.Error())
	}

	var argsRb [][]byte
	argsRb = append(argsRb, []byte(crossChainID))
	argsRbBytes, err := json.Marshal(argsRb)
	if err != nil {
		return shim.Error(err.Error())
	}

	var argsCb [][]byte
	argsCb = append(argsCb, []byte(crossChainID))
	argsCbBytes, err := json.Marshal(argsCb)
	if err != nil {
		return shim.Error(err.Error())
	}

	//begin cross chain
	b := util.ToChaincodeArgs(emitInterchainEventFunc, dstServiceID, "ladingBillCrossChainCall", string(typAndArgsBytes), "transferLadingBillCrossParamsCallBack", string(argsCbBytes), "transferLadingBillCrossParamsRollback", string(argsRbBytes), strconv.FormatBool(false))
	response := stub.InvokeChaincode(brokerContractName, b, channelID)
	if response.Status != shim.OK {
		return shim.Error(fmt.Errorf("invoke broker chaincode: %d - %s", response.Status, response.Message).Error())
	}

	//change cross chain status
	crossChainStatusMap, err := t.getCrossChainStatusMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call issueLadingBillCrossParams meet error: %s", err.Error()))
	}
	crossChainStatusMap[ladingBillCrossParamsObject.CrossChainID] = CrossChainForwarded

	err = t.putCrossChainStatusMap(stub, crossChainStatusMap)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *Transfer) transferLadingBillCrossParamsRollback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	crossChainID := args[0]

	//change cross chain status
	crossChainStatusMap, err := t.getCrossChainStatusMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call transferLadingBillCrossParamsRollback meet error: %s", err.Error()))
	}
	crossChainStatusMap[crossChainID] = CrossChainRollback

	err = t.putCrossChainStatusMap(stub, crossChainStatusMap)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *Transfer) transferLadingBillCrossParamsCallBack(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	crossChainID := args[0]

	//change cross chain status
	crossChainStatusMap, err := t.getCrossChainStatusMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call transferLadingBillCrossParamsCallBack meet error: %s", err.Error()))
	}
	crossChainStatusMap[crossChainID] = CrossChainReceiptReceived

	err = t.putCrossChainStatusMap(stub, crossChainStatusMap)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *Transfer) ladingBillCrossChainCall(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	callArgs := make([]string, 1)
	//注意：args[0]为参数类型，非实际参数
	err := json.Unmarshal([]byte(args[1]), &callArgs)
	if err != nil {
		return shim.Error(err.Error())
	}
	ladingBillCrossParamsObject := ladingBillCrossParams{}
	err = json.Unmarshal([]byte(callArgs[0]), &ladingBillCrossParamsObject)
	if err != nil {
		return shim.Error(err.Error())
	}
	ladingBillCrossParamsMap, err := t.getLadingBillCrossParamsMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call ladingBillCrossChainCall meet error: %s", err.Error()))
	}

	crossChainStatusMap, err := t.getCrossChainStatusMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call ladingBillCrossChainCall meet error: %s", err.Error()))
	}

	//crossChainID存在且crossChainStatus不等于CrossChainReceiptSent,说明该提单是本侧未跨链转发的新提单，对端不能跨链覆盖
	_, ok := ladingBillCrossParamsMap[ladingBillCrossParamsObject.CrossChainID]
	if ok && crossChainStatusMap[ladingBillCrossParamsObject.CrossChainID] != CrossChainReceiptSent {
		return shim.Error(fmt.Sprintf("call ladingBillCrossChainCall meet error:crossChainID %s already exist!", ladingBillCrossParamsObject.CrossChainID))
	}

	ladingBillCrossParamsMap[ladingBillCrossParamsObject.CrossChainID] = ladingBillCrossParamsObject
	err = t.putLadingBillCrossParamsMap(stub, ladingBillCrossParamsMap)
	if err != nil {
		return shim.Error(err.Error())
	}

	//change cross chain status
	crossChainStatusMap[ladingBillCrossParamsObject.CrossChainID] = CrossChainReceiptSent

	err = t.putCrossChainStatusMap(stub, crossChainStatusMap)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(Transfer))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
