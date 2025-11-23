package aws

import (
	"context"
	"errors"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	promcfg "github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	promaws "github.com/prometheus/prometheus/discovery/aws"

	"github.com/grafana/alloy/internal/component"
	"github.com/grafana/alloy/internal/component/common/config"
	"github.com/grafana/alloy/internal/component/discovery"
	"github.com/grafana/alloy/internal/featuregate"
	"github.com/grafana/alloy/syntax/alloytypes"
)

func init() {
	component.Register(component.Registration{
		Name:      "discovery.aws",
		Stability: featuregate.StabilityGenerallyAvailable,
		Args:      AWSArguments{},
		Exports:   discovery.Exports{},
		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			return discovery.NewFromConvertibleConfig(opts, args.(AWSArguments))
		},
	})
}

// AWSArguments is the configuration for AWS based service discovery.
type AWSArguments struct {
	Endpoint        string            `alloy:"endpoint,attr,optional"`
	Region          string            `alloy:"region,attr,optional"`
	AccessKey       string            `alloy:"access_key,attr,optional"`
	SecretKey       alloytypes.Secret `alloy:"secret_key,attr,optional"`
	Profile         string            `alloy:"profile,attr,optional"`
	RoleARN         string            `alloy:"role_arn,attr,optional"`
	RefreshInterval time.Duration     `alloy:"refresh_interval,attr,optional"`
	Port            int               `alloy:"port,attr,optional"`

	// aws role (e.g. ec2, ecs, msk)
	Role string `alloy:"role,attr,optional"`

	// ec2 specific
	Filters []*EC2Filter `alloy:"filter,block,optional"`

	// ecs specific
	Clusters []string `alloy:"cluster,attr,optional"`

	// msk specific
	ClusterNameFilter string `alloy:"cluster_name_filter,attr,optional"`

	HTTPClientConfig config.HTTPClientConfig `alloy:",squash"`
}

func (args AWSArguments) Convert() discovery.DiscovererConfig {
	cfg := &promaws.AWSSDConfig{
		Endpoint:          args.Endpoint,
		Region:            args.Region,
		AccessKey:         args.AccessKey,
		SecretKey:         promcfg.Secret(args.SecretKey),
		Profile:           args.Profile,
		RoleARN:           args.RoleARN,
		RefreshInterval:   model.Duration(args.RefreshInterval),
		Port:              args.Port,
		Role:              args.Role,
		Clusters:          args.Clusters,
		ClusterNameFilter: args.ClusterNameFilter,
		HTTPClientConfig:  *args.HTTPClientConfig.Convert(),
	}
	for _, f := range args.Filters {
		cfg.Filters = append(cfg.Filters, &promaws.EC2Filter{
			Name:   f.Name,
			Values: f.Values,
		})
	}
	return cfg
}

var DefaultAWSSDConfig = AWSArguments{
	Port:             80,
	RefreshInterval:  60 * time.Second,
	HTTPClientConfig: config.DefaultHTTPClientConfig,
}

// SetToDefault implements syntax.Defaulter.
func (args *AWSArguments) SetToDefault() {
	*args = DefaultAWSSDConfig
}

// Validate implements syntax.Validator.
func (args *AWSArguments) Validate() error {
	if args.Region == "" {
		cfgCtx := context.TODO()
		cfg, err := awsConfig.LoadDefaultConfig(cfgCtx)
		if err != nil {
			return err
		}

		client := imds.NewFromConfig(cfg)
		region, err := client.GetRegion(cfgCtx, &imds.GetRegionInput{})
		if err != nil {
			return errors.New("AWS SD configuration requires a region")
		}
		args.Region = region.Region
	}
	return nil
}
