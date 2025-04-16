package aws

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
)

type AWSNodeConfig struct {
	ImageID            string   `json:"image_id"`
	SubnetID           string   `json:"subnet_id"`
	SecurityGroupIDs   []string `json:"security_group_ids"`
	InstanceType       string   `json:"instance_type"`
	BlockDeviceGb      int      `json:"block_device_gb"`
	BlockDeviceName    string   `json:"block_device_name"` //default /dev/xvda
	BlockDeviceType    string   `json:"block_device_type"`
	PlacementGroupName string   `json:"placement_group_name"`
}

type AWSNode struct {
	NodeName  string         `json:"name"`
	IPAddress string         `json:"ip_address"`
	NodeRole  string         `json:"role"`
	Config    *AWSNodeConfig `json:"config"`
}

func (c AWSNode) Name() (nodeName string) {
	return c.NodeName
}

func (c AWSNode) Role() (role string) {
	return c.NodeRole
}

func (c AWSNode) IP() (ip string) {
	return c.IPAddress
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
