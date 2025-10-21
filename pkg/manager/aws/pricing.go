package aws

import (
	"fmt"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/pkg/errors"
)

// InstanceSpecs holds CPU and memory specifications for an instance type.
type InstanceSpecs struct {
	VCPUs     int
	MemoryGiB float64
}

// getInstanceSpecs returns CPU and memory specifications for AWS instance types.
func getInstanceSpecs(instanceType string) (specs InstanceSpecs, err error) {
	specsMap := map[string]InstanceSpecs{
		// General Purpose (T family)
		"t3.nano":   {VCPUs: 2, MemoryGiB: 0.5},
		"t3.micro":  {VCPUs: 2, MemoryGiB: 1},
		"t3.small":  {VCPUs: 2, MemoryGiB: 2},
		"t3.medium": {VCPUs: 2, MemoryGiB: 4},
		"t3.large":  {VCPUs: 2, MemoryGiB: 8},
		"t3.xlarge": {VCPUs: 4, MemoryGiB: 16},

		// General Purpose (M5 family)
		"m5.large":   {VCPUs: 2, MemoryGiB: 8},
		"m5.xlarge":  {VCPUs: 4, MemoryGiB: 16},
		"m5.2xlarge": {VCPUs: 8, MemoryGiB: 32},
		"m5.4xlarge": {VCPUs: 16, MemoryGiB: 64},
		"m5.8xlarge": {VCPUs: 32, MemoryGiB: 128},

		// Compute Optimized (C5 family)
		"c5.large":   {VCPUs: 2, MemoryGiB: 4},
		"c5.xlarge":  {VCPUs: 4, MemoryGiB: 8},
		"c5.2xlarge": {VCPUs: 8, MemoryGiB: 16},
		"c5.4xlarge": {VCPUs: 16, MemoryGiB: 32},
		"c5.9xlarge": {VCPUs: 36, MemoryGiB: 72},

		// Memory Optimized (R5 family)
		"r5.large":   {VCPUs: 2, MemoryGiB: 16},
		"r5.xlarge":  {VCPUs: 4, MemoryGiB: 32},
		"r5.2xlarge": {VCPUs: 8, MemoryGiB: 64},
		"r5.4xlarge": {VCPUs: 16, MemoryGiB: 128},
		"r5.8xlarge": {VCPUs: 32, MemoryGiB: 256},
	}

	var exists bool
	specs, exists = specsMap[instanceType]
	if !exists {
		err = fmt.Errorf("no instance specs available for instance type %s", instanceType)
		return specs, err
	}

	return specs, err
}

// GetInstanceSpecs returns CPU and memory specs for an instance type (exported for external use).
func GetInstanceSpecs(instanceType string) (vcpus int, memoryGiB float64, err error) {
	specs, specsErr := getInstanceSpecs(instanceType)
	if specsErr != nil {
		err = specsErr
		return vcpus, memoryGiB, err
	}

	vcpus = specs.VCPUs
	memoryGiB = specs.MemoryGiB
	return vcpus, memoryGiB, err
}

// AWSPricingEstimator provides cost estimation for AWS EC2 instances.
type AWSPricingEstimator struct {
	Region        string
	CustomPricing map[string]float64 // Optional: override pricing for specific instance types
}

// NewAWSPricingEstimator creates a new AWS pricing estimator.
func NewAWSPricingEstimator(region string, customPricing map[string]float64) (estimator *AWSPricingEstimator) {
	estimator = &AWSPricingEstimator{
		Region:        region,
		CustomPricing: customPricing,
	}
	return estimator
}

// EstimateHourlyCost returns the estimated cost per hour for the given instance type in USD.
func (e *AWSPricingEstimator) EstimateHourlyCost(instanceType string) (costPerHour float64, err error) {
	// Check custom pricing first
	if e.CustomPricing != nil {
		if cost, exists := e.CustomPricing[instanceType]; exists {
			costPerHour = cost
			return costPerHour, err
		}
	}

	// Fall back to embedded pricing table
	costPerHour, err = getAWSOnDemandPrice(e.Region, instanceType)
	return costPerHour, err
}

