// Package cmd implements command to start the RDS exporter
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/servicequotas"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/qonto/prometheus-rds-exporter/internal/app/exporter"
	"github.com/qonto/prometheus-rds-exporter/internal/infra/build"
	"github.com/qonto/prometheus-rds-exporter/internal/infra/http"
	"github.com/qonto/prometheus-rds-exporter/internal/infra/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	configErrorExitCode   = 1
	httpErrorExitCode     = 2
	exporterErrorExitCode = 3
	awsErrorExitCode      = 4
)

var cfgFile string
var actRegionClient []exporter.AccountRegionClients

type exporterConfig struct {
	Debug                  bool   `mapstructure:"debug"`
	LogFormat              string `mapstructure:"log-format"`
	TLSCertPath            string `mapstructure:"tls-cert-path"`
	TLSKeyPath             string `mapstructure:"tls-key-path"`
	MetricPath             string `mapstructure:"metrics-path"`
	ListenAddress          string `mapstructure:"listen-address"`
	AWSAssumeRoleSession   string `mapstructure:"aws-assume-role-session"`
	AWSAssumeRoleArn       string `mapstructure:"aws-assume-role-arn"`
	CollectInstanceMetrics bool   `mapstructure:"collect-instance-metrics"`
	CollectInstanceTags    bool   `mapstructure:"collect-instance-tags"`
	CollectInstanceTypes   bool   `mapstructure:"collect-instance-types"`
	CollectLogsSize        bool   `mapstructure:"collect-logs-size"`
	CollectMaintenances    bool   `mapstructure:"collect-maintenances"`
	CollectQuotas          bool   `mapstructure:"collect-quotas"`
	CollectUsages          bool   `mapstructure:"collect-usages"`
	OTELTracesEnabled      bool   `mapstructure:"enable-otel-traces"`
	AwsCredentials         AWSCredentials
}

