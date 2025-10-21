package cloudflare

import (
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/pkg/errors"
)

type CloudFlareManager struct {
	apiToken string
	zoneID   string
}

func NewCloudFlareManager(zoneID, apiToken string) (manager CloudFlareManager) {
	manager = CloudFlareManager{
		apiToken: apiToken,
		zoneID:   zoneID,
	}

	return manager
}

func (c CloudFlareManager) RegisterNode(ctx context.Context, node manager.ClusterNode, verbose bool) (err error) {
	manager.VerboseOutput(verbose, "Registering DNS for node\n")

	client := cloudflare.NewClient(
		option.WithAPIToken(c.apiToken),
	)

	aRecord := dns.ARecordParam{
		Content: cloudflare.F(node.IP()),
		Name:    cloudflare.F(fmt.Sprintf("%s.%s", node.Name(), node.Domain())),
		Type:    cloudflare.F(dns.ARecordTypeA),
	}

	params := dns.RecordNewParams{
		ZoneID: cloudflare.F(c.zoneID),
		Record: aRecord,
	}

	_, err = client.DNS.Records.New(ctx, params)
	if err != nil {
		err = errors.Wrapf(err, "failed setting dns record for %s", node.Name())
		return err
	}

	return err
}

func (c CloudFlareManager) DeregisterNode(ctx context.Context, nodeName string, verbose bool) (err error) {
	manager.VerboseOutput(verbose, "Deregistering DNS for node\n")

	client := cloudflare.NewClient(
		option.WithAPIToken(c.apiToken),
	)

	listParams := dns.RecordListParams{
		ZoneID: cloudflare.F(c.zoneID),
		Name: cloudflare.F(dns.RecordListParamsName{
			Contains: cloudflare.F(nodeName),
		}),
	}

	resp, listErr := client.DNS.Records.List(ctx, listParams)
	if listErr != nil {
		err = errors.Wrapf(listErr, "failed listing DNS records in zone.")
		return err
	}

	deleteParams := dns.RecordDeleteParams{
		ZoneID: cloudflare.F(c.zoneID),
	}

	for _, record := range resp.Result {
		_, err = client.DNS.Records.Delete(ctx, record.ID, deleteParams)
		if err != nil {
			err = errors.Wrapf(err, "failed deleting DNS record %s for %s", record.ID, record.Name)
			return err
		}
	}

	return err
}
