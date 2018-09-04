package messenger

import (
	log "github.com/sirupsen/logrus"

	"github.com/thetatoken/ukulele/common"
	"github.com/thetatoken/ukulele/p2p"
	pr "github.com/thetatoken/ukulele/p2p/peer"
	p2ptypes "github.com/thetatoken/ukulele/p2p/types"
)

//
// Messenger implements the Network interface
//
type Messenger struct {
	discMgr       *PeerDiscoveryManager
	msgHandlerMap map[common.ChannelIDEnum](p2p.MessageHandler)
	peerTable     pr.PeerTable
	nodeInfo      p2ptypes.NodeInfo // information of our blockchain node
}

//
// MessengerConfig specifies the configuration for Messenger
//
type MessengerConfig struct {
}

// CreateMessenger creates an instance of Messenger
func CreateMessenger(nodeInfo p2ptypes.NodeInfo, addrBookFilePath string, routabilityRestrict bool, selfNetAddressStr string,
	seedPeerNetAddressStrs []string, networkProtocol string, localNetworkAddr string,
	skipUPNP bool) (*Messenger, error) {

	messenger := &Messenger{
		msgHandlerMap: make(map[common.ChannelIDEnum](p2p.MessageHandler)),
		peerTable:     pr.CreatePeerTable(),
		nodeInfo:      nodeInfo,
	}

	discMgrConfig := GetDefaultPeerDiscoveryManagerConfig()
	discMgr, err := CreatePeerDiscoveryManager(messenger, &nodeInfo, addrBookFilePath,
		routabilityRestrict, selfNetAddressStr, seedPeerNetAddressStrs, networkProtocol,
		localNetworkAddr, skipUPNP, &messenger.peerTable, discMgrConfig)
	if err != nil {
		log.Errorf("[p2p] Failed to create CreatePeerDiscoveryManager")
		return messenger, err
	}

	discMgr.SetMessenger(messenger)
	messenger.SetPeerDiscoveryManager(discMgr)

	return messenger, nil
}

// SetPeerDiscoveryManager sets the PeerDiscoveryManager for the Messenger
func (msgr *Messenger) SetPeerDiscoveryManager(discMgr *PeerDiscoveryManager) {
	msgr.discMgr = discMgr
}

// OnStart is called when the Messenger starts
func (msgr *Messenger) OnStart() error {
	err := msgr.discMgr.OnStart()
	return err
}

// OnStop is called when the Messenger stops
func (msgr *Messenger) OnStop() {
	msgr.discMgr.OnStop()
}

// Broadcast broadcasts the given message to all the connected peers
func (msgr *Messenger) Broadcast(message p2ptypes.Message) (successes chan bool) {
	log.Debugf("[p2p] Broadcasting messages...")
	allPeers := msgr.peerTable.GetAllPeers()
	successes = make(chan bool, len(*allPeers))
	for _, peer := range *allPeers {
		log.Debugf("[p2p] Broadcasting \"%v\" to %v", message.Content, peer.ID())
		go func(peer *pr.Peer) {
			success := msgr.Send(peer.ID(), message)
			successes <- success
		}(peer)
	}
	return successes
}

// Send sends the given message to the specified peer
func (msgr *Messenger) Send(peerID string, message p2ptypes.Message) bool {
	peer := msgr.peerTable.GetPeer(peerID)
	if peer == nil {
		return false
	}

	success := peer.Send(message.ChannelID, message.Content)

	return success
}

// AddMessageHandler adds the message handler
func (msgr *Messenger) AddMessageHandler(msgHandler p2p.MessageHandler) bool {
	channelIDs := msgHandler.GetChannelIDs()
	for _, channelID := range channelIDs {
		if msgr.msgHandlerMap[channelID] != nil {
			log.Errorf("[p2p] Message handlered is already added for channelID: %v", channelID)
			return false
		}
		msgr.msgHandlerMap[channelID] = msgHandler
	}
	return true
}

// ID returns the ID of the current node
func (msgr *Messenger) ID() string {
	return msgr.nodeInfo.Address
}

// AttachMessageHandlersToPeer attaches the registerred message handlers to the given peer
func (msgr *Messenger) AttachMessageHandlersToPeer(peer *pr.Peer) {
	messageParser := func(channelID common.ChannelIDEnum, rawMessageBytes common.Bytes) (p2ptypes.Message, error) {
		msgHandler := msgr.msgHandlerMap[channelID]
		if msgHandler == nil {
			log.Errorf("[p2p] Failed to setup message parser for channelID %v", channelID)
		}
		message, err := msgHandler.ParseMessage(channelID, rawMessageBytes)
		return message, err
	}
	peer.GetConnection().SetMessageParser(messageParser)

	receiveHandler := func(message p2ptypes.Message) error {
		channelID := message.ChannelID
		msgHandler := msgr.msgHandlerMap[channelID]
		if msgHandler == nil {
			log.Errorf("[p2p] Failed to setup message handler for channelID %v", channelID)
		}
		peerID := peer.ID()
		err := msgHandler.HandleMessage(peerID, message)
		return err
	}
	peer.GetConnection().SetReceiveHandler(receiveHandler)

	// TODO: error handling..
	// errorHandler := func(interface{}) {
	// 	msgr.discMgr.HandlePeerWithErrors(peer)
	// }
	// peer.GetConnection().SetErrorHandler(errorHandler)
}
