package rds

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_rds "github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/patrickmn/go-cache"
)

// EngineVersionLifecycle represents the lifecycle support information for a specific engine version
type EngineVersionLifecycle struct {
	Engine                 string     // e.g., "postgres"
	MajorEngineVersion     string     // e.g., "13"
	StandardSupportEndDate *time.Time // End of standard support
	ExtendedSupportEndDate *time.Time // End of extended support
}

// EngineSupportMetrics contains the calculated remaining days for support levels
type EngineSupportMetrics struct {
	StandardSupportRemainingDays *int64
	ExtendedSupportRemainingDays *int64
}

// EngineSupportService handles retrieval and caching of engine version lifecycle data
type EngineSupportService struct {
	client RDSClient
	cache  *cache.Cache
	logger *slog.Logger
}

// NewEngineSupportService creates a new EngineSupportService with initialized cache
func NewEngineSupportService(client RDSClient, logger *slog.Logger) *EngineSupportService {
	// Initialize cache with 24-hour TTL and 1-hour cleanup interval
	c := cache.New(24*time.Hour, 1*time.Hour)

	return &EngineSupportService{
		client: client,
		cache:  c,
		logger: logger,
	}
}

// NewEngineSupportServiceWithCache creates a new EngineSupportService with a custom cache
func NewEngineSupportServiceWithCache(client RDSClient, logger slog.Logger, cache *cache.Cache) *EngineSupportService {
	return &EngineSupportService{
		client: client,
		cache:  cache,
		logger: &logger,
	}
}

// extractMajorVersion extracts the major version from a full version string
// e.g., "13.7" -> "13", "14.9" -> "14"
func (s *EngineSupportService) extractMajorVersion(fullVersion string) string {
	parts := strings.Split(fullVersion, ".")
	if len(parts) > 0 {
		return parts[0]
	}
	return fullVersion
}

// GetEngineSupportMetrics retrieves engine support metrics for a given engine and version
func (s *EngineSupportService) GetEngineSupportMetrics(ctx context.Context, engine, version string) (EngineSupportMetrics, error) {
	// Validate input parameters
	if engine == "" {
		s.logger.Error("Empty engine parameter provided")
		return EngineSupportMetrics{}, fmt.Errorf("engine parameter cannot be empty")
	}
	if version == "" {
		s.logger.Error("Empty version parameter provided", "engine", engine)
		return EngineSupportMetrics{}, fmt.Errorf("version parameter cannot be empty")
	}

	// Fetch engine version lifecycle data (uses cache if available)
	lifecycles, err := s.FetchEngineVersionLifecycles(ctx, engine)
	if err != nil {
		s.logger.Error("Failed to fetch engine lifecycle data",
			"engine", engine,
			"version", version,
			"error", err)
		return EngineSupportMetrics{}, fmt.Errorf("failed to fetch engine lifecycle data: %w", err)
	}

	// Extract major version from full version (e.g., "13.7" -> "13")
	majorVersion := s.extractMajorVersion(version)

	// Find the lifecycle data for the specific major version
	var targetLifecycle *EngineVersionLifecycle
	for _, lifecycle := range lifecycles {
		if lifecycle.MajorEngineVersion == majorVersion {
			targetLifecycle = &lifecycle
			break
		}
	}

	// If no lifecycle data found for this version, return empty metrics
	if targetLifecycle == nil {
		s.logger.Debug("No lifecycle data found for engine version", "engine", engine, "full_version", version, "major_version", majorVersion)
		return EngineSupportMetrics{}, nil
	}

	// Log if LifecycleSupportStartDate is missing (graceful handling)
	if targetLifecycle.StandardSupportEndDate == nil && targetLifecycle.ExtendedSupportEndDate == nil {
		s.logger.Debug("No support end dates available for engine version",
			"engine", engine,
			"version", version)
		return EngineSupportMetrics{}, nil
	}

	// Calculate remaining days for both support levels
	metrics := EngineSupportMetrics{
		StandardSupportRemainingDays: s.CalculateRemainingDays(targetLifecycle.StandardSupportEndDate),
		ExtendedSupportRemainingDays: s.CalculateRemainingDays(targetLifecycle.ExtendedSupportEndDate),
	}

	s.logger.Debug("Calculated engine support metrics",
		"engine", engine,
		"version", version,
		"standard_days", metrics.StandardSupportRemainingDays,
		"extended_days", metrics.ExtendedSupportRemainingDays)

	return metrics, nil
}

