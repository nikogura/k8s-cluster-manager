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

// VaultApiConfig creates a vault api config in a standard fashion
func VaultApiConfig(address string) (config *api.Config, err error) {
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

	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		err = errors.Wrapf(err, "failed to get system cert pool")
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
	apiConfig, err := VaultApiConfig(os.Getenv("VAULT_ADDR"))

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

	v2Path, err := V2Path(path)
	if err != nil {
		err = errors.Wrapf(err, "failed creating v2 secret path from %q", path)
		return data, err
	}

	s, err := client.Logical().Read(v2Path)
	if err != nil {
		err = errors.Wrapf(err, "Failed to lookup path: %s", path)
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

func ConfigsFromSecret(client *api.Client, mount string, clusterName string, nodeRole string, cloudProvider string, verbose bool) (config []byte, patch []byte, node []byte, err error) {
	secretPath := fmt.Sprintf("%s/cluster-%s-%s", mount, clusterName, nodeRole)
	if verbose {
		fmt.Printf("Loading machine config from %s\n", secretPath)
	}

	secretData, err := SecretData(client, secretPath, verbose)
	if err != nil {
		err = errors.Wrapf(err, "failed getting secret from path %s", secretPath)
		return config, patch, node, err
	}

	nodeKey := fmt.Sprintf("node-%s.yaml", cloudProvider)

	c, ok := secretData["config.yaml"].(string)
	if !ok {
		err = errors.New("Could not extract bytes for config.yaml from secret")
		return config, patch, node, err
	}

	config = []byte(c)

	p, ok := secretData["patch.yaml"].(string)
	if !ok {
		err = errors.New("Could not extract bytes for patch.yaml from secret")
		return config, patch, node, err
	}

	patch = []byte(p)

	n, ok := secretData[nodeKey].(string)
	if !ok {
		err = errors.New(fmt.Sprintf("Could not extract bytes for %s from secret", nodeKey))
		return config, patch, node, err
	}

	node = []byte(n)

	//config, err = yaml.Marshal(secretData["config.yaml"])
	//if err != nil {
	//	err = errors.Wrapf(err, "failed marshalling secret data from path %s", secretPath)
	//	return config, patch, node, err
	//}
	//
	//patch, err = yaml.Marshal(secretData["patch.yaml"])
	//if err != nil {
	//	err = errors.Wrapf(err, "failed marshalling secret data from path %s", secretPath)
	//	return config, patch, node, err
	//}
	//
	//node, err = yaml.Marshal(secretData[nodeKey])
	//if err != nil {
	//	err = errors.Wrapf(err, "failed marshalling secret data from path %s", secretPath)
	//	return config, patch, node, err
	//}

	return config, patch, node, err
}
