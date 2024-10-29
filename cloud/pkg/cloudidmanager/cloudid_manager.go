package cloudidmanager

import (
	"github.com/kubeedge/kubeedge/cloud/pkg/cloudidmanager/identity"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha1"
)

var CloudIDManager *cloudIDManager

type cloudIDManager struct {
	enable bool

	ID *identity.CloudIdentity
}

func Register(modules *v1alpha1.Modules) {
	CloudIDManager = newCloudIDManager(modules)
}

func newCloudIDManager(modules *v1alpha1.Modules) *cloudIDManager {
	CloudIDManager = &cloudIDManager{
		enable: modules.CloudIDManager.Enable,
		ID:     identity.NewCloudIdentity(modules),
	}
	return CloudIDManager
}

func (c *cloudIDManager) Enable() bool {
	return c.enable
}

func (c *cloudIDManager) GetCloudID() string {
	return c.ID.ID()
}

func (c *cloudIDManager) IsIDSelf(id string) bool {
	return c.ID.ID() == id
}

func (c *cloudIDManager) IsNodeConnectSelf(nodeID string) bool {
	return c.IsIDSelf(getNodeConnected())
}

func getNodeConnected() string {
	return "TODO"
}
