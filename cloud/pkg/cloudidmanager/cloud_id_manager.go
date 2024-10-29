package cloudidmanager

import (
	"github.com/kubeedge/kubeedge/cloud/pkg/cloudidmanager/identity"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha1"
)

var CloudIDManager *cloudIDManager

type cloudIDManager struct {
	enable bool

	Identity *identity.CloudIdentity

	ConnectionInfoManager *Manager
}

func Register(modules *v1alpha1.Modules) {
	CloudIDManager = newCloudIDManager(modules)
}

func newCloudIDManager(modules *v1alpha1.Modules) *cloudIDManager {
	CloudIDManager = &cloudIDManager{
		enable:                modules.CloudIDManager.Enable,
		Identity:              identity.NewCloudIdentity(modules),
		ConnectionInfoManager: NewCloudConnectionManager(),
	}
	return CloudIDManager
}

func (c *cloudIDManager) Enable() bool {
	return c.enable
}

func (c *cloudIDManager) GetCloudID() string {
	return c.Identity.ID()
}

func (c *cloudIDManager) IsIDSelf(id string) bool {
	return c.Identity.ID() == id
}