func run(configuration exporterConfig) {
	logger, err := logger.New(configuration.Debug, configuration.LogFormat)
	if err != nil {
		fmt.Println("ERROR: Fail to initialize logger: %w", err)
		panic(err)
	}
	collectorConfiguration := exporter.Configuration{
		CollectInstanceMetrics: configuration.CollectInstanceMetrics,
		CollectInstanceTypes:   configuration.CollectInstanceTypes,
		CollectInstanceTags:    configuration.CollectInstanceTags,
		CollectLogsSize:        configuration.CollectLogsSize,
		CollectMaintenances:    configuration.CollectMaintenances,
		CollectQuotas:          configuration.CollectQuotas,
		CollectUsages:          configuration.CollectUsages,
	}

	cfgs, err := getAWSConfigurationByCredentials(logger, configuration)
	if err != nil {
		logger.Error("can't initialize AWS configuration", "reason", err)
		os.Exit(awsErrorExitCode)
	}
	if cfgs == nil {
		logger.Info("Didn't configure aws IAM User credentials in configuration file, will use default aws configuration")
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
		ec2Client := ec2.NewFromConfig(cfg)
		cloudWatchClient := cloudwatch.NewFromConfig(cfg)
		servicequotasClient := servicequotas.NewFromConfig(cfg)

		collector := exporter.NewCollector(*logger, collectorConfiguration, awsAccountID, awsRegion, rdsClient, ec2Client, cloudWatchClient, servicequotasClient)

		prometheus.MustRegister(collector)

	} else {
		for _, cfg := range cfgs {
			awsAccountID, awsRegion, err := getAWSSessionInformation(cfg)
			if err != nil {
				logger.Error("can't identify AWS account and/or region", "reason", err)
				os.Exit(awsErrorExitCode)
			}

			rdsClient := rds.NewFromConfig(cfg)
			ec2Client := ec2.NewFromConfig(cfg)
			cloudWatchClient := cloudwatch.NewFromConfig(cfg)
			servicequotasClient := servicequotas.NewFromConfig(cfg)

			var accountRegionClients exporter.AccountRegionClients
			accountRegionClients.AwsAccountID = awsAccountID
			accountRegionClients.AwsRegion = awsRegion
			accountRegionClients.RdsClient = rdsClient
			accountRegionClients.Ec2Client = ec2Client
			accountRegionClients.CloudWatchClient = cloudWatchClient
			accountRegionClients.ServicequotasClient = servicequotasClient
			actRegionClient = append(actRegionClient, accountRegionClients)
		}
		collector := exporter.NewMultiCollector(*logger, collectorConfiguration, actRegionClient)
		prometheus.MustRegister(collector)
	}

	// http configurations for exporter service
	serverConfiguration := http.Config{
		ListenAddress: configuration.ListenAddress,
		MetricPath:    configuration.MetricPath,
		TLSCertPath:   configuration.TLSCertPath,
		TLSKeyPath:    configuration.TLSKeyPath,
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
			var c exporterConfig
			err := viper.Unmarshal(&c)
			if err != nil {
				fmt.Println("ERROR: Unable to decode configuration, %w", err)

				return
			}
			viper.UnmarshalKey("accounts", &c.AwsCredentials.Accounts)
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

	err := viper.BindPFlag("debug", cmd.Flags().Lookup("debug"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'debug' parameter: %w", err)
	}

	err = viper.BindPFlag("log-format", cmd.Flags().Lookup("log-format"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'log-format' parameter: %w", err)
	}

	err = viper.BindPFlag("enable-otel-traces", cmd.Flags().Lookup("enable-otel-traces"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'enable-otel-traces' parameter: %w", err)
	}

	err = viper.BindPFlag("metrics-path", cmd.Flags().Lookup("metrics-path"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'metrics-path' parameter: %w", err)
	}

	err = viper.BindPFlag("tls-cert-path", cmd.Flags().Lookup("tls-cert-path"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'tls-cert-path' parameter: %w", err)
	}

	err = viper.BindPFlag("tls-key-path", cmd.Flags().Lookup("tls-key-path"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'tls-key-path' parameter: %w", err)
	}

	err = viper.BindPFlag("listen-address", cmd.Flags().Lookup("listen-address"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'listen-address' parameter: %w", err)
	}

	err = viper.BindPFlag("aws-assume-role-arn", cmd.Flags().Lookup("aws-assume-role-arn"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'aws-assume-role-arn' parameter: %w", err)
	}

	err = viper.BindPFlag("aws-assume-role-session", cmd.Flags().Lookup("aws-assume-role-session"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'aws-assume-role-session' parameter: %w", err)
	}

	err = viper.BindPFlag("collect-instance-metrics", cmd.Flags().Lookup("collect-instance-metrics"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'collect-instance-metrics' parameter: %w", err)
	}

	err = viper.BindPFlag("collect-instance-tags", cmd.Flags().Lookup("collect-instance-tags"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'collect-instance-tags' parameter: %w", err)
	}

	err = viper.BindPFlag("collect-instance-types", cmd.Flags().Lookup("collect-instance-types"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'collect-instance-types' parameter: %w", err)
	}

	err = viper.BindPFlag("collect-quotas", cmd.Flags().Lookup("collect-quotas"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'collect-quotas' parameter: %w", err)
	}

	err = viper.BindPFlag("collect-usages", cmd.Flags().Lookup("collect-usages"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'collect-usages' parameter: %w", err)
	}

	err = viper.BindPFlag("collect-logs-size", cmd.Flags().Lookup("collect-logs-size"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'collect-logs-size' parameter: %w", err)
	}

	err = viper.BindPFlag("collect-maintenances", cmd.Flags().Lookup("collect-maintenances"))
	if err != nil {
		return cmd, fmt.Errorf("failed to bind 'collect-maintenances' parameter: %w", err)
	}

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
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory or current directory with name "prometheus-rds-exporter.yaml"

		configurationFilename := "prometheus-rds-exporter.yaml"
		currentPathFilename := configurationFilename
		homeFilename := filepath.Join(home, configurationFilename)

		if _, err := os.Stat(homeFilename); err == nil {
			viper.SetConfigFile(homeFilename)
		}

		if _, err := os.Stat(currentPathFilename); err == nil {
			viper.SetConfigFile(currentPathFilename)
		}
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	viper.SetEnvPrefix("prometheus_rds_exporter") // will be uppercased automatically
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}
