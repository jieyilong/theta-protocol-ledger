package core

import (
	"github.com/thetatoken/ukulele/common"
	"github.com/thetatoken/ukulele/common/result"
)

type ViewSelector int

const (
	DeliveredView ViewSelector = 1
	CheckedView   ViewSelector = 2
	ScreenedView  ViewSelector = 3
)

//
// Ledger defines the interface of the ledger
//
type Ledger interface {
	ScreenTx(rawTx common.Bytes) result.Result
	ProposeBlockTxs() (stateRootHash common.Hash, blockRawTxs []common.Bytes, res result.Result)
	ApplyBlockTxs(blockRawTxs []common.Bytes, expectedStateRoot common.Hash) result.Result
	ResetState(height uint32, rootHash common.Hash) result.Result
	Query()
}