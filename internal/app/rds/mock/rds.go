// Package mocks contains mock for RDS client
package mocks

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_rds "github.com/aws/aws-sdk-go-v2/service/rds"
	aws_rds_types "github.com/aws/aws-sdk-go-v2/service/rds/types"
)

type RDSClient struct {
	DescribeDBInstancesOutput               *aws_rds.DescribeDBInstancesOutput
	DescribePendingMaintenanceActionsOutput *aws_rds.DescribePendingMaintenanceActionsOutput
	DescribeDBLogFilesOutput                *aws_rds.DescribeDBLogFilesOutput
	DescribeDBLogFilesOutputError           error
	Error                                   error
}

func (m RDSClient) DescribeDBInstancesPages(input *aws_rds.DescribeDBInstancesInput, fn func(*aws_rds.DescribeDBInstancesOutput, bool) bool) error {
	fn(m.DescribeDBInstancesOutput, false)

	return nil
}

func (m RDSClient) DescribePendingMaintenanceActions(context.Context, *aws_rds.DescribePendingMaintenanceActionsInput, ...func(*aws_rds.Options)) (*aws_rds.DescribePendingMaintenanceActionsOutput, error) {
	return m.DescribePendingMaintenanceActionsOutput, m.Error
}

func (m RDSClient) DescribeDBLogFiles(ctx context.Context, input *aws_rds.DescribeDBLogFilesInput, fn ...func(*aws_rds.Options)) (*aws_rds.DescribeDBLogFilesOutput, error) {
	return m.DescribeDBLogFilesOutput, m.DescribeDBLogFilesOutputError
}

func (m RDSClient) DescribeDBInstances(context.Context, *aws_rds.DescribeDBInstancesInput, ...func(*aws_rds.Options)) (*aws_rds.DescribeDBInstancesOutput, error) {
	return m.DescribeDBInstancesOutput, nil
}

// RandomString returns a random alphanumeric string of the specified length
func RandomString(length int) string {
	buf := make([]byte, length)

	_, err := rand.Read(buf)
	if err != nil {
		panic(err) // out of randomness, should never happen
	}

	return fmt.Sprintf("%x", buf)
}

func newRdsCertificateDetails() *aws_rds_types.CertificateDetails {
	return &aws_rds_types.CertificateDetails{
		CAIdentifier: aws.String("rds-ca-2019"),
		ValidTill: aws.Time(time.Date(
			2024, time.August, 22,
			17, 8, 50, 0, time.UTC,
		)),
	}
}

//nolint:golint,gomnd
func NewRdsInstance() *aws_rds_types.DBInstance {
	awsRegion := "eu-west-3"
	awsAccountID := "123456789012"
	DBInstanceIdentifier := RandomString(10)
	arn := fmt.Sprintf("arn:aws:rds:%s:%s:db:%s", awsRegion, awsAccountID, DBInstanceIdentifier)

	now := time.Now()

	return &aws_rds_types.DBInstance{
		AllocatedStorage:           aws.Int32(5),
		BackupRetentionPeriod:      aws.Int32(7),
		DBInstanceArn:              aws.String(arn),
		DBInstanceClass:            aws.String("t3.large"),
		DBInstanceIdentifier:       aws.String(DBInstanceIdentifier),
		DBInstanceStatus:           aws.String("available"),
		DbiResourceId:              aws.String("resource1"),
		DeletionProtection:         aws.Bool(true),
		Engine:                     aws.String("postgres"),
		EngineVersion:              aws.String("14.9"),
		Iops:                       aws.Int32(3000),
		MaxAllocatedStorage:        aws.Int32(10),
		MultiAZ:                    aws.Bool(true),
		PerformanceInsightsEnabled: aws.Bool(true),
		PubliclyAccessible:         aws.Bool(true),
		StorageType:                aws.String("gp3"),
		CACertificateIdentifier:    aws.String("rds-ca-2019"),
		CertificateDetails:         newRdsCertificateDetails(),
		InstanceCreateTime:         &now,
		TagList:                    []aws_rds_types.Tag{{Key: aws.String("Environment"), Value: aws.String("unittest")}, {Key: aws.String("Team"), Value: aws.String("sre")}},
	}
}
