package aws

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{
			"one",
			`{
  "image_id": "ami-00000000000001111111",
  "subnet_id": "subnet-0babababababababaa7f",
  "instance_type": "blarg",
  "block_device_name": "/dev/xvda",
  "block_device_gb": 100,
  "block_device_type": "hd2",
  "placement_group_name": "blah"
}`,
		},
	}

	for _, tc := range cases {
		configFile := fmt.Sprintf("%s/%s.json", tmpDir, tc.name)
		err := os.WriteFile(configFile, []byte(tc.input), 0644)
		if err != nil {
			t.Errorf("failed writing config file: %s", err)
		}

		actual, err := LoadAWSNodeConfigFromFile(configFile)
		if err != nil {
			t.Errorf("failed loading config file: %s", err)
		}

		expected := AWSNodeConfig{}

		err = json.Unmarshal([]byte(tc.input), &expected)
		if err != nil {
			t.Errorf("failed unmarshalling test data: %s", err)
		}

		assert.True(t, reflect.DeepEqual(expected, actual), "Loaded config fails to meet expectations.")

	}

}
