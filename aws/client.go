package aws

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// CustodianClient holds all AWS service clients
type CustodianClient struct {
	Config  aws.Config
	EC2     *ec2.Client
	S3      *s3.Client
	RDS     *rds.Client
	Lambda  *lambda.Client
	IAM     *iam.Client
	Region  string
	Profile string
	DryRun  bool
}

// ClientConfig for initializing the AWS client
type ClientConfig struct {
	Region          string
	Profile         string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
	AssumeRoleARN   string
	DryRun          bool
	Timeout         time.Duration
}

// NewCustodianClient creates a new AWS client for Custodian Killer
func NewCustodianClient(cfg ClientConfig) (*CustodianClient, error) {
	fmt.Println("🚀 Initializing AWS SDK clients...")

	// Set default region if not provided
	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	// Set default timeout
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	ctx := context.Background()

	// Load AWS configuration
	var awsConfig aws.Config
	var err error

	// Configure based on available credentials
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		// Use provided credentials
		fmt.Println("🔑 Using provided AWS credentials")
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				cfg.SessionToken,
			)),
		)
	} else if cfg.Profile != "" {
		// Use named profile
		fmt.Printf("👤 Using AWS profile: %s\n", cfg.Profile)
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
			config.WithSharedConfigProfile(cfg.Profile),
		)
	} else {
		// Use default credential chain (environment, instance profile, etc.)
		fmt.Println("🔧 Using default AWS credential chain")
		awsConfig, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(cfg.Region),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("🚨 failed to load AWS configuration: %v", err)
	}

	// Create service clients
	client := &CustodianClient{
		Config:  awsConfig,
		EC2:     ec2.NewFromConfig(awsConfig),
		S3:      s3.NewFromConfig(awsConfig),
		RDS:     rds.NewFromConfig(awsConfig),
		Lambda:  lambda.NewFromConfig(awsConfig),
		IAM:     iam.NewFromConfig(awsConfig),
		Region:  cfg.Region,
		Profile: cfg.Profile,
		DryRun:  cfg.DryRun,
	}

	// Test connection
	if err := client.TestConnection(); err != nil {
		return nil, fmt.Errorf("🚨 AWS connection test failed: %v", err)
	}

	fmt.Printf("✅ AWS clients initialized successfully in region: %s\n", cfg.Region)
	return client, nil
}

