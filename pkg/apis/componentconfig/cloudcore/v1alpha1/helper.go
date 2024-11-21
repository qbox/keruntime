/*
Copyright 2019 The KubeEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/kubeedge/kubeedge/pkg/util/validation"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

type IptablesMgrMode string

const (
	InternalMode IptablesMgrMode = "internal"
	ExternalMode IptablesMgrMode = "external"
)

func (c *CloudCoreConfig) Parse(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		klog.Errorf("Failed to read configfile %s: %v", filename, err)
		return err
	}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		klog.Errorf("Failed to unmarshal configfile %s: %v", filename, err)
		return err
	}
	return nil
}

func (c *CloudCoreConfig) ReadNodeID(filename string) error {
	nodeIDFile := ""
	// Read node id from default file
	if validation.FileIsExist(filename) {
		nodeIDFile = filename
	}
	// Read node id from config file
	if c.CloudCoreNodeIDFile != "" && validation.FileIsExist(c.CloudCoreNodeIDFile) {
		nodeIDFile = c.CloudCoreNodeIDFile
	}
	// Check node id exist
	if nodeIDFile == "" && !validation.FileIsExist(nodeIDFile) {
		// Check default node id exist
		if c.CloudCoreNodeID == "" {
			err := errors.New("failed to read configfile")
			klog.Errorf("Failed to read both configfile %s and %s: %v", filename, c.CloudCoreNodeIDFile, err)
			return err
		} else {
			if c.Modules.CloudIdentity.IDType == 2 {
				klog.Infof("Using CloudCoreNodeID from config file as node id: %s", c.CloudCoreNodeID)
				c.Modules.CloudIdentity.ID = c.CloudCoreNodeID
			}
			return nil
		}
	}
	data, err := os.ReadFile(nodeIDFile)
	if err != nil {
		klog.Errorf("Failed to read configfile %s: %v", nodeIDFile, err)
		return err
	}
	c.CloudCoreNodeID = string(data)
	if c.Modules.CloudIdentity.IDType == 2 {
		klog.Infof("Using nodeIDFile data as node id, file: %s, nodeID: %s", nodeIDFile, c.CloudCoreNodeID)
		c.Modules.CloudIdentity.ID = string(data)
	}
	return nil
}

func (in *IptablesManager) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	// Define a secondary type so that we don't end up with a recursive call to json.Unmarshal
	type IM IptablesManager
	var out = (*IM)(in)
	err := json.Unmarshal(data, &out)
	if err != nil {
		return err
	}

	// Validate the valid enum values
	switch in.Mode {
	case InternalMode, ExternalMode:
		return nil
	default:
		in.Mode = ""
		return errors.New("invalid value for iptablesmgr mode")
	}
}