// FetchEngineVersionLifecycles retrieves engine version lifecycle data from AWS API
func (s *EngineSupportService) FetchEngineVersionLifecycles(ctx context.Context, engine string) ([]EngineVersionLifecycle, error) {
	// Validate input
	if engine == "" {
		s.logger.Error("Empty engine parameter provided to FetchEngineVersionLifecycles")
		return nil, fmt.Errorf("engine parameter cannot be empty")
	}

	// Generate cache key for this engine
	cacheKey := s.GenerateCacheKey(engine)

	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if lifecycles, ok := cached.([]EngineVersionLifecycle); ok {
			s.logger.Debug("Retrieved engine lifecycle data from cache", "engine", engine, "count", len(lifecycles))
			return lifecycles, nil
		} else {
			// Cache corruption - log and continue with API call
			s.logger.Error("Invalid cached data type for engine lifecycle",
				"engine", engine,
				"expected", "[]EngineVersionLifecycle",
				"actual", fmt.Sprintf("%T", cached))
			// Remove corrupted cache entry
			s.cache.Delete(cacheKey)
		}
	}

	// Cache miss - fetch from AWS API
	s.logger.Debug("Cache miss, fetching engine lifecycle data from AWS API", "engine", engine)

	input := &aws_rds.DescribeDBMajorEngineVersionsInput{
		Engine: aws.String(engine),
	}

	output, err := s.client.DescribeDBMajorEngineVersions(ctx, input)
	if err != nil {
		s.logger.Error("Failed to describe DB major engine versions", "engine", engine, "error", err)

		// Check for specific AWS error types for better error handling
		if strings.Contains(err.Error(), "AccessDenied") {
			s.logger.Error("Access denied when calling DescribeDBMajorEngineVersions - check IAM permissions",
				"engine", engine,
				"required_permission", "rds:DescribeDBMajorEngineVersions")
		} else if strings.Contains(err.Error(), "InvalidParameterValue") {
			s.logger.Error("Invalid engine parameter provided to AWS API", "engine", engine)
		} else if strings.Contains(err.Error(), "RequestLimitExceeded") || strings.Contains(err.Error(), "Throttling") {
			s.logger.Error("AWS API rate limit exceeded for DescribeDBMajorEngineVersions", "engine", engine)
		}

		return nil, fmt.Errorf("failed to describe DB major engine versions for %s: %w", engine, err)
	}

	// Validate API response
	if output == nil {
		s.logger.Error("Received nil response from DescribeDBMajorEngineVersions API", "engine", engine)
		return nil, fmt.Errorf("received nil response from AWS API for engine %s", engine)
	}

	// Parse API response and extract lifecycle data
	lifecycles := s.parseEngineVersionLifecycles(engine, output)

	// Log if no lifecycle data was found
	if len(lifecycles) == 0 {
		s.logger.Debug("No engine lifecycle data found in API response", "engine", engine)
	}

	// Store in cache with TTL
	s.cache.Set(cacheKey, lifecycles, cache.DefaultExpiration)

	s.logger.Debug("Fetched and cached engine lifecycle data", "engine", engine, "count", len(lifecycles))

	return lifecycles, nil
}

// GenerateCacheKey creates a cache key for the given engine
func (s *EngineSupportService) GenerateCacheKey(engine string) string {
	return fmt.Sprintf("engine_lifecycle_%s", engine)
}

// GetCache returns the cache instance for testing purposes
func (s *EngineSupportService) GetCache() *cache.Cache {
	return s.cache
}

