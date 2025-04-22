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
	Domain             string `yaml:"domain"`
}

type AWSNode struct {
	NodeName   string         `yaml:"name"`
	IPAddress  string         `yaml:"ip_address"`
	NodeRole   string         `yaml:"role"`
	NodeID     string         `yaml:"id"`
	Config     *AWSNodeConfig `yaml:"config"`
	NodeDomain string         `yaml:"domain"`
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
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		err = errors.Wrapf(err, "failed unmarshalling data into struct")
		return config, err
	}

	return config, err
}
