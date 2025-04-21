package aws

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
)

type AWSNodeConfig struct {
	ImageID            string `json:"image_id"`
	SubnetID           string `json:"subnet_id"`
	InstanceType       string `json:"instance_type"`
	BlockDeviceGb      int    `json:"block_device_gb"`
	BlockDeviceName    string `json:"block_device_name"` //default /dev/xvda
	BlockDeviceType    string `json:"block_device_type"`
	PlacementGroupName string `json:"placement_group_name"`
	Domain             string `json:"domain"`
}

type AWSNode struct {
	NodeName   string         `json:"name"`
	IPAddress  string         `json:"ip_address"`
	NodeRole   string         `json:"role"`
	NodeID     string         `json:"id"`
	Config     *AWSNodeConfig `json:"config"`
	NodeDomain string         `json:"domain"`
}

func (n AWSNode) Name() (nodeName string) {
	return n.NodeName
}

func (n AWSNode) Role() (role string) {
	return n.NodeRole
}

func (n AWSNode) IP() (ip string) {
	return n.IPAddress
}

func (n AWSNode) ID() (id string) {
	return n.NodeID
}

func (n AWSNode) Domain() (domain string) {
	return n.NodeDomain
}

func LoadAWSNodeConfigFromFile(filePath string) (config AWSNodeConfig, err error) {
	configBytes, err := os.ReadFile(filePath)
	if err != nil {
		err = errors.Wrapf(err, "failed reading file %s", filePath)
		return config, err
	}

	return LoadAWSNodeConfig(configBytes)
}

func LoadAWSNodeConfig(data []byte) (config AWSNodeConfig, err error) {
	err = json.Unmarshal(data, &config)
	if err != nil {
		err = errors.Wrapf(err, "failed unmarshalling data into struct")
		return config, err
	}

	return config, err
}
