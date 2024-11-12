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
	case "removeLadingBillCrossParams":
		return t.removeLadingBillCrossParams(stub, args)
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
	if ladingBillCrossParamsObject.LadingBillCR.Tdbh == "" {
		return shim.Error("call issueLadingBillCrossParams meet error:tdbh can not be empty")
	}

	ladingBillCrossParamsMap, err := t.getLadingBillCrossParamsMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call issueLadingBillCrossParams meet error: %s", err.Error()))
	}

	crossChainStatusMap, err := t.getCrossChainStatusMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call issueLadingBillCrossParams meet error: %s", err.Error()))
	}

	//跨链接收到的提单不允许更新
	oldLadingBillCrossParamsObject, ok := ladingBillCrossParamsMap[ladingBillCrossParamsObject.LadingBillCR.Tdbh]
	if ok && crossChainStatusMap[ladingBillCrossParamsObject.LadingBillCR.Tdbh] == CrossChainReceiptSent {
		return shim.Error(fmt.Sprintf("this ladingBillCrossParams received from partner and update is not allowed,ladingBillNumber:%s", ladingBillCrossParamsObject.LadingBillCR.Tdbh))
	}

	//不能更新被冻结的提单
	if ok && oldLadingBillCrossParamsObject.Freeze {
		return shim.Error(fmt.Sprintf("this ladingBillCrossParams has been frozen and update is not allowed,ladingBillNumber:%s", ladingBillCrossParamsObject.LadingBillCR.Tdbh))
	}

	//change cross chain status
	crossChainStatusMap[ladingBillCrossParamsObject.LadingBillCR.Tdbh] = CrossChainOnChain
	err = t.putCrossChainStatusMap(stub, crossChainStatusMap)
	if err != nil {
		return shim.Error(err.Error())
	}

	ladingBillCrossParamsMap[ladingBillCrossParamsObject.LadingBillCR.Tdbh] = ladingBillCrossParamsObject
	err = t.putLadingBillCrossParamsMap(stub, ladingBillCrossParamsMap)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(ladingBillCrossParamsObject.LadingBillCR.Tdbh))
}

// removeLadingBillCrossParams remove lading bill cross params
func (t *Transfer) removeLadingBillCrossParams(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("call removeLadingBillCrossParams meet error: incorrect number of arguments")
	}
	ladingBillNumber := args[0]

	ladingBillCrossParamsMap, err := t.getLadingBillCrossParamsMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call removeLadingBillCrossParams meet error: %s", err.Error()))
	}

	crossChainStatusMap, err := t.getCrossChainStatusMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call removeLadingBillCrossParams meet error: %s", err.Error()))
	}

	//change cross chain status
	delete(crossChainStatusMap, ladingBillNumber)
	err = t.putCrossChainStatusMap(stub, crossChainStatusMap)
	if err != nil {
		return shim.Error(err.Error())
	}

	delete(ladingBillCrossParamsMap, ladingBillNumber)
	err = t.putLadingBillCrossParamsMap(stub, ladingBillCrossParamsMap)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(ladingBillNumber))
}

// queryLadingBillCrossParams query lading bill cross params
func (t *Transfer) queryLadingBillCrossParams(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("call queryLadingBillCrossParams meet error: incorrect number of arguments")
	}
	ladingBillNumber := args[0]

	ladingBillCrossParamsMap, err := t.getLadingBillCrossParamsMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call queryLadingBillCrossParams meet error: %s", err.Error()))
	}

	ladingBillCrossParamsObject, ok := ladingBillCrossParamsMap[ladingBillNumber]
	if !ok {
		return shim.Success([]byte("ladingBillNumber:" + ladingBillNumber + " not found"))
	}
	ladingBillCrossParamsBytes, err := json.Marshal(ladingBillCrossParamsObject)
	return shim.Success(ladingBillCrossParamsBytes)
}

// queryCrossChainStatus query cross chain status
func (t *Transfer) queryCrossChainStatus(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("call queryCrossChainStatus meet error: incorrect number of arguments")
	}
	ladingBillNumber := args[0]

	crossChainStatusMap, err := t.getCrossChainStatusMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call queryCrossChainStatus meet error: %s", err.Error()))
	}
	status := crossChainStatusMap[ladingBillNumber]

	return shim.Success([]byte(status.String()))
}

