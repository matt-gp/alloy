package github_test

import (
	"testing"
	"time"

	otelcolCfg "github.com/grafana/alloy/internal/component/otelcol/config"
	"github.com/grafana/alloy/internal/component/otelcol/receiver/github"
	"github.com/grafana/alloy/syntax"
	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/githubreceiver"
	"github.com/stretchr/testify/require"
)

func TestArguments_UnmarshalAlloy(t *testing.T) {
	tests := []struct {
		testName string
		cfg      string
		validate func(t *testing.T, cfg *githubreceiver.Config)
	}{
		{
			testName: "minimal scraper configuration",
			cfg: `
				collection_interval = "60s"
				scraper {
					github_org = "my-org"
					search_query = "is:pr is:open"
					auth {
						authenticator = "my_auth"
				output {}
			`,
			validate: func(t *testing.T, cfg *githubreceiver.Config) {
				require.NotNil(t, cfg)
				require.Equal(t, 60*time.Second, cfg.ControllerConfig.CollectionInterval)
				require.Equal(t, time.Duration(0), cfg.ControllerConfig.InitialDelay)
				require.NotNil(t, cfg.Scrapers)
			},
		},
		{
			testName: "full scraper configuration with metrics",
			cfg: `
				initial_delay = "1s"
				collection_interval = "30s"
				scraper {
					github_org = "grafana"
					search_query = "is:pr is:merged"
					endpoint = "https://api.github.com"
					auth {
						authenticator = "github_auth"
					metrics {
						vcs.change.count {
							enabled = true
						}
						vcs.change.duration {
							enabled = true
						}
						vcs.change.time_to_approval {
							enabled = false
						}
						vcs.change.time_to_merge {
							enabled = true
						}
						vcs.ref.count {
							enabled = true
						}
						vcs.ref.lines_delta {
							enabled = false
						}
						vcs.ref.revisions_delta {
							enabled = true
						}
						vcs.ref.time {
							enabled = true
						}
						vcs.repository.count {
							enabled = true
						}
						vcs.contributor.count {
							enabled = false
						}
				output {}
			`,
			validate: func(t *testing.T, cfg *githubreceiver.Config) {
				require.NotNil(t, cfg)
				require.Equal(t, 1*time.Second, cfg.ControllerConfig.InitialDelay)
				require.Equal(t, 30*time.Second, cfg.ControllerConfig.CollectionInterval)
				require.NotNil(t, cfg.Scrapers)
			},
		},
		{
			testName: "webhook configuration",
			cfg: `
					collection_interval = "30s"
					scraper {
						github_org = "my-org"
						search_query = "is:pr"
						auth {
							authenticator = "my_auth"
						}
					webhook {
						endpoint = "0.0.0.0:8080"
						path = "/webhooks"
						health_path = "/healthz"
						secret = "my-webhook-secret"
						required_headers = {
							"X-Custom-Header" = "value",
						}
				output {}
			`,
			validate: func(t *testing.T, cfg *githubreceiver.Config) {
				require.NotNil(t, cfg)
				require.NotNil(t, cfg.WebHook)
			},
		},
		{
			testName: "webhook with defaults",
			cfg: `
					collection_interval = "30s"
					scraper {
						github_org = "my-org"
						search_query = "is:pr"
						auth {
							authenticator = "my_auth"
						}
					webhook {
						endpoint = "0.0.0.0:9090"
				output {}
			`,
			validate: func(t *testing.T, cfg *githubreceiver.Config) {
				require.NotNil(t, cfg)
				require.NotNil(t, cfg.WebHook)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var args github.Arguments
			err := syntax.Unmarshal([]byte(tc.cfg), &args)
			require.NoError(t, err)

			actualPtr, err := args.Convert()
			require.NoError(t, err)

			actual := actualPtr.(*githubreceiver.Config)
			tc.validate(t, actual)
		})
	}
}

