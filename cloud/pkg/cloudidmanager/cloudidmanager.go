package cloudidmanager

import (
	"github.com/kubeedge/beehive/pkg/core"
	"github.com/kubeedge/kubeedge/cloud/pkg/cloudidmanager/config"
	"github.com/kubeedge/kubeedge/cloud/pkg/cloudidmanager/identity"
	"github.com/kubeedge/kubeedge/cloud/pkg/common/modules"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha1"
)

var CloudIDManager *cloudIDManager

type cloudIDManager struct {
	enable bool

	ID *identity.CloudIdentity
}

// Group implements core.Module.
func (c *cloudIDManager) Group() string {
	return modules.CloudIDManagerGroup
}

// Name implements core.Module.
func (c *cloudIDManager) Name() string {
	return modules.CloudIDManagerName
}

// Start implements core.Module.
func (c *cloudIDManager) Start() {

}

func Register(cidm *v1alpha1.CloudIDManager) {
	config.InitConfig(cidm)
	core.Register(newCloudIDManager(cidm))
}

func newCloudIDManager(cidm *v1alpha1.CloudIDManager) *cloudIDManager {
	CloudIDManager = &cloudIDManager{
		enable: true,
		ID:     identity.NewCloudIdentity(cidm.Enable, cidm.ID),
	}
	return CloudIDManager
}

func (c *cloudIDManager) Enable() bool {
	return c.enable
}

func (c *cloudIDManager) GetCloudID() string {
	return c.ID.Hash()
}

func (c *cloudIDManager) IsIDSelf(id string) bool {
	return c.ID.Hash() == id
}