// TestConnection verifies AWS connectivity
func (c *CustodianClient) TestConnection() error {
	fmt.Println("🧪 Testing AWS connection...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to describe regions (lightweight call)
	input := &ec2.DescribeRegionsInput{}
	if c.DryRun {
		input.DryRun = aws.Bool(true)
	}

	result, err := c.EC2.DescribeRegions(ctx, input)
	if err != nil {
		// If dry-run fails due to permissions, that's still a successful connection
		if c.DryRun && isUnauthorizedError(err) {
			fmt.Println("✅ AWS connection successful (dry-run permission denied is expected)")
			return nil
		}
		return fmt.Errorf("connection test failed: %v", err)
	}

	fmt.Printf("✅ Connected to AWS! Available regions: %d\n", len(result.Regions))
	return nil
}

// GetCallerIdentity returns information about the AWS credentials being used
func (c *CustodianClient) GetCallerIdentity() (*CallerInfo, error) {
	// We'd need to add STS client for this, but for now return basic info
	return &CallerInfo{
		Account: "unknown",
		UserID:  "unknown",
		Arn:     "unknown",
		Region:  c.Region,
		Profile: c.Profile,
	}, nil
}

// CallerInfo contains information about the current AWS caller
type CallerInfo struct {
	Account string
	UserID  string
	Arn     string
	Region  string
	Profile string
}

// SetDryRun enables or disables dry-run mode
func (c *CustodianClient) SetDryRun(dryRun bool) {
	c.DryRun = dryRun
	if dryRun {
		fmt.Println("🧪 Dry-run mode ENABLED - no actual changes will be made")
	} else {
		fmt.Println("💥 Live mode ENABLED - changes will be made to AWS resources!")
	}
}

// GetRegions returns list of available AWS regions
func (c *CustodianClient) GetRegions() ([]string, error) {
	ctx := context.Background()
	result, err := c.EC2.DescribeRegions(ctx, &ec2.DescribeRegionsInput{})
	if err != nil {
		return nil, err
	}

	var regions []string
	for _, region := range result.Regions {
		if region.RegionName != nil {
			regions = append(regions, *region.RegionName)
		}
	}

	return regions, nil
}

// SwitchRegion changes the active region for all clients
func (c *CustodianClient) SwitchRegion(region string) error {
	fmt.Printf("🌍 Switching to region: %s\n", region)

	// Update config
	c.Config.Region = region
	c.Region = region

	// Recreate clients with new region
	c.EC2 = ec2.NewFromConfig(c.Config)
	c.S3 = s3.NewFromConfig(c.Config)
	c.RDS = rds.NewFromConfig(c.Config)
	c.Lambda = lambda.NewFromConfig(c.Config)
	c.IAM = iam.NewFromConfig(c.Config)

	// Test new connection
	if err := c.TestConnection(); err != nil {
		return fmt.Errorf("failed to switch to region %s: %v", region, err)
	}

	fmt.Printf("✅ Switched to region: %s\n", region)
	return nil
}

// Helper function to check if error is unauthorized (for dry-run detection)
func isUnauthorizedError(err error) bool {
	return err != nil && (fmt.Sprintf("%v", err) == "UnauthorizedOperation" ||
		fmt.Sprintf("%v", err) == "DryRunOperation")
}

// LogAWSCall logs AWS API calls for debugging
func (c *CustodianClient) LogAWSCall(service, operation string, dryRun bool) {
	dryRunStr := ""
	if dryRun {
		dryRunStr = " (DRY RUN)"
	}

	if os.Getenv("CUSTODIAN_DEBUG") == "true" {
		log.Printf("🔍 AWS API Call: %s.%s%s in %s", service, operation, dryRunStr, c.Region)
	}
}

// GetServiceQuotas returns current service quotas (mock for now)
func (c *CustodianClient) GetServiceQuotas() map[string]interface{} {
	return map[string]interface{}{
		"ec2_instances":    20,
		"s3_buckets":       100,
		"rds_instances":    40,
		"lambda_functions": 1000,
		"region":           c.Region,
		"dry_run_mode":     c.DryRun,
	}
}

// Close gracefully closes the AWS client (cleanup if needed)
func (c *CustodianClient) Close() error {
	fmt.Println("🧹 Cleaning up AWS client connections...")
	// No explicit cleanup needed for AWS SDK v2, but we could add connection pooling cleanup here
	return nil
}

// Batch operation helper for processing resources in chunks
func (c *CustodianClient) ProcessInBatches(
	items []string,
	batchSize int,
	processor func([]string) error,
) error {
	if batchSize <= 0 {
		batchSize = 10 // Default batch size
	}

	for i := 0; i < len(items); i += batchSize {
		end := i + batchSize
		if end > len(items) {
			end = len(items)
		}

		batch := items[i:end]
		fmt.Printf("🔄 Processing batch %d-%d of %d items...\n", i+1, end, len(items))

		if err := processor(batch); err != nil {
			return fmt.Errorf("batch processing failed at items %d-%d: %v", i+1, end, err)
		}
	}

	return nil
}

// WaitForCompletion waits for async operations to complete
func (c *CustodianClient) WaitForCompletion(
	ctx context.Context,
	checkFunc func() (bool, error),
	maxWait time.Duration,
) error {
	timeout := time.After(maxWait)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("operation timed out after %v", maxWait)
		case <-ticker.C:
			completed, err := checkFunc()
			if err != nil {
				return err
			}
			if completed {
				return nil
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
