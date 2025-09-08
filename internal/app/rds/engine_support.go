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
	Engine                   string     // e.g., "postgres"
	MajorEngineVersion      string     // e.g., "13"
	StandardSupportEndDate  *time.Time // End of standard support
	ExtendedSupportEndDate  *time.Time // End of extended support
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

// GetEngineSupportMetrics retrieves engine support metrics for a given engine and version
func (s *EngineSupportService) GetEngineSupportMetrics(ctx context.Context, engine, version string) (EngineSupportMetrics, error) {
	// Fetch engine version lifecycle data (uses cache if available)
	lifecycles, err := s.FetchEngineVersionLifecycles(ctx, engine)
	if err != nil {
		return EngineSupportMetrics{}, fmt.Errorf("failed to fetch engine lifecycle data: %w", err)
	}
	
	// Find the lifecycle data for the specific version
	var targetLifecycle *EngineVersionLifecycle
	for _, lifecycle := range lifecycles {
		if lifecycle.MajorEngineVersion == version {
			targetLifecycle = &lifecycle
			break
		}
	}
	
	// If no lifecycle data found for this version, return empty metrics
	if targetLifecycle == nil {
		s.logger.Debug("No lifecycle data found for engine version", "engine", engine, "version", version)
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
	// Generate cache key for this engine
	cacheKey := s.GenerateCacheKey(engine)
	
	// Check cache first
	if cached, found := s.cache.Get(cacheKey); found {
		if lifecycles, ok := cached.([]EngineVersionLifecycle); ok {
			s.logger.Debug("Retrieved engine lifecycle data from cache", "engine", engine, "count", len(lifecycles))
			return lifecycles, nil
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
		return nil, fmt.Errorf("failed to describe DB major engine versions for %s: %w", engine, err)
	}
	
	// Parse API response and extract lifecycle data
	lifecycles := s.parseEngineVersionLifecycles(engine, output)
	
	// Store in cache with TTL
	s.cache.Set(cacheKey, lifecycles, cache.DefaultExpiration)
	
	s.logger.Debug("Fetched and cached engine lifecycle data", "engine", engine, "count", len(lifecycles))
	
	return lifecycles, nil
}

// GenerateCacheKey creates a cache key for the given engine
func (s *EngineSupportService) GenerateCacheKey(engine string) string {
	return fmt.Sprintf("engine_lifecycle_%s", engine)
}

// parseEngineVersionLifecycles parses the AWS API response and extracts lifecycle data
func (s *EngineSupportService) parseEngineVersionLifecycles(engine string, output *aws_rds.DescribeDBMajorEngineVersionsOutput) []EngineVersionLifecycle {
	var lifecycles []EngineVersionLifecycle
	
	if output == nil || output.DBMajorEngineVersions == nil {
		return lifecycles
	}
	
	for _, majorVersion := range output.DBMajorEngineVersions {
		if majorVersion.MajorEngineVersion == nil {
			continue
		}
		
		lifecycle := EngineVersionLifecycle{
			Engine:             engine,
			MajorEngineVersion: aws.ToString(majorVersion.MajorEngineVersion),
		}
		
		// Parse supported engine lifecycles to extract standard and extended support end dates
		if majorVersion.SupportedEngineLifecycles != nil {
			for _, supportedLifecycle := range majorVersion.SupportedEngineLifecycles {
				if supportedLifecycle.LifecycleSupportEndDate == nil {
					continue
				}
				
				supportName := strings.ToLower(string(supportedLifecycle.LifecycleSupportName))
				endDate := supportedLifecycle.LifecycleSupportEndDate
				
				// Map support types to our lifecycle structure
				if strings.Contains(supportName, "standard") {
					lifecycle.StandardSupportEndDate = endDate
				} else if strings.Contains(supportName, "extended") {
					lifecycle.ExtendedSupportEndDate = endDate
				}
			}
		}
		
		// Only add lifecycle if we have at least one support end date
		if lifecycle.StandardSupportEndDate != nil || lifecycle.ExtendedSupportEndDate != nil {
			lifecycles = append(lifecycles, lifecycle)
		}
	}
	
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