func TestArguments_Validate(t *testing.T) {
	tests := []struct {
		testName      string
		cfg           string
		expectedError string
	}{
		{
			testName: "missing github_org",
			cfg: `
					collection_interval = "30s"
					scraper {
						search_query = "is:pr"
						auth {
							authenticator = "my_auth"
						}
				output {}
			`,
			expectedError: "github_org",
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var args github.Arguments
			require.ErrorContains(t, syntax.Unmarshal([]byte(tc.cfg), &args), tc.expectedError)
		})
	}
}

func TestArguments_SetToDefault(t *testing.T) {
	tests := []struct {
		testName string
		cfg      string
		validate func(t *testing.T, args *github.Arguments)
	}{
		{
			testName: "default collection_interval",
			cfg: `
					scraper {
						github_org = "my-org"
						search_query = "is:pr"
						auth {
							authenticator = "my_auth"
						}
				output {}
			`,
			validate: func(t *testing.T, args *github.Arguments) {
				cfg, err := args.Convert()
				require.NoError(t, err)
				githubCfg := cfg.(*githubreceiver.Config)
				require.Equal(t, 30*time.Second, githubCfg.ControllerConfig.CollectionInterval)
			},
		},
		{
			testName: "default webhook values",
			cfg: `
					collection_interval = "30s"
					scraper {
						github_org = "my-org"
						search_query = "is:pr"
						auth {
							authenticator = "my_auth"
						}
					webhook {
				output {}
			`,
			validate: func(t *testing.T, args *github.Arguments) {
				// Webhook should have default values set
				require.NotNil(t, args)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var args github.Arguments
			err := syntax.Unmarshal([]byte(tc.cfg), &args)
			require.NoError(t, err)
			tc.validate(t, &args)
		})
	}
}

func TestDebugMetricsConfig(t *testing.T) {
	tests := []struct {
		testName string
		alloyCfg string
		expected otelcolCfg.DebugMetricsArguments
	}{
		{
			testName: "default",
			alloyCfg: `
					collection_interval = "30s"
					scraper {
						github_org = "my-org"
						search_query = "is:pr"
						auth {
							authenticator = "my_auth"
						}
				output {}
			`,
			expected: otelcolCfg.DebugMetricsArguments{
				DisableHighCardinalityMetrics: true,
				Level:                         otelcolCfg.LevelDetailed,
			},
		},
		{
			testName: "explicit_false",
			alloyCfg: `
					collection_interval = "30s"
					scraper {
						github_org = "my-org"
						search_query = "is:pr"
						auth {
							authenticator = "my_auth"
						}

				debug_metrics {
					disable_high_cardinality_metrics = false
				output {}
			`,
			expected: otelcolCfg.DebugMetricsArguments{
				DisableHighCardinalityMetrics: false,
				Level:                         otelcolCfg.LevelDetailed,
			},
		},
		{
			testName: "explicit_true",
			alloyCfg: `
					collection_interval = "30s"
					scraper {
						github_org = "my-org"
						search_query = "is:pr"
						auth {
							authenticator = "my_auth"
						}

				debug_metrics {
					disable_high_cardinality_metrics = true
				output {}
			`,
			expected: otelcolCfg.DebugMetricsArguments{
				DisableHighCardinalityMetrics: true,
				Level:                         otelcolCfg.LevelDetailed,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var args github.Arguments
			require.NoError(t, syntax.Unmarshal([]byte(tc.alloyCfg), &args))
			_, err := args.Convert()
			require.NoError(t, err)

			require.Equal(t, tc.expected, args.DebugMetricsConfig())
		})
	}
}

func TestConfig(t *testing.T) {
	alloyCfg := `
		initial_delay = "5s"
		collection_interval = "60s"
		scraper {
			github_org = "test-org"
			search_query = "is:pr is:open"
			endpoint = "https://api.github.com"
			auth {
				authenticator = "test_auth"
			}
		}
		output {}
	`
	var args github.Arguments
	err := syntax.Unmarshal([]byte(alloyCfg), &args)
	require.NoError(t, err)
	require.NoError(t, args.Validate())

	comCfg, err := args.Convert()
	require.NoError(t, err)

	githubComCfg, ok := comCfg.(*githubreceiver.Config)
	require.True(t, ok)

	require.Equal(t, 5*time.Second, githubComCfg.ControllerConfig.InitialDelay)
	require.Equal(t, 60*time.Second, githubComCfg.ControllerConfig.CollectionInterval)
	require.NotNil(t, githubComCfg.Scrapers)
}

