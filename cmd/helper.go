package cmd

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
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

	return cfg, nil
}

func getAWSSessionInformation(cfg aws.Config) (string, string, error) {
	client := sts.NewFromConfig(cfg)

	output, err := client.GetCallerIdentity(context.TODO(), nil)
	if err != nil {
		return "", "", fmt.Errorf("can't fetch information about current session: %w", err)
	}

	return cfg.Region, *output.Account, nil
}
