package identity

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"net"

	"github.com/google/uuid"
	"github.com/kubeedge/kubeedge/pkg/apis/componentconfig/cloudcore/v1alpha1"
)

// IDType define a way to config the id type between UUID, Hash, Configured, System.
type IDType uint

const (
	// UUID will automatically generate a id for cloud.
	UUID IDType = 0
	// Hash will generate a hash id over cloud mac address and addr.
	Hash IDType = 1
	// Configured will use the id user config.
	Configured IDType = 2
)

// CloudInfo is some cloud info to generate hash.
type CloudInfo struct {
	MacAddr  string `json:"macAddr,omitempty"`
	HostAddr string `json:"hostAddr,omitempty"`
	WsAddr   string `json:"wsAddr,omitempty"`
}

// CloudIdentity cache the cloud core id.
type CloudIdentity struct {
	// Configured indicates whether the identifier displays the ID as automatically generated or manually configured
	idType IDType

	// cloudInfo is only use while idType is Hash.
	cloudInfo *CloudInfo

	// id is the cloud id, readonly.
	id string
}

// NewCloudIdentity need modules config, return CloudIdentity, will generate id while new CloudIdentity.
func NewCloudIdentity(modules *v1alpha1.Modules) *CloudIdentity {
	conf := &CloudIdentity{
		idType: IDType(modules.CloudIdentity.IDType),
	}

	switch conf.idType {
	case Configured:
		if err := conf.validateID(modules.CloudIdentity.ID); err != nil {
			return nil
		}
	case Hash:
		conf.cloudInfo = &CloudInfo{
			MacAddr:  netMac(),
			HostAddr: modules.CloudHub.HTTPS.Address,
			WsAddr:   modules.CloudHub.WebSocket.Address,
		}
		if err := conf.hashID(); err != nil {
			return nil
		}
	case UUID:
		if err := conf.uuid(); err != nil {
			return nil
		}
	default:
		conf.id = modules.CloudIdentity.ID
	}

	if conf.id == "" {
		if err := conf.uuid(); err != nil {
			return nil
		}
		conf.idType = UUID
	}
	return conf
}

// CloudID get cloud id.
func (c *CloudIdentity) CloudID() string {
	return c.id
}

// validateID check the id is not empty.
func (c *CloudIdentity) validateID(id string) error {
	if id == "" {
		return errors.New("id validate failed ")
	}
	c.id = id
	return nil
}

// uuid generate the id for CloudIdentity.
func (c *CloudIdentity) uuid() error {
	uid, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	c.id = uid.String()
	return nil
}

// hashID generate the id for CloudIdentity.
func (c *CloudIdentity) hashID() error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(c.cloudInfo); err != nil {
		return err
	}
	h := sha256.Sum256(buffer.Bytes())
	c.id = string(h[:])
	return nil
}

// netMac read system eth port interfaces and return it hardware mac addr.
func netMac() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			return iface.HardwareAddr.String()
		}
	}
	return ""
}
