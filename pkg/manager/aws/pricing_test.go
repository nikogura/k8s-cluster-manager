package aws

import (
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"math"
	"testing"
)

const floatTolerance = 0.0001

func TestAWSPricingEstimator_EstimateHourlyCost(t *testing.T) {
	tests := []struct {
		name          string
		region        string
		customPricing map[string]float64
		instanceType  string
		wantCost      float64
		wantErr       bool
		errorContains string
	}{
		{
			name:          "known instance type in us-east-1",
			region:        "us-east-1",
			customPricing: nil,
			instanceType:  "m5.large",
			wantCost:      0.096,
			wantErr:       false,
		},
		{
			name:          "known instance type in us-west-1 with multiplier",
			region:        "us-west-1",
			customPricing: nil,
			instanceType:  "m5.large",
			wantCost:      0.096 * 1.05, // us-west-1 has 1.05 multiplier
			wantErr:       false,
		},
		{
			name:          "custom pricing overrides base pricing",
			region:        "us-east-1",
			customPricing: map[string]float64{"m5.large": 0.08},
			instanceType:  "m5.large",
			wantCost:      0.08,
			wantErr:       false,
		},
		{
			name:          "unknown instance type",
			region:        "us-east-1",
			customPricing: nil,
			instanceType:  "unknown.large",
			wantCost:      0,
			wantErr:       true,
			errorContains: "no pricing data available",
		},
		{
			name:          "t3 micro instance",
			region:        "us-east-1",
			customPricing: nil,
			instanceType:  "t3.micro",
			wantCost:      0.0104,
			wantErr:       false,
		},
		{
			name:          "c5 xlarge instance",
			region:        "us-east-1",
			customPricing: nil,
			instanceType:  "c5.xlarge",
			wantCost:      0.17,
			wantErr:       false,
		},
		{
			name:          "r5 large instance",
			region:        "us-east-1",
			customPricing: nil,
			instanceType:  "r5.large",
			wantCost:      0.126,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimator := NewAWSPricingEstimator(tt.region, tt.customPricing)
			gotCost, err := estimator.EstimateHourlyCost(tt.instanceType)

			if (err != nil) != tt.wantErr {
				t.Errorf("EstimateHourlyCost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errorContains != "" {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorContains)
					return
				}
			}

			if !tt.wantErr && gotCost != tt.wantCost {
				t.Errorf("EstimateHourlyCost() gotCost = %v, want %v", gotCost, tt.wantCost)
			}
		})
	}
}

func TestAWSPricingEstimator_EstimateDailyCost(t *testing.T) {
	tests := []struct {
		name          string
		region        string
		customPricing map[string]float64
		instanceType  string
		wantCost      float64
		wantErr       bool
	}{
		{
			name:          "daily cost is 24x hourly",
			region:        "us-east-1",
			customPricing: nil,
			instanceType:  "m5.large",
			wantCost:      0.096 * 24,
			wantErr:       false,
		},
		{
			name:          "daily cost with custom pricing",
			region:        "us-east-1",
			customPricing: map[string]float64{"m5.large": 0.1},
			instanceType:  "m5.large",
			wantCost:      0.1 * 24,
			wantErr:       false,
		},
		{
			name:          "unknown instance type returns error",
			region:        "us-east-1",
			customPricing: nil,
			instanceType:  "unknown.large",
			wantCost:      0,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimator := NewAWSPricingEstimator(tt.region, tt.customPricing)
			gotCost, err := estimator.EstimateDailyCost(tt.instanceType)

			if (err != nil) != tt.wantErr {
				t.Errorf("EstimateDailyCost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && math.Abs(gotCost-tt.wantCost) > floatTolerance {
				t.Errorf("EstimateDailyCost() gotCost = %v, want %v", gotCost, tt.wantCost)
			}
		})
	}
}

