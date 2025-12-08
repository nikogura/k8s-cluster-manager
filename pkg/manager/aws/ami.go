package aws

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/pkg/errors"
)

const (
	// SideroLabsOwnerID is the AWS account ID for Sidero Labs (Talos creators).
	SideroLabsOwnerID = "540036508848"
	// TalosInstallerImagePrefix is the base image reference for Talos installers.
	TalosInstallerImagePrefix = "ghcr.io/siderolabs/installer"
)

// AWSImageDiscovery implements manager.ImageDiscovery for AWS.
type AWSImageDiscovery struct {
	EC2Client *ec2.Client
	Region    string
}

// DiscoverImage finds the Talos AMI for a specific version in the configured region.
func (a *AWSImageDiscovery) DiscoverImage(ctx context.Context, version string, region string) (imageID string, err error) {
	if region == "" {
		region = a.Region
	}

	// Clean version string (handle both "v1.10.8" and "1.10.8")
	cleanVersion := strings.TrimPrefix(version, "v")

	// Build AMI name pattern: talos-v1.10.8-{region}-amd64
	namePattern := fmt.Sprintf("talos-v%s-%s-amd64", cleanVersion, region)

	input := &ec2.DescribeImagesInput{
		Owners: []string{SideroLabsOwnerID},
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{namePattern},
			},
			{
				Name:   aws.String("state"),
				Values: []string{"available"},
			},
			{
				Name:   aws.String("architecture"),
				Values: []string{"x86_64"},
			},
		},
	}

	output, descErr := a.EC2Client.DescribeImages(ctx, input)
	if descErr != nil {
		err = errors.Wrapf(descErr, "failed to describe Talos AMIs for version %s", version)
		return imageID, err
	}

	if len(output.Images) == 0 {
		err = fmt.Errorf("no Talos AMI found for version %s in region %s", version, region)
		return imageID, err
	}

	// Return the first (and should be only) match
	imageID = *output.Images[0].ImageId

	return imageID, err
}

// GetInstallerImage returns the full installer image reference for a Talos version.
func (a *AWSImageDiscovery) GetInstallerImage(version string) (installerImage string) {
	// Clean version string
	cleanVersion := strings.TrimPrefix(version, "v")

	// Return full installer image reference
	installerImage = fmt.Sprintf("%s:v%s", TalosInstallerImagePrefix, cleanVersion)

	return installerImage
}

// DiscoverLatestPatchVersion finds the latest patch version for a given minor version.
func (a *AWSImageDiscovery) DiscoverLatestPatchVersion(ctx context.Context, minorVersion string, region string) (imageID string, fullVersion string, err error) {
	if region == "" {
		region = a.Region
	}

	// Clean version string (handle both "v1.10" and "1.10")
	cleanMinor := strings.TrimPrefix(minorVersion, "v")

	// Build AMI name pattern: talos-v1.10*-{region}-amd64
	namePattern := fmt.Sprintf("talos-v%s*-%s-amd64", cleanMinor, region)

	input := &ec2.DescribeImagesInput{
		Owners: []string{SideroLabsOwnerID},
		Filters: []types.Filter{
			{
				Name:   aws.String("name"),
				Values: []string{namePattern},
			},
			{
				Name:   aws.String("state"),
				Values: []string{"available"},
			},
			{
				Name:   aws.String("architecture"),
				Values: []string{"x86_64"},
			},
		},
	}

	output, descErr := a.EC2Client.DescribeImages(ctx, input)
	if descErr != nil {
		err = errors.Wrapf(descErr, "failed to describe Talos AMIs for minor version %s", minorVersion)
		return imageID, fullVersion, err
	}

	if len(output.Images) == 0 {
		err = fmt.Errorf("no Talos AMIs found for minor version %s in region %s", minorVersion, region)
		return imageID, fullVersion, err
	}

	// Sort images by creation date (newest first)
	sort.Slice(output.Images, sortByCreationDate(output.Images))

	// Get the newest image
	newestImage := output.Images[0]
	imageID = *newestImage.ImageId

	// Extract version from image name (e.g., "talos-v1.10.8-eu-west-2-amd64" -> "v1.10.8")
	nameParts := strings.Split(*newestImage.Name, "-")
	if len(nameParts) >= 2 {
		fullVersion = nameParts[1] // This should be "v1.10.8"
	} else {
		err = fmt.Errorf("unable to parse version from AMI name: %s", *newestImage.Name)
		return imageID, fullVersion, err
	}

	return imageID, fullVersion, err
}

// sortByCreationDate returns a comparison function for sorting images by creation date.
func sortByCreationDate(images []types.Image) (comparator func(i int, j int) (less bool)) {
	comparator = func(i int, j int) (less bool) {
		// Parse creation date strings directly
		iDate := aws.ToString(images[i].CreationDate)
		jDate := aws.ToString(images[j].CreationDate)
		less = iDate > jDate // ISO 8601 format allows string comparison (newest first)
		return less
	}
	return comparator
}

// ValidateTalosVersion checks if a version string is valid.
func ValidateTalosVersion(version string) (valid bool, err error) {
	// Clean version
	cleanVersion := strings.TrimPrefix(version, "v")

	// Split into parts
	parts := strings.Split(cleanVersion, ".")

	// Should have at least major.minor (patch is optional)
	if len(parts) < 2 {
		err = fmt.Errorf("invalid Talos version format: %s (expected format: v1.10.8 or 1.10.8)", version)
		return valid, err
	}

	valid = true
	return valid, err
}
