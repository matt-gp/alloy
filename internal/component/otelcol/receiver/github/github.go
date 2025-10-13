package github

import (
	"time"
	"unsafe"

	"github.com/grafana/alloy/internal/component"
	"github.com/grafana/alloy/internal/component/otelcol"
	otelcolCfg "github.com/grafana/alloy/internal/component/otelcol/config"
	"github.com/grafana/alloy/internal/component/otelcol/extension"
	"github.com/grafana/alloy/internal/component/otelcol/receiver"
	"github.com/grafana/alloy/internal/featuregate"
	"github.com/mitchellh/mapstructure"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/githubreceiver"
	otelcomponent "go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pipeline"
	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

func init() {
	component.Register(component.Registration{
		Name:      "otelcol.receiver.github",
		Stability: featuregate.StabilityExperimental,
		Args:      Arguments{},

		Build: func(opts component.Options, args component.Arguments) (component.Component, error) {
			fact := githubreceiver.NewFactory()
			return receiver.New(opts, fact, args.(Arguments))
		},
	})
}

// Arguments configures the otelcol.receiver.github component.
type Arguments struct {
	InitialDelay       time.Duration                    `alloy:"initial_delay,attr,optional"`
	CollectionInterval time.Duration                    `alloy:"collection_interval,attr,optional"`
	Scraper            *ScraperConfig                   `alloy:"scraper,block,optional"`
	Webhook            *WebhookConfig                   `alloy:"webhook,block,optional"`
	Storage            *extension.ExtensionHandler      `alloy:"storage,attr,optional"`
	DebugMetrics       otelcolCfg.DebugMetricsArguments `alloy:"debug_metrics,block,optional"`

	// Output configures where to send received data. Required.
	Output *otelcol.ConsumerArguments `alloy:"output,block"`
}

var _ receiver.Arguments = Arguments{}

func (args *Arguments) Validate() error {
	if args.Scraper != nil {
		if err := args.Scraper.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func (args *Arguments) SetToDefault() {
	if args.InitialDelay == 0 {
		args.InitialDelay = 0
	}
	if args.CollectionInterval == 0 {
		args.CollectionInterval = 30 * time.Second
	}

	if args.Scraper != nil {
		args.Scraper.SetToDefault()
	}

	if args.Webhook != nil {
		args.Webhook.SetToDefault()
	}

	args.DebugMetrics.SetToDefault()
}

func (args Arguments) Convert() (otelcomponent.Config, error) {
	config := &githubreceiver.Config{
		ControllerConfig: scraperhelper.ControllerConfig{
			InitialDelay:       args.InitialDelay,
			CollectionInterval: args.CollectionInterval,
		},
	}

	if args.Scraper != nil {
		scrapers := map[string]interface{}{
			"scraper": args.Scraper.Convert(),
		}
		*(*map[string]interface{})(unsafe.Pointer(&config.Scrapers)) = scrapers
	}

	if args.Webhook != nil {
		err := mapstructure.Decode(args.Webhook.Convert(), &config.WebHook)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

// Extensions implements receiver.Arguments.
func (args Arguments) Extensions() map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// Exporters implements receiver.Arguments.
func (args Arguments) Exporters() map[pipeline.Signal]map[otelcomponent.ID]otelcomponent.Component {
	return nil
}

// NextConsumers implements receiver.Arguments.
func (args Arguments) NextConsumers() *otelcol.ConsumerArguments {
	return args.Output
}

// DebugMetricsConfig implements receiver.Arguments.
func (args Arguments) DebugMetricsConfig() otelcolCfg.DebugMetricsArguments {
	return args.DebugMetrics
}