func TestConfigDefault(t *testing.T) {
	args := github.Arguments{}
	args.SetToDefault()

	// Default config should have 30s collection interval
	fCfg, err := args.Convert()
	require.NoError(t, err)
	cfg := fCfg.(*githubreceiver.Config)
	require.Equal(t, 30*time.Second, cfg.ControllerConfig.CollectionInterval)

	// Validation should pass even with no scrapers since they are optional in the config
	require.NoError(t, args.Validate())
}

func TestMetricsConfig(t *testing.T) {
	tests := []struct {
		testName string
		cfg      string
		validate func(t *testing.T, cfg *githubreceiver.Config)
	}{
		{
			testName: "all metrics enabled explicitly",
			cfg: `
					collection_interval = "30s"
					scraper {
						github_org = "my-org"
						search_query = "is:pr"
						auth {
							authenticator = "my_auth"
						}
					metrics {
						vcs.change.count {
							enabled = true
						}
						vcs.change.duration {
							enabled = true
						}
						vcs.change.time_to_approval {
							enabled = true
						}
						vcs.change.time_to_merge {
							enabled = true
						}
						vcs.ref.count {
							enabled = true
						}
						vcs.ref.lines_delta {
							enabled = true
						}
						vcs.ref.revisions_delta {
							enabled = true
						}
						vcs.ref.time {
							enabled = true
						}
						vcs.repository.count {
							enabled = true
						}
						vcs.contributor.count {
							enabled = true
						}
				output {}
			`,
			validate: func(t *testing.T, cfg *githubreceiver.Config) {
				require.NotNil(t, cfg)
				require.NotNil(t, cfg.Scrapers)
			},
		},
		{
			testName: "selective metrics disabled",
			cfg: `
					collection_interval = "30s"
					scraper {
						github_org = "my-org"
						search_query = "is:pr"
						auth {
							authenticator = "my_auth"
						}
					metrics {
						vcs.change.count {
							enabled = true
						}
						vcs.change.duration {
							enabled = false
						}
						vcs.contributor.count {
							enabled = false
						}
				output {}
			`,
			validate: func(t *testing.T, cfg *githubreceiver.Config) {
				require.NotNil(t, cfg)
				require.NotNil(t, cfg.Scrapers)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var args github.Arguments
			err := syntax.Unmarshal([]byte(tc.cfg), &args)
			require.NoError(t, err)

			actualPtr, err := args.Convert()
			require.NoError(t, err)

			actual := actualPtr.(*githubreceiver.Config)
			tc.validate(t, actual)
		})
	}
}

func TestWebhookConfig(t *testing.T) {
	tests := []struct {
		testName string
		cfg      string
		validate func(t *testing.T, cfg *githubreceiver.Config)
	}{
		{
			testName: "webhook with custom values",
			cfg: `
				collection_interval = "30s"
				scraper {
					github_org = "my-org"
					search_query = "is:pr"
					auth {
						authenticator = "my_auth"
				webhook {
					endpoint = "0.0.0.0:8080"
					path = "/github/events"
					health_path = "/github/health"
					secret = "super-secret"
					required_headers = {
						"X-GitHub-Event" = "pull_request",
				output {}
			`,
			validate: func(t *testing.T, cfg *githubreceiver.Config) {
				require.NotNil(t, cfg)
				require.NotNil(t, cfg.WebHook)
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			var args github.Arguments
			err := syntax.Unmarshal([]byte(tc.cfg), &args)
			require.NoError(t, err)

			actualPtr, err := args.Convert()
			require.NoError(t, err)

			actual := actualPtr.(*githubreceiver.Config)
			tc.validate(t, actual)
		})
	}
}
