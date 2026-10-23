// Package mocks contains mock for RDS client
package mocks

import (
	"context"
	"crypto/rand"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_rds "github.com/aws/aws-sdk-go-v2/service/rds"
	aws_rds_types "github.com/aws/aws-sdk-go-v2/service/rds/types"
)

const (
	mockAWSRegion    = "eu-west-3"
	mockAWSAccountID = "123456789012"
)

type RDSClient struct {
	DescribeDBClustersOutput                *aws_rds.DescribeDBClustersOutput
	DescribeDBInstancesOutput               *aws_rds.DescribeDBInstancesOutput
	DescribeDBLogFilesOutput                *aws_rds.DescribeDBLogFilesOutput
	DescribeDBLogFilesOutputError           error
	DescribePendingMaintenanceActionsOutput *aws_rds.DescribePendingMaintenanceActionsOutput
	DescribeDBEngineVersionsOutput          *aws_rds.DescribeDBEngineVersionsOutput
	DescribeDBMajorEngineVersionsOutput     *aws_rds.DescribeDBMajorEngineVersionsOutput
	DescribeDBMajorEngineVersionsError      error
	DescribeDBMajorEngineVersionsCallCount  int
	DescribeDBCapacityReservationsPages     [][]aws_rds_types.DBCapacityReservation
	DescribeDBCapacityReservationsError     error
	Error                                   error
}

func NewRDSClient() *RDSClient {
	client := &RDSClient{
		DescribeDBClustersOutput: &aws_rds.DescribeDBClustersOutput{
			DBClusters: []aws_rds_types.DBCluster{},
		},
		DescribeDBInstancesOutput: &aws_rds.DescribeDBInstancesOutput{
			DBInstances: []aws_rds_types.DBInstance{},
		},
		DescribeDBLogFilesOutput: &aws_rds.DescribeDBLogFilesOutput{
			DescribeDBLogFiles: []aws_rds_types.DescribeDBLogFilesDetails{},
		},
		DescribePendingMaintenanceActionsOutput: &aws_rds.DescribePendingMaintenanceActionsOutput{
			PendingMaintenanceActions: []aws_rds_types.ResourcePendingMaintenanceActions{},
		},
		DescribeDBEngineVersionsOutput: &aws_rds.DescribeDBEngineVersionsOutput{
			DBEngineVersions: []aws_rds_types.DBEngineVersion{},
		},
		DescribeDBMajorEngineVersionsOutput: &aws_rds.DescribeDBMajorEngineVersionsOutput{
			DBMajorEngineVersions: []aws_rds_types.DBMajorEngineVersion{},
		},
	}

	return client
}

func (m *RDSClient) WithDBInstances(instances ...aws_rds_types.DBInstance) *RDSClient {
	m.DescribeDBInstancesOutput = &aws_rds.DescribeDBInstancesOutput{
		DBInstances: instances,
	}

	return m
}

func (m *RDSClient) WithDBClusters(clusters ...aws_rds_types.DBCluster) *RDSClient {
	m.DescribeDBClustersOutput = &aws_rds.DescribeDBClustersOutput{
		DBClusters: clusters,
	}

	return m
}

func (m *RDSClient) WithLogFiles(files []aws_rds_types.DescribeDBLogFilesDetails) *RDSClient {
	m.DescribeDBLogFilesOutput = &aws_rds.DescribeDBLogFilesOutput{
		DescribeDBLogFiles: files,
	}

	return m
}

func (m *RDSClient) WithLogFilesOutputError(output error) *RDSClient {
	m.DescribeDBLogFilesOutputError = output

	return m
}

func (m *RDSClient) WithDescribeDBMajorEngineVersionsOutput(output *aws_rds.DescribeDBMajorEngineVersionsOutput) *RDSClient {
	m.DescribeDBMajorEngineVersionsOutput = output

	return m
}

func (m *RDSClient) WithDescribeDBMajorEngineVersionsError(err error) *RDSClient {
	m.DescribeDBMajorEngineVersionsError = err

	return m
}

func (m *RDSClient) GetDescribeDBMajorEngineVersionsCallCount() int {
	return m.DescribeDBMajorEngineVersionsCallCount
}

