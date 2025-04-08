// Package cmd implements command to start the RDS exporter
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/pi"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qonto/prometheus-rds-exporter/internal/app/exporter"
	"github.com/qonto/prometheus-rds-exporter/internal/infra/build"
	"github.com/qonto/prometheus-rds-exporter/internal/infra/http"
	"github.com/qonto/prometheus-rds-exporter/internal/infra/logger"
	"github.com/spf13/cobra"
)

const (
	configErrorExitCode   = 1
	httpErrorExitCode     = 2
	exporterErrorExitCode = 3
	awsErrorExitCode      = 4
)

var (
	cfgFile string
	k       = koanf.New(".")
)

type exporterConfig struct {
	Debug                      bool                `koanf:"debug"`
	LogFormat                  string              `koanf:"log-format"`
	TLSCertPath                string              `koanf:"tls-cert-path"`
	TLSKeyPath                 string              `koanf:"tls-key-path"`
	MetricPath                 string              `koanf:"metrics-path"`
	ListenAddress              string              `koanf:"listen-address"`
	AWSAssumeRoleSession       string              `koanf:"aws-assume-role-session"`
	AWSAssumeRoleArn           string              `koanf:"aws-assume-role-arn"`
	CollectInstanceMetrics     bool                `koanf:"collect-instance-metrics"`
	CollectInstanceTags        bool                `koanf:"collect-instance-tags"`
	CollectInstanceTypes       bool                `koanf:"collect-instance-types"`
	CollectLogsSize            bool                `koanf:"collect-logs-size"`
	CollectMaintenances        bool                `koanf:"collect-maintenances"`
	CollectQuotas              bool                `koanf:"collect-quotas"`
	CollectUsages              bool                `koanf:"collect-usages"`
	CollectPerformanceInsights bool                `koanf:"collect-performance-insights"`
	OTELTracesEnabled          bool                `koanf:"enable-otel-traces"`
	TagSelections              map[string][]string `koanf:"tag-selections"`
}

func run(configuration exporterConfig) {
	logger, err := logger.New(configuration.Debug, configuration.LogFormat)
	if err != nil {
		fmt.Println("ERROR: Fail to initialize logger: %w", err)
		panic(err)
	}

	logger.Debug(fmt.Sprintf("Config: %+v\n", configuration))

	cfg, err := getAWSConfiguration(logger, configuration.AWSAssumeRoleArn, configuration.AWSAssumeRoleSession)
	if err != nil {
		logger.Error("can't initialize AWS configuration", "reason", err)
		os.Exit(awsErrorExitCode)
	}

	awsAccountID, awsRegion, err := getAWSSessionInformation(cfg)
	if err != nil {
		logger.Error("can't identify AWS account and/or region", "reason", err)
		os.Exit(awsErrorExitCode)
	}

	rdsClient := rds.NewFromConfig(cfg)

	var tagClient *resourcegroupstaggingapi.Client

	if configuration.TagSelections != nil {
		tagClient = resourcegroupstaggingapi.NewFromConfig(cfg)
	}

	ec2Client := ec2.NewFromConfig(cfg)
	cloudWatchClient := cloudwatch.NewFromConfig(cfg)
	servicequotasClient := servicequotas.NewFromConfig(cfg)
	piClient := pi.NewFromConfig(cfg)

	collectorConfiguration := exporter.Configuration{
		CollectInstanceMetrics:     configuration.CollectInstanceMetrics,
		CollectInstanceTypes:       configuration.CollectInstanceTypes,
		CollectInstanceTags:        configuration.CollectInstanceTags,
		CollectLogsSize:            configuration.CollectLogsSize,
		CollectMaintenances:        configuration.CollectMaintenances,
		CollectQuotas:              configuration.CollectQuotas,
		CollectUsages:              configuration.CollectUsages,
		CollectPerformanceInsights: configuration.CollectPerformanceInsights,
		TagSelections:              configuration.TagSelections,
	}

	collector := exporter.NewCollector(*logger, collectorConfiguration, awsAccountID, awsRegion, rdsClient, ec2Client, cloudWatchClient, piClient, servicequotasClient, tagClient)

	prometheus.MustRegister(collector)

	serverConfiguration := http.Config{
		ListenAddress:     configuration.ListenAddress,
		MetricPath:        configuration.MetricPath,
		TLSCertPath:       configuration.TLSCertPath,
		TLSKeyPath:        configuration.TLSKeyPath,
		OTELTracesEnabled: configuration.OTELTracesEnabled,
	}

	server := http.New(*logger, serverConfiguration)

	err = server.Start()
	if err != nil {
		logger.Error("web server error", "reason", err)
		os.Exit(httpErrorExitCode)
	}
}

func NewRootCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:     "rds-exporter",
		Version: fmt.Sprintf("%s, commit %s, built at %s", build.Version, build.CommitSHA, build.Date),
		Short:   "Prometheus exporter for AWS RDS",
		Long: `Collect AWS RDS key metrics from AWS APIs
	and expose them as Prometheus metrics.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := k.Load(posflag.Provider(cmd.Flags(), ".", k), nil)
			if err != nil {
				fmt.Printf("ERROR: Unable to interpret flags, %v\n", err)

				return
			}

			var c exporterConfig
			if err := k.Unmarshal("", &c); err != nil {
				fmt.Printf("ERROR: Unable to decode configuration, %v\n", err)

				return
			}
			run(c)
		},
	}

	cobra.OnInitialize(initConfig)

	cmd.Flags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/prometheus-rds-exporter.yaml)")
	cmd.Flags().BoolP("debug", "d", false, "Enable debug mode")
	cmd.Flags().BoolP("enable-otel-traces", "", false, "Enable OpenTelemetry traces")
	cmd.Flags().StringP("log-format", "l", "json", "Log format (text or json)")
	cmd.Flags().StringP("metrics-path", "", "/metrics", "Path under which to expose metrics")
	cmd.Flags().StringP("tls-cert-path", "", "", "Path to TLS certificate")
	cmd.Flags().StringP("tls-key-path", "", "", "Path to private key for TLS")
	cmd.Flags().StringP("listen-address", "", ":9043", "Address to listen on for web interface")
	cmd.Flags().StringP("aws-assume-role-arn", "", "", "AWS IAM ARN role to assume to fetch metrics")
	cmd.Flags().StringP("aws-assume-role-session", "", "prometheus-rds-exporter", "AWS assume role session name")
	cmd.Flags().BoolP("collect-instance-tags", "", true, "Collect AWS RDS tags")
	cmd.Flags().BoolP("collect-instance-types", "", true, "Collect AWS instance types")
	cmd.Flags().BoolP("collect-instance-metrics", "", true, "Collect AWS instance metrics")
	cmd.Flags().BoolP("collect-logs-size", "", true, "Collect AWS instances logs size")
	cmd.Flags().BoolP("collect-maintenances", "", true, "Collect AWS instances maintenances")
	cmd.Flags().BoolP("collect-quotas", "", true, "Collect AWS RDS quotas")
	cmd.Flags().BoolP("collect-usages", "", true, "Collect AWS RDS usages")
	cmd.Flags().BoolP("collect-performance-insights", "", false, "Collect AWS DB Performance Insights usages")

	return cmd, nil
}

func Execute() {
	cmd, err := NewRootCommand()
	if err != nil {
		fmt.Println("ERROR: Failed to load configuration: %w", err)
		os.Exit(configErrorExitCode)
	}

	err = cmd.Execute()
	if err != nil {
		fmt.Println("ERROR: Failed to execute exporter: %w", err)
		os.Exit(exporterErrorExitCode)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		err := k.Load(file.Provider(cfgFile), yaml.Parser())
		cobra.CheckErr(err)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory or current directory with name "prometheus-rds-exporter.yaml".
		configurationFilename := "prometheus-rds-exporter.yaml"
		currentPathFilename := configurationFilename
		homeFilename := filepath.Join(home, configurationFilename)

		if err := k.Load(file.Provider(homeFilename), yaml.Parser()); err == nil {
			fmt.Printf("Using config file: %s\n", homeFilename)
		}

		if err := k.Load(file.Provider(currentPathFilename), yaml.Parser()); err == nil {
			fmt.Printf("Using config file: %s\n", currentPathFilename)
		}
	}

	// Set environment variables.
	err := k.Load(env.Provider("PROMETHEUS_RDS_EXPORTER_", ".", func(s string) string {
		return strings.ReplaceAll(strings.ToLower(strings.TrimPrefix(s, "PROMETHEUS_RDS_EXPORTER_")), "_", ".")
	}), nil)
	cobra.CheckErr(err)
}
