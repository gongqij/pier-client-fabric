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
	case "transfer":
		return t.transfer(stub, args)
	case "getBalance":
		return t.getBalance(stub, args)
	case "setBalance":
		return t.setBalance(stub, args)
	case "interchainCharge":
		return t.interchainCharge(stub, args)
	case "interchainRollback":
		return t.interchainRollback(stub, args)
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

func (t *Transfer) transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	switch len(args) {
	case 3:
		sender := args[0]
		receiver := args[1]
		amountArg := args[2]
		amount, err := getAmountArg(amountArg)
		if err != nil {
			return shim.Error(fmt.Errorf("get amount from arg: %w", err).Error())
		}

		balance, err := getUint64(stub, sender)
		if err != nil {
			return shim.Error(fmt.Errorf("got account value from %s %w", sender, err).Error())
		}

		if balance < amount {
			return shim.Error("not sufficient funds")
		}

		balance -= amount

		err = stub.PutState(sender, []byte(strconv.FormatUint(balance, 10)))
		if err != nil {
			return shim.Error(err.Error())
		}

		receiverBalance, err := getUint64(stub, receiver)
		if err != nil {
			return shim.Error(fmt.Errorf("got account value from %s %w", receiver, err).Error())
		}

		err = stub.PutState(receiver, []byte(strconv.FormatUint(receiverBalance+amount, 10)))
		if err != nil {
			return shim.Error(err.Error())
		}

		return shim.Success(nil)
	case 4:
		dstServiceID := args[0]
		sender := args[1]
		receiver := args[2]
		amountArg := args[3]

		amount, err := getAmountArg(amountArg)
		if err != nil {
			return shim.Error(fmt.Errorf("get amount from arg: %w", err).Error())
		}

		balance, err := getUint64(stub, sender)
		if err != nil {
			return shim.Error(fmt.Errorf("got account value from %s %w", sender, err).Error())
		}

		if balance < amount {
			return shim.Error("not sufficient funds")
		}

		balance -= amount

		err = stub.PutState(sender, []byte(strconv.FormatUint(balance, 10)))
		if err != nil {
			return shim.Error(err.Error())
		}

		var argsRb [][]byte
		var callArgs []string
		transferAmount := strconv.FormatUint(amount, 10)
		callArgs = append(callArgs, sender)
		callArgs = append(callArgs, receiver)
		callArgs = append(callArgs, transferAmount)
		callArgsBytes, err := json.Marshal(callArgs)
		if err != nil {
			return shim.Error(err.Error())
		}
		var typAndArgs [][]byte
		typAndArgs = append(typAndArgs, []byte("java.util.List<java.lang.String>"), callArgsBytes)
		typAndArgsBytes, err := json.Marshal(typAndArgs)
		if err != nil {
			return shim.Error(err.Error())
		}

		argsRb = append(argsRb, []byte(sender))
		argsRb = append(argsRb, []byte(transferAmount))
		argsRbBytes, err := json.Marshal(argsRb)
		if err != nil {
			return shim.Error(err.Error())
		}

		b := util.ToChaincodeArgs(emitInterchainEventFunc, dstServiceID, "interchainCharge", string(typAndArgsBytes), "", "", "interchainRollback", string(argsRbBytes), strconv.FormatBool(false))
		response := stub.InvokeChaincode(brokerContractName, b, channelID)
		if response.Status != shim.OK {
			return shim.Error(fmt.Errorf("invoke broker chaincode: %d - %s", response.Status, response.Message).Error())
		}

		return shim.Success(nil)
	default:
		return shim.Error(fmt.Sprintf("incorrect number of arguments %d", len(args)))
	}
}

// getBalance gets account balance
func (t *Transfer) getBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("incorrect number of arguments")
	}

	name := args[0]

	value, err := stub.GetState(name)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(value)
}

// setBalance sets account balance
func (t *Transfer) setBalance(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("incorrect number of arguments")
	}

	name := args[0]
	amount := args[1]

	if err := stub.PutState(name, []byte(amount)); err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// charge user,amount
func (t *Transfer) interchainCharge(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//["java.util.List<java.lang.String>",`["daming","xiaoming","10"]`,"false"]
	callArgs := make([]string, 3)
	err := json.Unmarshal([]byte(args[1]), &callArgs)
	if err != nil {
		return shim.Error(err.Error())
	}
	if len(callArgs) != 3 {
		return shim.Error("incorrect number of arguments, expect 3")
	}
	sender := callArgs[0]
	receiver := callArgs[1]
	transferAmount, err := strconv.ParseUint(callArgs[2], 10, 64)
	if err != nil {
		return shim.Error(fmt.Errorf("strconv.ParseUint meet err: %w", err).Error())
	}
	isRollback := args[2]

	// check for sender info
	if sender == "" {
		return shim.Error("incorrect sender info")
	}

	balance, err := getUint64(stub, receiver)
	if err != nil {
		return shim.Error(fmt.Errorf("get balancee from %s %w", receiver, err).Error())
	}

	// TODO: deal with rollback failure (balance not enough)
	if isRollback == "true" {
		balance -= transferAmount
	} else {
		balance += transferAmount
	}
	err = stub.PutState(receiver, []byte(strconv.FormatUint(balance, 10)))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func (t *Transfer) interchainRollback(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	//["java.util.List<java.lang.String>",`["sender","amount"]`,"true"]
	rollbackCallArgs := make([]string, 2)
	err := json.Unmarshal([]byte(args[1]), &rollbackCallArgs)
	if err != nil {
		return shim.Error(err.Error())
	}
	if len(rollbackCallArgs) != 2 {
		return shim.Error("incorrect number of arguments, expecting 2")
	}

	name := rollbackCallArgs[0]
	transferAmount, err := strconv.ParseUint(rollbackCallArgs[1], 10, 64)

	balance, err := getUint64(stub, name)
	if err != nil {
		return shim.Error(fmt.Errorf("get balancee from %s %w", name, err).Error())
	}

	balance += transferAmount
	err = stub.PutState(name, []byte(strconv.FormatUint(balance, 10)))
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