// begin cross-chain call
func (t *Transfer) transferLadingBillCrossParams(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("call transferLadingBillCrossParams meet error: incorrect number of arguments")
	}
	dstServiceID := args[0]
	ladingBillNumber := args[1]

	ladingBillCrossParamsMap, err := t.getLadingBillCrossParamsMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call transferLadingBillCrossParams meet error: %s", err.Error()))
	}

	ladingBillCrossParamsObject, ok := ladingBillCrossParamsMap[ladingBillNumber]
	if !ok {
		return shim.Success([]byte(ladingBillNumber + " not found"))
	}

	//若提单被冻结，无法转发
	if ladingBillCrossParamsObject.Freeze {
		return shim.Error(fmt.Sprintf("this ladingBillCrossParams has been frozen and transfer is not allowed,ladingBillNumber:%s", ladingBillCrossParamsObject.LadingBillCR.Tdbh))
	}

	//每次跨链转发提单都赋予唯一的跨链ID及时间戳
	timestamp, _ := stub.GetTxTimestamp()
	ladingBillCrossParamsObject.CrossChainID = strconv.FormatInt(int64(timestamp.GetNanos()), 10)
	ladingBillCrossParamsObject.Timestamp = int64(timestamp.GetNanos())
	//跨链转发-回退完成期间，冻结提单
	ladingBillCrossParamsObject.Freeze = true

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
	argsRb = append(argsRb, []byte(ladingBillNumber))
	argsRbBytes, err := json.Marshal(argsRb)
	if err != nil {
		return shim.Error(err.Error())
	}

	var argsCb [][]byte
	argsCb = append(argsCb, []byte(ladingBillNumber))
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

	//跨链转发期间，修改跨链状态为forward
	crossChainStatusMap, err := t.getCrossChainStatusMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call issueLadingBillCrossParams meet error: %s", err.Error()))
	}
	crossChainStatusMap[ladingBillCrossParamsObject.LadingBillCR.Tdbh] = CrossChainForwarded

	err = t.putCrossChainStatusMap(stub, crossChainStatusMap)
	if err != nil {
		return shim.Error(err.Error())
	}
	ladingBillCrossParamsMap[ladingBillCrossParamsObject.LadingBillCR.Tdbh] = ladingBillCrossParamsObject
	err = t.putLadingBillCrossParamsMap(stub, ladingBillCrossParamsMap)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *Transfer) transferLadingBillCrossParamsRollback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	ladingBillNumber := args[0]

	//跨链失败，修改跨链状态为rollback
	crossChainStatusMap, err := t.getCrossChainStatusMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call transferLadingBillCrossParamsRollback meet error: %s", err.Error()))
	}
	crossChainStatusMap[ladingBillNumber] = CrossChainRollback

	err = t.putCrossChainStatusMap(stub, crossChainStatusMap)
	if err != nil {
		return shim.Error(err.Error())
	}
	//跨链失败，解冻提单
	ladingBillCrossParamsMap, err := t.getLadingBillCrossParamsMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call transferLadingBillCrossParams meet error: %s", err.Error()))
	}

	ladingBillCrossParamsObject, ok := ladingBillCrossParamsMap[ladingBillNumber]
	if !ok {
		return shim.Success([]byte(ladingBillNumber + " not found"))
	}

	ladingBillCrossParamsObject.Freeze = false
	ladingBillCrossParamsMap[ladingBillCrossParamsObject.LadingBillCR.Tdbh] = ladingBillCrossParamsObject
	err = t.putLadingBillCrossParamsMap(stub, ladingBillCrossParamsMap)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

func (t *Transfer) transferLadingBillCrossParamsCallBack(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	ladingBillNumber := args[0]

	//change cross chain status
	crossChainStatusMap, err := t.getCrossChainStatusMap(stub)
	if err != nil {
		return shim.Error(fmt.Sprintf("call transferLadingBillCrossParamsCallBack meet error: %s", err.Error()))
	}
	crossChainStatusMap[ladingBillNumber] = CrossChainReceiptReceived

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

	//提单存在且crossChainStatus不等于CrossChainReceiptSent,说明该提单是本侧未跨链转发的新提单，对端不能跨链覆盖
	_, ok := ladingBillCrossParamsMap[ladingBillCrossParamsObject.LadingBillCR.Tdbh]
	if ok && crossChainStatusMap[ladingBillCrossParamsObject.LadingBillCR.Tdbh] != CrossChainReceiptSent {
		return shim.Error(fmt.Sprintf("call ladingBillCrossChainCall meet error:tdbh %s already exist!", ladingBillCrossParamsObject.LadingBillCR.Tdbh))
	}

	//如果是跨链回退，解冻原提单
	oldLadingBillNumber := strings.TrimSuffix(ladingBillCrossParamsObject.LadingBillCR.Tdbh, "-back")
	if oldLadingBillNumber != ladingBillCrossParamsObject.LadingBillCR.Tdbh {
		oldLadingBillCrossParamsObject, ok := ladingBillCrossParamsMap[oldLadingBillNumber]
		if ok {
			oldLadingBillCrossParamsObject.Freeze = false
			ladingBillCrossParamsMap[oldLadingBillNumber] = oldLadingBillCrossParamsObject
			err = t.putLadingBillCrossParamsMap(stub, ladingBillCrossParamsMap)
			if err != nil {
				return shim.Error(fmt.Sprintf("call ladingBillCrossChainCall:unfreeze meet error:%s", err.Error()))
			}
		}
	}

	ladingBillCrossParamsMap[ladingBillCrossParamsObject.LadingBillCR.Tdbh] = ladingBillCrossParamsObject
	err = t.putLadingBillCrossParamsMap(stub, ladingBillCrossParamsMap)
	if err != nil {
		return shim.Error(err.Error())
	}

	//change cross chain status
	crossChainStatusMap[ladingBillCrossParamsObject.LadingBillCR.Tdbh] = CrossChainReceiptSent

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
