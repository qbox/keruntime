package config

import (
	"sync"

	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha1"
)

var Config Configure
var once sync.Once

type Configure struct {
	v1alpha1.CloudIDManager
}

func InitConfig(cidm *v1alpha1.CloudIDManager) {
	once.Do(func() {
		Config = Configure{
			CloudIDManager: *cidm,
		}
	})
}
