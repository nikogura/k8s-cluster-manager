package vault

import (
	"context"
	"fmt"
	"strings"

	vault "github.com/hashicorp/vault/api"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// VaultSecretManager implements manager.SecretManager for HashiCorp Vault.
type VaultSecretManager struct {
	Client    *vault.Client
	MountPath string
}

// NewVaultSecretManager creates a new Vault secret manager.
func NewVaultSecretManager(address string, token string, mountPath string) (vsm *VaultSecretManager, err error) {
	config := vault.DefaultConfig()
	config.Address = address

	client, clientErr := vault.NewClient(config)
	if clientErr != nil {
		err = errors.Wrapf(clientErr, "failed creating Vault client")
		return vsm, err
	}

	client.SetToken(token)

	vsm = &VaultSecretManager{
		Client:    client,
		MountPath: mountPath,
	}

	return vsm, err
}

// GetClusterSecret retrieves the complete secret for a cluster role.
func (v *VaultSecretManager) GetClusterSecret(ctx context.Context, clusterName string, role string) (secret manager.ClusterSecret, err error) {
	// Build secret path: {mount}/cluster-{clusterName}-{role}
	secretPath := fmt.Sprintf("cluster-%s-%s", clusterName, role)

	// Read secret from Vault
	vaultSecret, readErr := v.Client.KVv2(v.MountPath).Get(ctx, secretPath)
	if readErr != nil {
		err = errors.Wrapf(readErr, "failed reading secret %s from Vault", secretPath)
		return secret, err
	}

	if vaultSecret == nil || vaultSecret.Data == nil {
		err = fmt.Errorf("secret %s not found in Vault", secretPath)
		return secret, err
	}

	// Populate cluster secret
	secret.ClusterName = clusterName
	secret.Role = role

	// Extract core fields
	secret.ConfigYAML = getStringField(vaultSecret.Data, "config.yaml")
	secret.PatchYAML = getStringField(vaultSecret.Data, "patch.yaml")

	// Extract ImageID from node-aws.yaml
	secret.ImageID = extractImageID(vaultSecret.Data)

	// Extract installer version from config.yaml
	secret.InstallerVersion = extractInstallerVersion(secret.ConfigYAML)

	// Store other fields in CloudProviderExtra
	secret.CloudProviderExtra = extractExtraFields(vaultSecret.Data)

	return secret, err
}

func getStringField(data map[string]interface{}, key string) (value string) {
	if val, ok := data[key].(string); ok {
		value = val
	}
	return value
}

func extractImageID(data map[string]interface{}) (imageID string) {
	nodeAWSYAML := getStringField(data, "node-aws.yaml")
	if nodeAWSYAML == "" {
		return imageID
	}

	var nodeConfig map[string]interface{}
	parseErr := yaml.Unmarshal([]byte(nodeAWSYAML), &nodeConfig)
	if parseErr != nil {
		return imageID
	}

	if imgID, ok := nodeConfig["image_id"].(string); ok {
		imageID = imgID
	}

	return imageID
}

func extractInstallerVersion(configYAML string) (version string) {
	if configYAML == "" {
		return version
	}

	var machineConfig map[string]interface{}
	parseErr := yaml.Unmarshal([]byte(configYAML), &machineConfig)
	if parseErr != nil {
		return version
	}

	machine, mOk := machineConfig["machine"].(map[string]interface{})
	if !mOk {
		return version
	}

	install, iOk := machine["install"].(map[string]interface{})
	if !iOk {
		return version
	}

	image, imgOk := install["image"].(string)
	if !imgOk {
		return version
	}

	// Extract version from image (e.g., "ghcr.io/siderolabs/installer:v1.10.8" -> "v1.10.8")
	parts := strings.Split(image, ":")
	if len(parts) == 2 {
		version = parts[1]
	}

	return version
}

func extractExtraFields(data map[string]interface{}) (extra map[string]string) {
	extra = make(map[string]string)
	for key, value := range data {
		if key != "config.yaml" && key != "patch.yaml" && key != "node-aws.yaml" {
			if strVal, ok := value.(string); ok {
				extra[key] = strVal
			}
		}
	}
	return extra
}

// UpdateClusterSecret updates the secret with new values.
func (v *VaultSecretManager) UpdateClusterSecret(ctx context.Context, secret manager.ClusterSecret) (err error) {
	// Build secret path
	secretPath := fmt.Sprintf("cluster-%s-%s", secret.ClusterName, secret.Role)

	// Read existing secret first to preserve all fields
	existing, readErr := v.GetClusterSecret(ctx, secret.ClusterName, secret.Role)
	if readErr != nil {
		err = errors.Wrapf(readErr, "failed reading existing secret")
		return err
	}

	// Build data map starting with existing extra fields
	data := buildDataMap(existing.CloudProviderExtra)

	// Set core fields
	data["config.yaml"] = secret.ConfigYAML
	data["patch.yaml"] = secret.PatchYAML

	// Update node-aws.yaml with new ImageID if provided
	if secret.ImageID != "" {
		nodeAWSYAML, updateErr := updateNodeAWSYAML(existing.CloudProviderExtra["node-aws.yaml"], secret.ImageID)
		if updateErr != nil {
			err = errors.Wrapf(updateErr, "failed updating node-aws.yaml")
			return err
		}
		data["node-aws.yaml"] = nodeAWSYAML
	}

	// Write to Vault
	_, writeErr := v.Client.KVv2(v.MountPath).Put(ctx, secretPath, data)
	if writeErr != nil {
		err = errors.Wrapf(writeErr, "failed writing secret to Vault")
		return err
	}

	return err
}

func buildDataMap(extra map[string]string) (data map[string]interface{}) {
	data = make(map[string]interface{})
	for key, value := range extra {
		data[key] = value
	}
	return data
}

func updateNodeAWSYAML(existingYAML string, imageID string) (updatedYAML string, err error) {
	var nodeConfig map[string]interface{}

	if existingYAML != "" {
		parseErr := yaml.Unmarshal([]byte(existingYAML), &nodeConfig)
		if parseErr != nil {
			err = errors.Wrapf(parseErr, "failed parsing node-aws.yaml")
			return updatedYAML, err
		}
	} else {
		nodeConfig = make(map[string]interface{})
	}

	// Update image_id
	nodeConfig["image_id"] = imageID

	// Marshal back to YAML
	nodeAWSBytes, marshalErr := yaml.Marshal(nodeConfig)
	if marshalErr != nil {
		err = errors.Wrapf(marshalErr, "failed marshaling node-aws.yaml")
		return updatedYAML, err
	}

	updatedYAML = string(nodeAWSBytes)
	return updatedYAML, err
}

// UpdateVersionInfo updates only the version-related fields in the secret.
func (v *VaultSecretManager) UpdateVersionInfo(ctx context.Context, clusterName string, role string, imageID string, version string) (err error) {
	// Build secret path
	secretPath := fmt.Sprintf("cluster-%s-%s", clusterName, role)

	// Read existing secret
	existing, readErr := v.GetClusterSecret(ctx, clusterName, role)
	if readErr != nil {
		err = errors.Wrapf(readErr, "failed reading existing secret")
		return err
	}

	// Parse config.yaml to update installer version
	var machineConfig map[string]interface{}
	parseErr := yaml.Unmarshal([]byte(existing.ConfigYAML), &machineConfig)
	if parseErr != nil {
		err = errors.Wrapf(parseErr, "failed parsing config.yaml")
		return err
	}

	// Update installer image version
	installerImage := fmt.Sprintf("ghcr.io/siderolabs/installer:%s", version)

	if machine, mOk := machineConfig["machine"].(map[string]interface{}); mOk {
		if install, iOk := machine["install"].(map[string]interface{}); iOk {
			install["image"] = installerImage
		} else {
			machine["install"] = map[string]interface{}{
				"image": installerImage,
			}
		}
	}

	// Marshal back to YAML
	configBytes, marshalErr := yaml.Marshal(machineConfig)
	if marshalErr != nil {
		err = errors.Wrapf(marshalErr, "failed marshaling config.yaml")
		return err
	}

	// Update node-aws.yaml with new AMI
	var nodeConfig map[string]interface{}
	if nodeAWSYAML, ok := existing.CloudProviderExtra["node-aws.yaml"]; ok && nodeAWSYAML != "" {
		parseErr = yaml.Unmarshal([]byte(nodeAWSYAML), &nodeConfig)
		if parseErr != nil {
			err = errors.Wrapf(parseErr, "failed parsing node-aws.yaml")
			return err
		}
	} else {
		nodeConfig = make(map[string]interface{})
	}

	nodeConfig["image_id"] = imageID

	nodeAWSBytes, marshalErr := yaml.Marshal(nodeConfig)
	if marshalErr != nil {
		err = errors.Wrapf(marshalErr, "failed marshaling node-aws.yaml")
		return err
	}

	// Build data map with all fields
	data := make(map[string]interface{})

	// Copy existing extra fields
	for key, value := range existing.CloudProviderExtra {
		if key != "node-aws.yaml" {
			data[key] = value
		}
	}

	// Set updated fields
	data["config.yaml"] = string(configBytes)
	data["patch.yaml"] = existing.PatchYAML
	data["node-aws.yaml"] = string(nodeAWSBytes)

	// Write to Vault
	_, writeErr := v.Client.KVv2(v.MountPath).Put(ctx, secretPath, data)
	if writeErr != nil {
		err = errors.Wrapf(writeErr, "failed writing secret to Vault")
		return err
	}

	return err
}
