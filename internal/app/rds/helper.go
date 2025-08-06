package rds

import (
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_rds_types "github.com/aws/aws-sdk-go-v2/service/rds/types"
	converter "github.com/qonto/prometheus-rds-exporter/internal/app/unit"
)

func ThresholdValue(lowerBoundary int64, value int64, higherBoundary int64) int64 {
	switch {
	case value < lowerBoundary:
		return lowerBoundary
	case value > higherBoundary:
		return higherBoundary
	default:
		return value
	}
}

// GetDBIdentifierFromARN returns instance identifier from its ARN
func GetDBIdentifierFromARN(arn string) string {
	arnChunk := strings.Split(arn, ":")

	return arnChunk[len(arnChunk)-1]
}

// GetDBInstanceStatusCode returns instance status numeric code
func GetDBInstanceStatusCode(status string) int {
	var instanceStatus int

	instanceStatus, isFound := instanceStatuses[status]
	if !isFound {
		return InstanceStatusUnknown
	}

	return instanceStatus
}

// getStorageMetrics returns storage metrics following AWS rules
func getStorageMetrics(storageType string, allocatedStorage int64, rawIops int64, rawStorageThroughput int64) (int64, int64) {
	// IOPS and throughput depends of the RDS storage class type and the allocated storage
	// https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Storage.html
	var iops, storageThroughput int64

	switch storageType {
	case "gp2":
		/*
			Baseline IOPS performance scales linearly between a minimum of 100 and a maximum of 16,000 at a rate of 3 IOPS per GiB of volume size. IOPS performance is provisioned as follows:
			- Volumes 33.33 GiB and smaller are provisioned with the minimum of 100 IOPS.
			- Volumes larger than 33.33 GiB are provisioned with 3 IOPS per GiB of volume size up to the maximum of 16,000 IOPS, which is reached at 5,334 GiB (3 X 5,334).
			- Volumes 5,334 GiB and larger are provisioned with 16,000 IOPS.
			https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/general-purpose.html#EBSVolumeTypes_gp2
		*/
		iops = ThresholdValue(gp2IOPSMin, allocatedStorage*gp2IOPSPerGB, gp2IOPSMax)

		/*
			Throughput performance is provisioned as follows:
			- Volumes that are 170 GiB and smaller deliver a maximum throughput of 128 MiB/s.
			- Volumes larger than 170 GiB but smaller than 334 GiB can burst to a maximum throughput of 250 MiB/s.
			- Volumes that are 334 GiB and larger deliver 250 MiB/s.
			https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/general-purpose.html#EBSVolumeTypes_gp2
		*/
		if allocatedStorage >= gp2StorageThroughputVolumeThreshold {
			storageThroughput = gp2StorageThroughputLargeVolume
		} else {
			storageThroughput = gp2StorageThroughputSmallVolume
		}
	case "gp3":
		// iops and storageThroughput are returned by AWS RDS API for GP3 class type
		iops = rawIops
		storageThroughput = rawStorageThroughput
	case "io1":
		iops = rawIops

		/*
			Volumes provisioned with more than 32,000 IOPS (up to the maximum of 64,000 IOPS) yield a linear increase in throughput at a rate of 16 KiB per provisioned IOPS.
			https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/provisioned-iops.html#EBSVolumeTypes_piops
		*/
		switch {
		case iops >= io1HighIOPSThroughputThreshold:
			storageThroughput = io1HighIOPSThroughputValue
		case iops >= io1LargeIOPSThroughputThreshold:
			storageThroughput = converter.KiloByteToMegaBytes(iops * io1LargeIOPSThroughputValue)
		case iops >= io1MediumIOPSThroughputThreshold:
			storageThroughput = io1MediumIOPSThroughputValue
		default:
			storageThroughput = converter.KiloByteToMegaBytes(iops * io1DefaultIOPSThroughputValue)
		}
	case "io2":
		iops = rawIops

		/*
			Throughput scales proportionally up to 0.256 MiB/s per provisioned IOPS.
			Maximum throughput of 4,000 MiB/s can be achieved at 256,000 IOPS with a 16-KiB I/O size and 16,000 IOPS or higher with a 256-KiB I/O size.
			For DB instances not based on the AWS Nitro System, maximum throughput of 2,000 MiB/s can be achieved at 128,000 IOPS with a 16-KiB I/O size.
			https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Storage.html#USER_PIOPS.io2
			https://docs.aws.amazon.com/ebs/latest/userguide/provisioned-iops.html#io2-block-express
		*/
		theoreticalThroughput := int64(float64(iops) * io2StorageThroughputPerIOPS)
		storageThroughput = ThresholdValue(io2StorageMinThroughput, theoreticalThroughput, io2StorageMaxThroughput)
	default:
		iops = rawIops
		storageThroughput = rawStorageThroughput
	}

	return iops, storageThroughput
}

// GetInstanceRole determines the role of the specified DB instance within its cluster context.
// It returns the role (e.g., primary, reader, replica) and the source instance identifier if applicable.
func GetInstanceRole(instance *aws_rds_types.DBInstance, cluster ClusterMetrics) (DBRole, string) {
	if instance == nil {
		return "", ""
	}

	// Handle DB cluster
	if instance.DBClusterIdentifier != nil && cluster.DBClusterIdentifier == *instance.DBClusterIdentifier {
		if role, exists := cluster.Members[*instance.DBInstanceIdentifier]; exists {
			if role == RoleWriter {
				return RoleWriter, ""
			}
			return RoleReader, cluster.WriterDBInstanceIdentifier
		}
	}

	// Handle traditional replica
	if instance.ReadReplicaSourceDBInstanceIdentifier != nil {
		return RoleReplica, *instance.ReadReplicaSourceDBInstanceIdentifier
	}

	// Fallback role
	return RolePrimary, ""
}

// ConvertRDSTagsToMap safely converts a slice of RDS tags into a map[string]string.
func ConvertRDSTagsToMap(tags []aws_rds_types.Tag) map[string]string {
	tagMap := make(map[string]string, len(tags))

	for _, tag := range tags {
		key := aws.ToString(tag.Key)
		value := aws.ToString(tag.Value)
		if key != "" {
			tagMap[key] = value
		}
	}

	return tagMap
}
