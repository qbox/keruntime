package identity

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
)

type CloudIdentity struct {
	// Configed indicates whether the identifier displays the ID as automatically generated or manually configured
	Configed bool

	Name  string   `json:"name"`
	Label []string `json:"label,omitempty"`

	id string
}

func NewCloudIdentity(configed bool, id string) *CloudIdentity {
	conf := &CloudIdentity{
		Configed: false,
	}
	if configed && id != "" {
		conf.Configed = configed
		conf.id = id
		return conf
	}
	if err := conf.hashID(); err != nil {
		return nil
	}
	return conf
}

func (c *CloudIdentity) Hash() string {
	return c.id
}

func (c *CloudIdentity) hashID() error {
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(c); err != nil {
		return err
	}
	h := sha256.Sum256(buffer.Bytes())
	c.id = string(h[:])
	return nil
}