// EstimateDailyCost returns the estimated cost per day (24 hours) for the given instance type in USD.
func (e *AWSPricingEstimator) EstimateDailyCost(instanceType string) (costPerDay float64, err error) {
	hourly, hourlyErr := e.EstimateHourlyCost(instanceType)
	if hourlyErr != nil {
		err = errors.Wrapf(hourlyErr, "failed getting hourly cost for instance type %s", instanceType)
		return costPerDay, err
	}

	costPerDay = hourly * 24
	return costPerDay, err
}

// getAWSOnDemandPrice returns the on-demand hourly price for a given instance type and region.
// This is a simplified implementation. For production use, consider:
// - AWS Pricing API integration.
// - Regular updates of pricing data.
// - Support for reserved instances, savings plans, spot pricing.
func getAWSOnDemandPrice(region, instanceType string) (price float64, err error) {
	// Pricing data for us-east-1 (as of 2025, approximate values)
	// This is intentionally minimal - users should provide custom pricing
	basePricing := map[string]float64{
		// General Purpose (T family)
		"t3.nano":   0.0052,
		"t3.micro":  0.0104,
		"t3.small":  0.0208,
		"t3.medium": 0.0416,
		"t3.large":  0.0832,
		"t3.xlarge": 0.1664,

		// General Purpose (M5 family)
		"m5.large":   0.096,
		"m5.xlarge":  0.192,
		"m5.2xlarge": 0.384,
		"m5.4xlarge": 0.768,
		"m5.8xlarge": 1.536,

		// Compute Optimized (C5 family)
		"c5.large":   0.085,
		"c5.xlarge":  0.17,
		"c5.2xlarge": 0.34,
		"c5.4xlarge": 0.68,
		"c5.9xlarge": 1.53,

		// Memory Optimized (R5 family)
		"r5.large":   0.126,
		"r5.xlarge":  0.252,
		"r5.2xlarge": 0.504,
		"r5.4xlarge": 1.008,
		"r5.8xlarge": 2.016,
	}

	var exists bool
	price, exists = basePricing[instanceType]
	if !exists {
		err = fmt.Errorf("no pricing data available for instance type %s in region %s (provide custom pricing or use AWS Pricing API)", instanceType, region)
		return price, err
	}

	// Region pricing multipliers (approximate, us-east-1 is baseline 1.0)
	regionMultiplier := getRegionMultiplier(region)
	price = price * regionMultiplier

	return price, err
}

// getRegionMultiplier returns a pricing multiplier for the given region.
// us-east-1 is the baseline (1.0). Other regions have approximate multipliers.
func getRegionMultiplier(region string) (multiplier float64) {
	multipliers := map[string]float64{
		"us-east-1":      1.0,
		"us-east-2":      1.0,
		"us-west-1":      1.05,
		"us-west-2":      1.0,
		"eu-west-1":      1.05,
		"eu-west-2":      1.05,
		"eu-central-1":   1.05,
		"ap-southeast-1": 1.1,
		"ap-southeast-2": 1.15,
		"ap-northeast-1": 1.15,
	}

	var exists bool
	multiplier, exists = multipliers[region]
	if !exists {
		multiplier = 1.0 // Default to us-east-1 pricing
	}

	return multiplier
}

// CalculateClusterDailyCost calculates the total estimated daily cost for all nodes in the cluster.
func CalculateClusterDailyCost(nodes []manager.NodeInfo, estimator manager.CostEstimator) (totalDailyCost float64, err error) {
	if estimator == nil {
		err = errors.New("cost estimator is nil")
		return totalDailyCost, err
	}

	for _, node := range nodes {
		if node.InstanceType == "" {
			continue // Skip nodes without instance type
		}

		dailyCost, costErr := estimator.EstimateDailyCost(node.InstanceType)
		if costErr != nil {
			// Log warning but continue with other nodes
			// Don't fail the entire calculation if one instance type is unknown
			continue
		}

		totalDailyCost += dailyCost
	}

	return totalDailyCost, err
}