// WithDBCapacityReservations returns the reservations in a single page
func (m *RDSClient) WithDBCapacityReservations(reservations ...aws_rds_types.DBCapacityReservation) *RDSClient {
	m.DescribeDBCapacityReservationsPages = [][]aws_rds_types.DBCapacityReservation{reservations}

	return m
}

// WithDBCapacityReservationsPages returns each list of reservations in a separate page
func (m *RDSClient) WithDBCapacityReservationsPages(pages ...[]aws_rds_types.DBCapacityReservation) *RDSClient {
	m.DescribeDBCapacityReservationsPages = pages

	return m
}

func (m *RDSClient) WithDBCapacityReservationsError(err error) *RDSClient {
	m.DescribeDBCapacityReservationsError = err

	return m
}

func (m RDSClient) DescribeDBClusters(ctx context.Context, params *aws_rds.DescribeDBClustersInput, optFns ...func(*aws_rds.Options)) (*aws_rds.DescribeDBClustersOutput, error) {
	return m.DescribeDBClustersOutput, nil
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

func (m RDSClient) DescribeDBEngineVersions(context.Context, *aws_rds.DescribeDBEngineVersionsInput, ...func(*aws_rds.Options)) (*aws_rds.DescribeDBEngineVersionsOutput, error) {
	return m.DescribeDBEngineVersionsOutput, nil
}

func (m *RDSClient) DescribeDBMajorEngineVersions(context.Context, *aws_rds.DescribeDBMajorEngineVersionsInput, ...func(*aws_rds.Options)) (*aws_rds.DescribeDBMajorEngineVersionsOutput, error) {
	m.DescribeDBMajorEngineVersionsCallCount++
	return m.DescribeDBMajorEngineVersionsOutput, m.DescribeDBMajorEngineVersionsError
}

// DescribeDBCapacityReservations paginates over DescribeDBCapacityReservationsPages, the marker being the next page index
func (m RDSClient) DescribeDBCapacityReservations(_ context.Context, input *aws_rds.DescribeDBCapacityReservationsInput, _ ...func(*aws_rds.Options)) (*aws_rds.DescribeDBCapacityReservationsOutput, error) {
	if m.DescribeDBCapacityReservationsError != nil {
		return nil, m.DescribeDBCapacityReservationsError
	}

	page := 0
	if input.Marker != nil {
		page, _ = strconv.Atoi(*input.Marker)
	}

	output := &aws_rds.DescribeDBCapacityReservationsOutput{}

	if page < len(m.DescribeDBCapacityReservationsPages) {
		output.DBCapacityReservations = m.DescribeDBCapacityReservationsPages[page]
	}

	if page+1 < len(m.DescribeDBCapacityReservationsPages) {
		output.Marker = aws.String(strconv.Itoa(page + 1))
	}

	return output, nil
}

//nolint:golint,mnd
func NewDBCapacityReservation(engine string, instanceClass string, reservedPerAZ int32, availableByAZ map[string]int32) aws_rds_types.DBCapacityReservation {
	awsRegion := mockAWSRegion
	awsAccountID := mockAWSAccountID
	identifier := "dbcr-" + RandomString(10)

	availability := make([]aws_rds_types.AvailableCapacityPerAvailabilityZone, 0, len(availableByAZ))
	for az, available := range availableByAZ {
		availability = append(availability, aws_rds_types.AvailableCapacityPerAvailabilityZone{
			AvailabilityZone:  aws.String(az),
			AvailableCapacity: aws.Int32(available),
		})
	}

	return aws_rds_types.DBCapacityReservation{
		AvailableCapacityPerAvailabilityZoneList: availability,
		CreateDate:                               aws.Time(time.Now()),
		DBCapacityReservationArn:                 aws.String(fmt.Sprintf("arn:aws:rds:%s:%s:db-capacity-reservation:%s", awsRegion, awsAccountID, identifier)),
		DBCapacityReservationIdentifier:          aws.String(identifier),
		DBCapacityReservationName:                aws.String("reservation-" + RandomString(4)),
		DBInstanceClass:                          aws.String(instanceClass),
		Engine:                                   aws.String(engine),
		ReservedCapacityPerAvailabilityZone:      aws.Int32(reservedPerAZ),
		State:                                    aws_rds_types.DBCapacityReservationStateActive,
	}
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

//nolint:golint,mnd
func NewRdsInstance() *aws_rds_types.DBInstance {
	awsRegion := mockAWSRegion
	awsAccountID := mockAWSAccountID
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
		AvailabilityZone:           aws.String("eu-west-3a"),
		TagList:                    []aws_rds_types.Tag{{Key: aws.String("Environment"), Value: aws.String("unittest")}, {Key: aws.String("Team"), Value: aws.String("sre")}},
	}
}

//nolint:golint,mnd
func NewRdsCluster() *aws_rds_types.DBCluster {
	awsRegion := mockAWSRegion
	awsAccountID := mockAWSAccountID
	DBClusterIdentifier := RandomString(10)
	DBClusterResourceID := RandomString(10)
	arn := fmt.Sprintf("arn:aws:rds:%s:%s:db:%s", awsRegion, awsAccountID, DBClusterIdentifier)

	now := time.Now()

	return &aws_rds_types.DBCluster{
		AllocatedStorage:           aws.Int32(5),
		BackupRetentionPeriod:      aws.Int32(7),
		DBClusterArn:               aws.String(arn),
		DBClusterInstanceClass:     aws.String("t3.large"),
		DBClusterIdentifier:        aws.String(DBClusterIdentifier),
		DbClusterResourceId:        aws.String(DBClusterResourceID),
		Status:                     aws.String("available"),
		DeletionProtection:         aws.Bool(true),
		Engine:                     aws.String("postgres"),
		EngineVersion:              aws.String("14.9"),
		Iops:                       aws.Int32(3000),
		MultiAZ:                    aws.Bool(true),
		PerformanceInsightsEnabled: aws.Bool(true),
		PubliclyAccessible:         aws.Bool(true),
		StorageType:                aws.String("gp3"),
		CertificateDetails:         newRdsCertificateDetails(),
		ClusterCreateTime:          &now,
		TagList:                    []aws_rds_types.Tag{{Key: aws.String("Environment"), Value: aws.String("unittest")}, {Key: aws.String("Team"), Value: aws.String("sre")}},
	}
}

func NewAuroraCluster() *aws_rds_types.DBCluster {
	cluster := NewRdsCluster()
	cluster.AllocatedStorage = aws.Int32(1) // AllocatedStorage always returns 1, because Aurora DB cluster storage size isn't fixed, but instead automatically adjusts as needed.
	cluster.StorageType = aws.String("gp3")

	return cluster
}

func NewAuroraServerlessCluster() *aws_rds_types.DBCluster {
	cluster := NewRdsCluster()
	cluster.AllocatedStorage = aws.Int32(1) // AllocatedStorage always returns 1, because Aurora DB cluster storage size isn't fixed, but instead automatically adjusts as needed.
	cluster.StorageType = aws.String("aurora-iopt1")
	cluster.ServerlessV2ScalingConfiguration = &aws_rds_types.ServerlessV2ScalingConfigurationInfo{
		MinCapacity: aws.Float64(0.5),
		MaxCapacity: aws.Float64(12.5),
	}

	return cluster
}

func NewMultiAZCluster() *aws_rds_types.DBCluster {
	cluster := NewRdsCluster()
	cluster.DBClusterMembers = []aws_rds_types.DBClusterMember{
		{
			DBClusterParameterGroupStatus: aws.String("in-sync"),
			DBInstanceIdentifier:          aws.String(*cluster.DBClusterIdentifier + "-instance-1"),
			IsClusterWriter:               aws.Bool(true),
			PromotionTier:                 aws.Int32(1),
		},
		{
			DBClusterParameterGroupStatus: aws.String("in-sync"),
			DBInstanceIdentifier:          aws.String(*cluster.DBClusterIdentifier + "-instance-2"),
			IsClusterWriter:               aws.Bool(false),
			PromotionTier:                 aws.Int32(1),
		},
		{
			DBClusterParameterGroupStatus: aws.String("in-sync"),
			DBInstanceIdentifier:          aws.String(*cluster.DBClusterIdentifier + "-instance-3"),
			IsClusterWriter:               aws.Bool(false),
			PromotionTier:                 aws.Int32(1),
		},
	}

	return cluster
}
