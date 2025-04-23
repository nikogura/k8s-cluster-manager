package cloudflare

import (
	"context"
	"fmt"
	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/dns"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/nikogura/k8s-cluster-manager/pkg/manager"
	"github.com/pkg/errors"
	"os"
)

func RegisterNode(ctx context.Context, node manager.ClusterNode, verbose bool) (err error) {
	manager.VerboseOutput(verbose, "Registering DNS for node\n")

	zoneID, ok := os.LookupEnv("CLOUDFLARE_ZONE_ID")
	if !ok {
		err = errors.New("no CLOUDFLARE_ZONE_ID in environment.  Cannot provision DNS.")
		return err
	}

	client := cloudflare.NewClient(
		option.WithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN")),
	)

	params := dns.RecordNewParams{
		ZoneID: cloudflare.F(zoneID),
		Record: dns.ARecordParam{
			Content: cloudflare.F(node.IP()),
			Name:    cloudflare.F(fmt.Sprintf("%s.%s", node.Name(), node.Domain())),
			Type:    cloudflare.F(dns.ARecordTypeA),
		},
	}

	_, err = client.DNS.Records.New(context.TODO(), params)
	if err != nil {
		err = errors.Wrapf(err, "failed setting dns record for %s", node.Name())
		return err
	}

	return err
}

func DeRegisterNode(ctx context.Context, nodeName string, verbose bool) (err error) {
	fmt.Printf("TODO: cloudflare.DeRegisterNode() Registering DNS for node\n")
	zoneID, ok := os.LookupEnv("CLOUDFLARE_ZONE_ID")

	if !ok {
		err = errors.New("no CLOUDFLARE_ZONE_ID in environment.  Cannot provision DNS.")
		return err
	}
	client := cloudflare.NewClient(
		option.WithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN")),
	)

	listParams := dns.RecordListParams{
		ZoneID: cloudflare.F(zoneID),
		//Comment:interface{}[dns.RecordListParamsComment]{},
		//Content:interface{}[dns.RecordListParamsContent]{},
		//Direction:interface{}[shared.SortDirection]{},
		//Match:interface{}[dns.RecordListParamsMatch]{},
		//Name:interface{}[dns.RecordListParamsName]{},
		Name: cloudflare.F(dns.RecordListParamsName{
			Contains: cloudflare.F(nodeName),
		}),
		//Order:interface{}[dns.RecordListParamsOrder]{},
		//Page:interface{}[float64]{},
		//PerPage:interface{}[float64]{},
		//Proxied:interface{}[bool]{},
		//Search:interface{}[string]{},
		//Tag:interface{}[dns.RecordListParamsTag]{},
		//TagMatch:interface{}[dns.RecordListParamsTagMatch]{},
		//Type:interface{}[dns.RecordListParamsType]{},
	}

	resp, err := client.DNS.Records.List(ctx, listParams)
	if err != nil {
		err = errors.Wrapf(err, "failed listing DNS records in zone.")
		return err
	}

	fmt.Printf("%d results\n", len(resp.Result))

	deleteParams := dns.RecordDeleteParams{
		ZoneID: cloudflare.F(zoneID),
	}

	for _, record := range resp.Result {
		fmt.Printf("%s %s\n", record.Name, record.ID)

		_, err = client.DNS.Records.Delete(ctx, record.ID, deleteParams)
		if err != nil {
			err = errors.Wrapf(err, "failed setting dns record %s for %s", record.ID, record.Name)
			return err
		}
	}

	return err
}
