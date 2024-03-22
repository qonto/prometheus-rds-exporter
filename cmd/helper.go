package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

func getAWSConfiguration(logger *slog.Logger, roleArn string, sessionName string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return aws.Config{}, fmt.Errorf("can't create AWS session: %w", err)
	}

	if roleArn != "" {
		logger.Debug("Assume role", "role", roleArn)

		client := sts.NewFromConfig(cfg)
		creds := stscreds.NewAssumeRoleProvider(client, roleArn, func(o *stscreds.AssumeRoleOptions) {
			o.RoleSessionName = sessionName
		})
		cfg.Credentials = aws.NewCredentialsCache(creds)
	}

	// Try to automatically find the current AWS region via AWS EC2 IMDS metadata
	if cfg.Region == "" {
		logger.Debug("search AWS region using IMDS")

		client := imds.NewFromConfig(cfg)

		response, err := client.GetRegion(context.TODO(), &imds.GetRegionInput{})
		if err == nil {
			cfg.Region = response.Region
			logger.Info("found AWS region via IMDS", "region", cfg.Region)
		}
	}

	return cfg, nil
}

func getAWSSessionInformation(cfg aws.Config) (string, string, error) {
	client := sts.NewFromConfig(cfg)

	output, err := client.GetCallerIdentity(context.TODO(), nil)
	if err != nil {
		return "", "", fmt.Errorf("can't fetch information about current session: %w", err)
	}

	return *output.Account, cfg.Region, nil
}

func getAWSConfigurationByCredentials(logger *slog.Logger, configuration exporterConfig) ([]aws.Config, error) {
	var configs []aws.Config
	accountsFromYaml := configuration.AwsCredentials
	if reflect.ValueOf(accountsFromYaml).IsZero() {
		logger.Error("AWS accounts not configured in yaml")
		return nil, nil
	} else {
		accounts := accountsFromYaml.Accounts
		for _, c := range accounts {
			aws_access_key_id := c.AwsAccessKeyID
			aws_secret_access_key := c.AwsSecretAccessKey
			staticProvider := credentials.NewStaticCredentialsProvider(
				aws_access_key_id,
				aws_secret_access_key,
				"",
			)
			cfg, err := config.LoadDefaultConfig(
				context.Background(),
				config.WithCredentialsProvider(staticProvider),
			)
			if err != nil {
				return nil, err
			}
			for _, region := range c.Regions {
				cfg.Region = region
				configs = append(configs, cfg)
			}
		}
	}
	return configs, nil
}
