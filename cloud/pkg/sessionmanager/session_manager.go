package sessionmanager

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/kubeedge/kubeedge/cloud/pkg/cloudhub/common"
	"github.com/kubeedge/kubeedge/cloud/pkg/common/monitor"
	"github.com/kubeedge/kubeedge/cloud/pkg/sessionmanager/identity"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha1"
	"k8s.io/klog/v2"
)

// NodeSession is the interface for node session. Each edge node has a NodeSession, which is used to manage the connection with the edge node.
type NodeSession interface {
	GetNodeID() string
	// GetNodeConnectedCloudID return the id which the cloud node connect to.
	GetNodeConnectedCloudID() string

	// KeepAliveMessage receive keepalive message from edge node
	KeepAliveMessage()
	// KeepAliveCheck A goroutine running KeepAliveCheck is started for each connection.
	KeepAliveCheck()

	// SendAckMessage loops forever sending message that require acknowledgment
	// to the edge node until an error is encountered (or the connection is closed).
	SendAckMessage()
	// SendNoAckMessage loops forever sending the message that does not require acknowledgment
	// to the edge node until an error is encountered (or the connection is closed).
	SendNoAckMessage()

	ReceiveMessageAck(parentID string)

	// Start the main goroutine responsible for serving node session
	Start()
	// Terminating terminates the node session and it is called when the client goes offline,
	// It will shutdown the message queue and the goroutine associated with it will exit.
	Terminating()

	// SetTerminateErr set the terminate error code.
	SetTerminateErr(terminateErr int32)
	// GetTerminateErr return terminate code.
	GetTerminateErr() int32

	// GetNodeMessagePool return the common.NodeMessagePool
	GetNodeMessagePool() *common.NodeMessagePool
}

// SessionManager is the manager for node session, and itself have a Identity to store the cloud identity.
type SessionManager struct {
	Identity *identity.CloudIdentity
	// NodeNumber is the number of currently connected edge
	// nodes for single cloudHub instance
	NodeNumber int32
	// NodeLimit is the maximum number of edge nodes that can
	// connected to single cloudHub instance
	NodeLimit int32
	// NodeSessions maps a node ID to NodeSession
	NodeSessions sync.Map
}

// NewSessionManager initializes a new SessionManager
func NewSessionManager(modules *v1alpha1.Modules) *SessionManager {
	return &SessionManager{
		Identity:     identity.NewCloudIdentity(modules),
		NodeLimit:    modules.CloudHub.NodeLimit,
		NodeSessions: sync.Map{},
	}
}

// AddSession add node session to the session manager
func (sm *SessionManager) AddSession(session NodeSession) {
	nodeID := session.GetNodeID()

	ons, exists := sm.NodeSessions.LoadAndDelete(nodeID)
	if exists {
		if oldSession, ok := ons.(NodeSession); ok {
			klog.Warningf("session exists for %s, close old session", nodeID)
			oldSession.Terminating()
			atomic.AddInt32(&sm.NodeNumber, -1)
		}
	}

	sm.NodeSessions.Store(nodeID, session)
	monitor.ConnectedNodes.Set(float64(atomic.AddInt32(&sm.NodeNumber, 1)))
}

// DeleteSession delete the node session from session manager
func (sm *SessionManager) DeleteSession(session NodeSession) {
	cacheSession, exist := sm.GetSession(session.GetNodeID())
	if !exist {
		klog.Warningf("session not found for node %s", session.GetNodeID())
		return
	}

	// This usually happens when the node is disconnect then quickly reconnect
	if cacheSession != session {
		klog.Warningf("the session %s already deleted", session.GetNodeID())
		return
	}

	sm.NodeSessions.Delete(session.GetNodeID())
	monitor.ConnectedNodes.Set(float64(atomic.AddInt32(&sm.NodeNumber, -1)))
}

// GetSession get the node session for the node
func (sm *SessionManager) GetSession(nodeID string) (NodeSession, bool) {
	ons, exists := sm.NodeSessions.Load(nodeID)
	if exists {
		return ons.(NodeSession), true
	}

	return nil, false
}

// ReachLimit checks whether the connected nodes exceeds the node limit number
func (sm *SessionManager) ReachLimit() bool {
	return atomic.LoadInt32(&sm.NodeNumber) >= sm.NodeLimit
}

// KeepAliveMessage receive keepalive message from edge node
func (sm *SessionManager) KeepAliveMessage(nodeID string) error {
	session, exist := sm.GetSession(nodeID)
	if !exist {
		return fmt.Errorf("session not found for node %s", nodeID)
	}

	session.KeepAliveMessage()
	return nil
}

// ReceiveMessageAck receive the message ack from edge node
func (sm *SessionManager) ReceiveMessageAck(nodeID, parentID string) error {
	session, exist := sm.GetSession(nodeID)
	if !exist {
		return fmt.Errorf("session not found for node %s", nodeID)
	}

	session.ReceiveMessageAck(parentID)
	return nil
}

// IsNodeConnectSelf find the node connected cloud, with is self.
func (sm *SessionManager) IsNodeConnectSelf(nodeID string) (string, bool) {
	nodeSession, exist := sm.GetSession(nodeID)
	if !exist {
		return "", false
	}

	isSelf := sm.IsCloudSelf(nodeSession.GetNodeConnectedCloudID())
	return nodeSession.GetNodeConnectedCloudID(), isSelf
}

// GetCloudID return this running cloud core id.
func (sm *SessionManager) GetCloudID() string {
	return sm.Identity.CloudID()
}

// IsCloudSelf return is the cloud core id itself.
func (sm *SessionManager) IsCloudSelf(id string) bool {
	return sm.Identity.CloudID() == id
}
