package manager

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"net/http"
	"os"
	"strings"
)

const CloudflareZoneIDEnvVar = "CLOUDFLARE_ZONE_ID"
const CloudflareAPITokenEnvVar = "CLOUDFLARE_API_TOKEN"

// VaultAPIConfig creates a vault api config in a standard fashion.
func VaultAPIConfig(address string) (config *api.Config, err error) {
	// read the environment and use that over anything
	config = api.DefaultConfig()

	err = config.ReadEnvironment()
	if err != nil {
		err = errors.Wrapf(err, "failed to inject environment into client config")
		return config, err
	}

	if config.Address == "https://127.0.0.1:8200" {
		if address != "" {
			config.Address = address
		}
	}

	rootCAs, rootCAsErr := x509.SystemCertPool()
	if rootCAsErr != nil {
		err = errors.Wrapf(rootCAsErr, "failed to get system cert pool")
		return config, err
	}

	// for using private CA's
	//if cacert != "" {
	//	ok := rootCAs.AppendCertsFromPEM([]byte(cacert))
	//	if !ok {
	//		err = errors.New("Failed to add root cert to system CA bundle")
	//		return config, err
	//	}
	//}

	clientConfig := &tls.Config{
		RootCAs: rootCAs,
	}

	config.HttpClient.Transport = &http.Transport{TLSClientConfig: clientConfig}

	return config, err
}

func NewVaultClient(token string, verbose bool) (client *api.Client, err error) {
	apiConfig, apiConfigErr := VaultAPIConfig(os.Getenv("VAULT_ADDR"))
	if apiConfigErr != nil {
		err = apiConfigErr
		return client, err
	}

	if verbose {
		fmt.Printf("Vault Address: %s\n", apiConfig.Address)
	}

	client, err = api.NewClient(apiConfig)
	if err != nil {
		err = errors.Wrapf(err, "failed to create vault api client")
		return client, err
	}

	client.SetToken(token)

	return client, err
}

func SecretData(client *api.Client, path string, verbose bool) (data map[string]interface{}, err error) {
	data = make(map[string]interface{})

	VerboseOutput(verbose, "Reading path: %s", path)

	v2Path, v2PathErr := V2Path(path)
	if v2PathErr != nil {
		err = errors.Wrapf(v2PathErr, "failed creating v2 secret path from %q", path)
		return data, err
	}

	s, readErr := client.Logical().Read(v2Path)
	if readErr != nil {
		err = errors.Wrapf(readErr, "Failed to lookup path: %s", path)
		return data, err
	}

	if s != nil {
		secretData, ok := s.Data["data"].(map[string]interface{})
		if ok {
			for k, v := range secretData {
				data[k] = v
			}
		} else {
			err = errors.New(fmt.Sprintf("unparsable secret found at %s", path))
		}
	}

	//verboseOutput(verbose, "\n\n")

	return data, err
}

func V2Path(path string) (v2Path string, err error) {
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		err = errors.New(fmt.Sprintf("Invalid path %q.  Path must consist of a mount and a secret separated by slashes", path))
		return v2Path, err
	}

	mount := parts[0]
	secretPath := parts[1]

	v2Path = fmt.Sprintf("%s/data/%s", mount, secretPath)

	return v2Path, err
}

type ConfigData struct {
	TalosMachineConfig      []byte
	TalosMachineConfigPatch []byte
	NodeConfig              []byte
	CloudflareAPIToken      string
	CloudflareZoneID        string
}

func ConfigsFromSecret(client *api.Client, mount string, clusterName string, nodeRole string, cloudProvider string, verbose bool) (data ConfigData, err error) {
	secretPath := fmt.Sprintf("%s/cluster-%s-%s", mount, clusterName, nodeRole)
	if verbose {
		fmt.Printf("Loading machine config from %s\n", secretPath)
	}

	secretData, secretErr := SecretData(client, secretPath, verbose)
	if secretErr != nil {
		err = errors.Wrapf(secretErr, "failed getting secret from path %s", secretPath)
		return data, err
	}

	nodeKey := fmt.Sprintf("node-%s.yaml", cloudProvider)

	c, ok := secretData["config.yaml"].(string)
	if !ok {
		err = errors.New("Could not extract bytes for config.yaml from secret")
		return data, err
	}

	data.TalosMachineConfig = []byte(c)

	p, ok := secretData["patch.yaml"].(string)
	if !ok {
		err = errors.New("Could not extract bytes for patch.yaml from secret")
		return data, err
	}

	data.TalosMachineConfigPatch = []byte(p)

	n, ok := secretData[nodeKey].(string)
	if !ok {
		err = errors.New(fmt.Sprintf("Could not extract bytes for %s from secret", nodeKey))
		return data, err
	}

	data.NodeConfig = []byte(n)

	zoneIDFromSecret, ok := secretData[CloudflareZoneIDEnvVar].(string)
	if ok {
		data.CloudflareZoneID = zoneIDFromSecret
	}

	cfTokenFromSecret, ok := secretData[CloudflareAPITokenEnvVar].(string)
	if ok {
		data.CloudflareAPIToken = cfTokenFromSecret
	}

	return data, err
}
