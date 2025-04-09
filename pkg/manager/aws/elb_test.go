package aws

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTargetGroupName(t *testing.T) {
	cases := []struct {
		name     string
		tls      bool
		expected string
	}{
		{
			"foo",
			false,
			"ingress-foo-clear",
		},
		{
			"bar",
			true,
			"ingress-bar-tls",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := TargetGroupName(tc.name, tc.tls)
			assert.Equal(t, tc.expected, actual, "actual target group name fails to meet expectations")
		})
	}

}

func TestLoadBalancerName(t *testing.T) {
	cases := []struct {
		name     string
		lbType   string
		expected string
	}{
		{
			"foo",
			"apiserver",
			"apiserver-foo",
		},
		{
			"bar",
			"int",
			"ingress-bar",
		},
		{
			"baz",
			"ext",
			"ingress-baz-ext",
		},
	}

	for _, tc := range cases {
		actual, err := LoadBalancerName(tc.name, tc.lbType)
		if err != nil {
			t.Errorf("failed generating lb name: %s", err)
		}
		assert.Equal(t, tc.expected, actual, "actual load balancer name does not meet expectations")
	}
}
