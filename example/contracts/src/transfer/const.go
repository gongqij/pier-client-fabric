package main

type CrossChainStatus int

const (
	CrossChainOnChain CrossChainStatus = iota + 1
	CrossChainForwarded
	CrossChainRollback
	CrossChainReceiptReceived
	CrossChainReceiptSent
)

func (c CrossChainStatus) String() string {
	switch c {
	case CrossChainOnChain:
		return "on chain"
	case CrossChainForwarded:
		return "forward"
	case CrossChainRollback:
		return "rollback"
	case CrossChainReceiptReceived:
		return "receipt received"
	case CrossChainReceiptSent:
		return "receipt sent"
	default:
		return "not found"
	}
}