func TestGetRegionMultiplier(t *testing.T) {
	tests := []struct {
		name       string
		region     string
		wantResult float64
	}{
		{
			name:       "us-east-1 baseline",
			region:     "us-east-1",
			wantResult: 1.0,
		},
		{
			name:       "us-west-1 multiplier",
			region:     "us-west-1",
			wantResult: 1.05,
		},
		{
			name:       "eu-west-1 multiplier",
			region:     "eu-west-1",
			wantResult: 1.05,
		},
		{
			name:       "ap-northeast-1 multiplier",
			region:     "ap-northeast-1",
			wantResult: 1.15,
		},
		{
			name:       "unknown region defaults to 1.0",
			region:     "unknown-region",
			wantResult: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getRegionMultiplier(tt.region)
			if got != tt.wantResult {
				t.Errorf("getRegionMultiplier() = %v, want %v", got, tt.wantResult)
			}
		})
	}
}

func TestCalculateClusterDailyCost(t *testing.T) {
	customPricing := map[string]float64{
		"m5.large":  0.1,
		"c5.xlarge": 0.2,
	}

	tests := []struct {
		name      string
		nodes     []manager.NodeInfo
		estimator manager.CostEstimator
		wantCost  float64
		wantErr   bool
	}{
		{
			name: "calculate cost for multiple nodes",
			nodes: []manager.NodeInfo{
				{Name: "node1", ID: "i-123", InstanceType: "m5.large"},
				{Name: "node2", ID: "i-456", InstanceType: "m5.large"},
				{Name: "node3", ID: "i-789", InstanceType: "c5.xlarge"},
			},
			estimator: NewAWSPricingEstimator("us-east-1", customPricing),
			wantCost:  (0.1 * 24 * 2) + (0.2 * 24), // 2 m5.large + 1 c5.xlarge
			wantErr:   false,
		},
		{
			name: "skip nodes without instance type",
			nodes: []manager.NodeInfo{
				{Name: "node1", ID: "i-123", InstanceType: "m5.large"},
				{Name: "node2", ID: "i-456", InstanceType: ""},
				{Name: "node3", ID: "i-789", InstanceType: "c5.xlarge"},
			},
			estimator: NewAWSPricingEstimator("us-east-1", customPricing),
			wantCost:  (0.1 * 24) + (0.2 * 24), // Only 2 nodes with types
			wantErr:   false,
		},
		{
			name: "skip nodes with unknown instance types",
			nodes: []manager.NodeInfo{
				{Name: "node1", ID: "i-123", InstanceType: "m5.large"},
				{Name: "node2", ID: "i-456", InstanceType: "unknown.type"},
				{Name: "node3", ID: "i-789", InstanceType: "c5.xlarge"},
			},
			estimator: NewAWSPricingEstimator("us-east-1", customPricing),
			wantCost:  (0.1 * 24) + (0.2 * 24), // Skip unknown type
			wantErr:   false,
		},
		{
			name:      "nil estimator returns error",
			nodes:     []manager.NodeInfo{{Name: "node1", ID: "i-123", InstanceType: "m5.large"}},
			estimator: nil,
			wantCost:  0,
			wantErr:   true,
		},
		{
			name:      "empty node list returns zero cost",
			nodes:     []manager.NodeInfo{},
			estimator: NewAWSPricingEstimator("us-east-1", customPricing),
			wantCost:  0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCost, err := CalculateClusterDailyCost(tt.nodes, tt.estimator)

			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateClusterDailyCost() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && math.Abs(gotCost-tt.wantCost) > floatTolerance {
				t.Errorf("CalculateClusterDailyCost() gotCost = %v, want %v", gotCost, tt.wantCost)
			}
		})
	}
}

func TestNewAWSPricingEstimator(t *testing.T) {
	customPricing := map[string]float64{"m5.large": 0.1}

	estimator := NewAWSPricingEstimator("us-west-2", customPricing)

	if estimator == nil {
		t.Fatal("NewAWSPricingEstimator() returned nil")
	}

	if estimator.Region != "us-west-2" {
		t.Errorf("NewAWSPricingEstimator() Region = %v, want %v", estimator.Region, "us-west-2")
	}

	if estimator.CustomPricing == nil {
		t.Error("NewAWSPricingEstimator() CustomPricing is nil")
	}

	if len(estimator.CustomPricing) != 1 {
		t.Errorf("NewAWSPricingEstimator() CustomPricing length = %v, want 1", len(estimator.CustomPricing))
	}
}
