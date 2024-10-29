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

type IDType uint

const (
	UUID       IDType = iota
	Hash       IDType = iota
	Configured IDType = iota
)

type CloudInfo struct {
	MacAddr  string `json:"macAddr,omitempty"`
	HostAddr string `json:"hostAddr,omitempty"`
	WsAddr   string `json:"wsAddr,omitempty"`
}

type CloudIdentity struct {
	// Configured indicates whether the identifier displays the ID as automatically generated or manually configured
	idType IDType

	cloudInfo *CloudInfo

	id string
}

func NewCloudIdentity(modules *v1alpha1.Modules) *CloudIdentity {
	conf := &CloudIdentity{
		idType: IDType(modules.CloudIDManager.IDType),
		cloudInfo: &CloudInfo{
			MacAddr:  netMac(),
			HostAddr: modules.CloudHub.HTTPS.Address,
			WsAddr:   modules.CloudHub.WebSocket.Address,
		},
	}

	switch conf.idType {
	case Configured:
		if err := conf.validateID(modules.CloudIDManager.ID); err != nil {
			return nil
		}
	case Hash:
		if err := conf.hashID(); err != nil {
			return nil
		}
	default:
		if err := conf.uuid(); err != nil {
			return nil
		}
	}

	if conf.id == "" {
		if err := conf.uuid(); err != nil {
			return nil
		}
		conf.idType = UUID
	}
	return conf
}

func (c *CloudIdentity) ID() string {
	return c.id
}

func (c *CloudIdentity) validateID(id string) error {
	if id == "" {
		return errors.New("id validate failed ")
	}
	c.id = id
	return nil
}

func (c *CloudIdentity) uuid() error {
	uid, err := uuid.NewUUID()
	if err != nil {
		return err
	}
	c.id = uid.String()
	return nil
}

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
