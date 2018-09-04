package node

import (
	"context"
	"sync"

	"github.com/thetatoken/ukulele/blockchain"
	"github.com/thetatoken/ukulele/consensus"
	"github.com/thetatoken/ukulele/dispatcher"
	"github.com/thetatoken/ukulele/netsync"
	"github.com/thetatoken/ukulele/p2p"
	"github.com/thetatoken/ukulele/store"
)

type Node struct {
	Store       store.Store
	Chain       *blockchain.Chain
	Consensus   consensus.Engine
	SyncManager *netsync.SyncManager
	Dispatcher  *dispatcher.Dispatcher
	Network     p2p.Network

	// Life cycle
	wg      *sync.WaitGroup
	quit    chan struct{}
	ctx     context.Context
	cancel  context.CancelFunc
	stopped bool
}

type Params struct {
	ChainID    string
	Root       *blockchain.Block
	Validators *consensus.ValidatorSet
	Network    p2p.Network
	Store      store.Store
}

func NewNode(params *Params) *Node {
	chain := blockchain.NewChain(params.ChainID, params.Store, params.Root)
	consensus := consensus.NewEngine(chain, params.Network, params.Validators)
	dispatcher := dispatcher.NewDispatcher(params.Network)
	syncMgr := netsync.NewSyncManager(chain, consensus, params.Network, dispatcher)

	return &Node{
		Store:       params.Store,
		Chain:       chain,
		Consensus:   consensus,
		SyncManager: syncMgr,
		Dispatcher:  dispatcher,
		Network:     params.Network,
	}
}

// Start starts sub components and kick off the main loop.
func (n *Node) Start(ctx context.Context) {
	c, cancel := context.WithCancel(ctx)
	n.ctx = c
	n.cancel = cancel

	n.Consensus.Start(n.ctx)
	n.SyncManager.Start(n.ctx)
}

// Stop notifies all sub components to stop without blocking.
func (n *Node) Stop() {
	n.cancel()
}

// Wait blocks until all sub components stop.
func (n *Node) Wait() {
	n.Consensus.Wait()
	n.SyncManager.Wait()
}
