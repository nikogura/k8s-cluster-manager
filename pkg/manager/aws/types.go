package aws

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"os"
)

type AWSNodeConfig struct {
	ImageID            string `yaml:"image_id"`
	SubnetID           string `yaml:"subnet_id"`
	InstanceType       string `yaml:"instance_type"`
	BlockDeviceGb      string `yaml:"block_device_gb"`
	BlockDeviceName    string `yaml:"block_device_name"`
	BlockDeviceType    string `yaml:"block_device_type"`
	PlacementGroupName string `yaml:"placement_group_name"`
	Domain             string `yaml:"Domain"`
}

type AWSNode struct {
	NodeName   string         `yaml:"name"`
	IPAddress  string         `yaml:"ip_address"`
	NodeRole   string         `yaml:"role"`
	NodeID     string         `yaml:"id"`
	Config     *AWSNodeConfig `yaml:"config"`
	NodeDomain string         `yaml:"Domain"`
}

func (n AWSNode) Name() (result string) {
	result = n.NodeName
	return result
}

func (n AWSNode) Role() (result string) {
	result = n.NodeRole
	return result
}

func (n AWSNode) IP() (result string) {
	result = n.IPAddress
	return result
}

func (n AWSNode) ID() (result string) {
	result = n.NodeID
	return result
}

func (n AWSNode) Domain() (result string) {
	result = n.NodeDomain
	return result
}

func LoadAWSNodeConfigFromFile(filePath string) (config AWSNodeConfig, err error) {
	configBytes, readErr := os.ReadFile(filePath)
	if readErr != nil {
		err = errors.Wrapf(readErr, "failed reading file %s", filePath)
		return config, err
	}

	config, err = LoadAWSNodeConfig(configBytes)
	return config, err
}

func LoadAWSNodeConfig(data []byte) (config AWSNodeConfig, err error) {
	unmarshalErr := yaml.Unmarshal(data, &config)
	if unmarshalErr != nil {
		err = errors.Wrapf(unmarshalErr, "failed unmarshalling data into struct")
		return config, err
	}

	return config, err
}
