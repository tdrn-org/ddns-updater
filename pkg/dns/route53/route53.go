// route53.go
//
// Copyright (C) 2023-2024 Holger de Carne
//
// This software may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.

package route53

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/tdrn-org/ddns-updater/internal/logging"
	"github.com/tdrn-org/ddns-updater/pkg/dns"
	"github.com/rs/zerolog"
)

const Name = "dns_route53"

type Route53UpdaterConfig struct {
	dns.UpdaterConfig
	Key    string `toml:"key"`
	Secret string `toml:"secret"`
	Region string `tom:"region"`
}

func NewRoute53Updater(cfg *Route53UpdaterConfig) dns.Updater {
	name := Name
	logger := logging.RootLogger().With().Str("updater", name).Logger()
	return &route53Updater{cfg: cfg, name: name, logger: logger}
}

type route53Updater struct {
	cfg    *Route53UpdaterConfig
	name   string
	logger zerolog.Logger
}

func (u *route53Updater) Name() string {
	return u.name
}

func (u *route53Updater) Merge(ips []net.IP, force bool, pretend bool) error {
	u.logger.Info().Msgf("Updating host '%s'...", u.cfg.Host)
	options := route53.Options{
		Credentials: credentials.NewStaticCredentialsProvider(u.cfg.Key, u.cfg.Secret, ""),
		Region:      u.cfg.Region,
	}
	ctx := context.Background()
	client := route53.New(options)
	zone, err := u.lookupZone(ctx, client)
	if err != nil {
		return err
	}
	rrss, err := u.lookupRRSs(ctx, client, zone)
	if err != nil {
		return err
	}
	rrss, err = u.mergeRRSs(rrss, ips)
	if err != nil {
		return err
	}
	return u.updateRRSs(ctx, client, zone, rrss, pretend)
}

func (u *route53Updater) lookupZone(ctx context.Context, client *route53.Client) (types.HostedZone, error) {
	fqHost := u.cfg.Host + "."
	lhzInput := &route53.ListHostedZonesInput{}
	for {
		lhzOutput, err := client.ListHostedZones(ctx, lhzInput)
		if err != nil {
			return types.HostedZone{}, fmt.Errorf("failed to list hosted zones\n\tcause: %v", err)
		}
		for _, zone := range lhzOutput.HostedZones {
			if strings.HasSuffix(fqHost, "."+*zone.Name) {
				return zone, nil
			}
		}
		if !lhzOutput.IsTruncated {
			break
		}
		lhzInput.Marker = lhzOutput.NextMarker
	}
	return types.HostedZone{}, fmt.Errorf("no hosted zone found matching host '%s'", u.cfg.Host)
}

func (u *route53Updater) lookupRRSs(ctx context.Context, client *route53.Client, zone types.HostedZone) ([]types.ResourceRecordSet, error) {
	fqHost := u.cfg.Host + "."
	rrss := make([]types.ResourceRecordSet, 0)
	lrrssInput := &route53.ListResourceRecordSetsInput{HostedZoneId: zone.Id, StartRecordName: &fqHost}
	for {
		lrrssOutput, err := client.ListResourceRecordSets(ctx, lrrssInput)
		if err != nil {
			return rrss, fmt.Errorf("failed to list resource record sets\n\tcause: %v", err)
		}
		for _, rrs := range lrrssOutput.ResourceRecordSets {
			if *rrs.Name != fqHost {
				break
			}
			if rrs.Type == types.RRTypeA || rrs.Type == types.RRTypeAaaa {
				rrss = append(rrss, rrs)
			}
		}
		if !lrrssOutput.IsTruncated {
			break
		}
		lrrssInput.StartRecordIdentifier = lrrssOutput.NextRecordIdentifier
	}
	return rrss, nil
}

func (u *route53Updater) mergeRRSs(rrss []types.ResourceRecordSet, ips []net.IP) ([]types.ResourceRecordSet, error) {
	merged := make([]types.ResourceRecordSet, 0)
	for _, rrs := range rrss {
		if rrs.Type == types.RRTypeA {
			rrs.ResourceRecords = make([]types.ResourceRecord, 0)
			for _, ip := range ips {
				if len(ip) == 4 {
					ipString := ip.String()
					rrs.ResourceRecords = append(rrs.ResourceRecords, types.ResourceRecord{Value: &ipString})
				}
			}
			if len(rrs.ResourceRecords) == 0 {
				return merged, fmt.Errorf("no A records to update")
			}
			merged = append(merged, rrs)
		} else if rrs.Type == types.RRTypeAaaa {
			rrs.ResourceRecords = make([]types.ResourceRecord, 0)
			for _, ip := range ips {
				if len(ip) == 16 {
					ipString := ip.String()
					rrs.ResourceRecords = append(rrs.ResourceRecords, types.ResourceRecord{Value: &ipString})
				}
			}
			if len(rrs.ResourceRecords) == 0 {
				return merged, fmt.Errorf("no AAAA records to update")
			}
			merged = append(merged, rrs)
		}
	}
	return merged, nil
}

func (u *route53Updater) updateRRSs(ctx context.Context, client *route53.Client, zone types.HostedZone, rrss []types.ResourceRecordSet, pretend bool) error {
	changeBatch := &types.ChangeBatch{Changes: make([]types.Change, len(rrss))}
	for i, rrs := range rrss {
		for _, rr := range rrs.ResourceRecords {
			fmt.Printf("DDNS update '%s' %s %s\n", u.cfg.Host, rrs.Type, *rr.Value)
		}
		changeBatch.Changes[i] = types.Change{
			Action: types.ChangeActionUpsert,
			ResourceRecordSet: &types.ResourceRecordSet{
				Name:            rrs.Name,
				Type:            rrs.Type,
				ResourceRecords: rrs.ResourceRecords,
				TTL:             rrs.TTL},
		}
	}
	crrssInput := &route53.ChangeResourceRecordSetsInput{ChangeBatch: changeBatch, HostedZoneId: zone.Id}
	if !pretend {
		crrssOutput, err := client.ChangeResourceRecordSets(ctx, crrssInput)
		if err != nil {
			return fmt.Errorf("changing resource record sets for zone id %s failed\t\ncause: %v", *zone.Id, err)
		}
		u.logger.Info().Msgf("DDNS update applied (%s:%s)", *crrssOutput.ChangeInfo.Id, crrssOutput.ChangeInfo.Status)
	} else {
		u.logger.Warn().Msg("DDNS update skipped due to pretend")
	}
	return nil
}