// parseEngineVersionLifecycles parses the AWS API response and extracts lifecycle data
func (s *EngineSupportService) parseEngineVersionLifecycles(engine string, output *aws_rds.DescribeDBMajorEngineVersionsOutput) []EngineVersionLifecycle {
	var lifecycles []EngineVersionLifecycle

	if output == nil {
		s.logger.Error("Received nil output in parseEngineVersionLifecycles", "engine", engine)
		return lifecycles
	}

	if output.DBMajorEngineVersions == nil {
		s.logger.Debug("No DBMajorEngineVersions in API response", "engine", engine)
		return lifecycles
	}

	s.logger.Debug("Parsing engine version lifecycle data",
		"engine", engine,
		"version_count", len(output.DBMajorEngineVersions))

	for i, majorVersion := range output.DBMajorEngineVersions {
		if majorVersion.MajorEngineVersion == nil {
			s.logger.Debug("Skipping engine version with nil MajorEngineVersion",
				"engine", engine,
				"index", i)
			continue
		}

		versionString := aws.ToString(majorVersion.MajorEngineVersion)
		lifecycle := EngineVersionLifecycle{
			Engine:             engine,
			MajorEngineVersion: versionString,
		}

		// Parse supported engine lifecycles to extract standard and extended support end dates
		if majorVersion.SupportedEngineLifecycles == nil {
			s.logger.Debug("No SupportedEngineLifecycles found for engine version",
				"engine", engine,
				"version", versionString)
		} else {
			s.logger.Debug("Processing supported lifecycles",
				"engine", engine,
				"version", versionString,
				"lifecycle_count", len(majorVersion.SupportedEngineLifecycles))

			for j, supportedLifecycle := range majorVersion.SupportedEngineLifecycles {
				if supportedLifecycle.LifecycleSupportEndDate == nil {
					s.logger.Debug("Skipping lifecycle with nil LifecycleSupportEndDate",
						"engine", engine,
						"version", versionString,
						"lifecycle_index", j,
						"support_name", string(supportedLifecycle.LifecycleSupportName))
					continue
				}

				supportName := strings.ToLower(string(supportedLifecycle.LifecycleSupportName))
				endDate := supportedLifecycle.LifecycleSupportEndDate

				s.logger.Debug("Processing lifecycle support",
					"engine", engine,
					"version", versionString,
					"support_name", supportName,
					"end_date", endDate.Format("2006-01-02"))

				// Map support types to our lifecycle structure
				if strings.Contains(supportName, "standard") {
					lifecycle.StandardSupportEndDate = endDate
					s.logger.Debug("Set standard support end date",
						"engine", engine,
						"version", versionString,
						"date", endDate.Format("2006-01-02"))
				} else if strings.Contains(supportName, "extended") {
					lifecycle.ExtendedSupportEndDate = endDate
					s.logger.Debug("Set extended support end date",
						"engine", engine,
						"version", versionString,
						"date", endDate.Format("2006-01-02"))
				} else {
					s.logger.Debug("Unknown support type, skipping",
						"engine", engine,
						"version", versionString,
						"support_name", supportName)
				}
			}
		}

		// Only add lifecycle if we have at least one support end date
		if lifecycle.StandardSupportEndDate != nil || lifecycle.ExtendedSupportEndDate != nil {
			lifecycles = append(lifecycles, lifecycle)
			s.logger.Debug("Added lifecycle data for engine version",
				"engine", engine,
				"version", versionString,
				"has_standard", lifecycle.StandardSupportEndDate != nil,
				"has_extended", lifecycle.ExtendedSupportEndDate != nil)
		} else {
			s.logger.Debug("Skipping engine version with no support end dates",
				"engine", engine,
				"version", versionString)
		}
	}

	s.logger.Debug("Completed parsing engine version lifecycle data",
		"engine", engine,
		"total_versions", len(output.DBMajorEngineVersions),
		"parsed_lifecycles", len(lifecycles))

	return lifecycles
}

// CalculateRemainingDays calculates the remaining days until the given end date
func (s *EngineSupportService) CalculateRemainingDays(endDate *time.Time) *int64 {
	if endDate == nil {
		return nil
	}

	now := time.Now()
	remaining := endDate.Sub(now).Hours() / 24
	days := int64(math.Ceil(remaining))

	return &days
}
