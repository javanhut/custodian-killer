package aws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Bucket represents an S3 bucket with security and cost info
type S3Bucket struct {
	Name                  string            `json:"name"`
	Region                string            `json:"region"`
	CreationDate          time.Time         `json:"creation_date"`
	Tags                  map[string]string `json:"tags"`
	PublicReadACL         bool              `json:"public_read_acl"`
	PublicWriteACL        bool              `json:"public_write_acl"`
	PublicReadPolicy      bool              `json:"public_read_policy"`
	PublicWritePolicy     bool              `json:"public_write_policy"`
	BlockPublicACLs       bool              `json:"block_public_acls"`
	BlockPublicPolicy     bool              `json:"block_public_policy"`
	IgnorePublicACLs      bool              `json:"ignore_public_acls"`
	RestrictPublicBuckets bool              `json:"restrict_public_buckets"`
	Versioning            string            `json:"versioning"` // Enabled, Suspended, or Disabled
	Encryption            S3Encryption      `json:"encryption"`
	ObjectCount           int64             `json:"object_count"`
	SizeBytes             int64             `json:"size_bytes"`
	StorageClass          map[string]int64  `json:"storage_class_breakdown"`
	MonthlyCostEstimate   float64           `json:"monthly_cost_estimate"`
	SecurityScore         int               `json:"security_score"` // 0-100
	ComplianceIssues      []string          `json:"compliance_issues"`
}

// S3Encryption represents bucket encryption details
type S3Encryption struct {
	Enabled          bool   `json:"enabled"`
	Algorithm        string `json:"algorithm,omitempty"` // AES256, aws:kms
	KMSKeyID         string `json:"kms_key_id,omitempty"`
	BucketKeyEnabled bool   `json:"bucket_key_enabled"`
}

// S3Filter represents filtering criteria for S3 buckets
type S3Filter struct {
	BucketNames        []string          // specific bucket names
	Tags               map[string]string // tag filters
	CreatedAfter       *time.Time        // buckets created after this time
	CreatedBefore      *time.Time        // buckets created before this time
	PublicAccessOnly   bool              // only public buckets
	UnencryptedOnly    bool              // only unencrypted buckets
	LargeSizeThreshold *int64            // buckets larger than this size
	MinSecurityScore   *int              // minimum security score
}

// GetS3Buckets retrieves S3 buckets based on filters
func (c *CustodianClient) GetS3Buckets(filters S3Filter) ([]S3Bucket, error) {
	c.LogAWSCall("S3", "ListBuckets", c.DryRun)

	fmt.Println("🪣 Scanning S3 buckets...")

	ctx := context.Background()

	// First, list all buckets
	listResult, err := c.S3.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %v", err)
	}

	var buckets []S3Bucket

	for _, bucket := range listResult.Buckets {
		bucketName := aws.ToString(bucket.Name)

		// Apply name filter early to avoid unnecessary API calls
		if len(filters.BucketNames) > 0 && !contains(filters.BucketNames, bucketName) {
			continue
		}

		fmt.Printf("📊 Analyzing bucket: %s\n", bucketName)

		s3Bucket, err := c.analyzeBucket(bucketName, aws.ToTime(bucket.CreationDate))
		if err != nil {
			fmt.Printf("⚠️  Warning: Failed to analyze bucket %s: %v\n", bucketName, err)
			continue
		}

		// Apply filters
		if c.matchesS3Filters(s3Bucket, filters) {
			buckets = append(buckets, s3Bucket)
		}
	}

	fmt.Printf("✅ Found %d S3 buckets matching criteria\n", len(buckets))
	return buckets, nil
}

// analyzeBucket performs deep analysis of a single bucket
func (c *CustodianClient) analyzeBucket(
	bucketName string,
	creationDate time.Time,
) (S3Bucket, error) {
	ctx := context.Background()

	bucket := S3Bucket{
		Name:             bucketName,
		CreationDate:     creationDate,
		Tags:             make(map[string]string),
		StorageClass:     make(map[string]int64),
		ComplianceIssues: make([]string, 0),
	}

	// Get bucket location
	locationResult, err := c.S3.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		bucket.Region = "us-east-1" // Default
	} else {
		if locationResult.LocationConstraint == "" {
			bucket.Region = "us-east-1"
		} else {
			bucket.Region = string(locationResult.LocationConstraint)
		}
	}

	// Get bucket tags
	bucket.Tags = c.getBucketTags(bucketName)

	// Analyze public access
	c.analyzeBucketPublicAccess(&bucket)

	// Get public access block configuration
	c.getBucketPublicAccessBlock(&bucket)

	// Get versioning status
	c.getBucketVersioning(&bucket)

	// Get encryption configuration
	c.getBucketEncryption(&bucket)

	// Get size and object count (this would be expensive for real buckets)
	c.estimateBucketSize(&bucket)

	// Calculate security score
	bucket.SecurityScore = c.calculateSecurityScore(bucket)

	// Identify compliance issues
	bucket.ComplianceIssues = c.identifyComplianceIssues(bucket)

	return bucket, nil
}

// getBucketTags retrieves bucket tags
func (c *CustodianClient) getBucketTags(bucketName string) map[string]string {
	ctx := context.Background()
	tags := make(map[string]string)

	result, err := c.S3.GetBucketTagging(ctx, &s3.GetBucketTaggingInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		// No tags or access denied
		return tags
	}

	for _, tag := range result.TagSet {
		tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	return tags
}

// analyzeBucketPublicAccess checks for public access via ACL
func (c *CustodianClient) analyzeBucketPublicAccess(bucket *S3Bucket) {
	ctx := context.Background()

	// Check bucket ACL
	aclResult, err := c.S3.GetBucketAcl(ctx, &s3.GetBucketAclInput{
		Bucket: aws.String(bucket.Name),
	})
	if err != nil {
		return
	}

	// Check for public access grants
	for _, grant := range aclResult.Grants {
		if grant.Grantee != nil {
			switch grant.Grantee.Type {
			case types.TypeGroup:
				uri := aws.ToString(grant.Grantee.URI)
				if strings.Contains(uri, "AllUsers") {
					if grant.Permission == types.PermissionRead {
						bucket.PublicReadACL = true
					}
					if grant.Permission == types.PermissionWrite {
						bucket.PublicWriteACL = true
					}
				}
			}
		}
	}
}

// getBucketPublicAccessBlock gets the public access block configuration
func (c *CustodianClient) getBucketPublicAccessBlock(bucket *S3Bucket) {
	ctx := context.Background()

	result, err := c.S3.GetPublicAccessBlock(ctx, &s3.GetPublicAccessBlockInput{
		Bucket: aws.String(bucket.Name),
	})
	if err != nil {
		// No public access block configured (more permissive)
		bucket.BlockPublicACLs = false
		bucket.BlockPublicPolicy = false
		bucket.IgnorePublicACLs = false
		bucket.RestrictPublicBuckets = false
		return
	}

	config := result.PublicAccessBlockConfiguration
	bucket.BlockPublicACLs = aws.ToBool(config.BlockPublicAcls)
	bucket.BlockPublicPolicy = aws.ToBool(config.BlockPublicPolicy)
	bucket.IgnorePublicACLs = aws.ToBool(config.IgnorePublicAcls)
	bucket.RestrictPublicBuckets = aws.ToBool(config.RestrictPublicBuckets)
}

// getBucketVersioning gets versioning configuration
func (c *CustodianClient) getBucketVersioning(bucket *S3Bucket) {
	ctx := context.Background()

	result, err := c.S3.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
		Bucket: aws.String(bucket.Name),
	})
	if err != nil {
		bucket.Versioning = "Disabled"
		return
	}

	if result.Status == "" {
		bucket.Versioning = "Disabled"
	} else {
		bucket.Versioning = string(result.Status)
	}
}

// getBucketEncryption gets encryption configuration
func (c *CustodianClient) getBucketEncryption(bucket *S3Bucket) {
	ctx := context.Background()

	result, err := c.S3.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucket.Name),
	})
	if err != nil {
		bucket.Encryption = S3Encryption{Enabled: false}
		return
	}

	if len(result.ServerSideEncryptionConfiguration.Rules) > 0 {
		rule := result.ServerSideEncryptionConfiguration.Rules[0]
		encryption := S3Encryption{Enabled: true}

		if rule.ApplyServerSideEncryptionByDefault != nil {
			encryption.Algorithm = string(rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm)
			if rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID != nil {
				encryption.KMSKeyID = aws.ToString(
					rule.ApplyServerSideEncryptionByDefault.KMSMasterKeyID,
				)
			}
		}

		if rule.BucketKeyEnabled != nil {
			encryption.BucketKeyEnabled = aws.ToBool(rule.BucketKeyEnabled)
		}

		bucket.Encryption = encryption
	} else {
		bucket.Encryption = S3Encryption{Enabled: false}
	}
}

// estimateBucketSize estimates bucket size and object count
func (c *CustodianClient) estimateBucketSize(bucket *S3Bucket) {
	// In a real implementation, this would use CloudWatch metrics or iterate through objects
	// For now, we'll simulate based on bucket age and name patterns

	ageInDays := int(time.Since(bucket.CreationDate).Hours() / 24)

	// Simulate bucket sizes based on naming patterns and age
	switch {
	case strings.Contains(strings.ToLower(bucket.Name), "log"):
		bucket.ObjectCount = int64(ageInDays * 1000)           // Lots of log files
		bucket.SizeBytes = int64(ageInDays * 50 * 1024 * 1024) // 50MB per day
		bucket.StorageClass["STANDARD"] = bucket.SizeBytes / 2
		bucket.StorageClass["STANDARD_IA"] = bucket.SizeBytes / 4
		bucket.StorageClass["GLACIER"] = bucket.SizeBytes / 4

	case strings.Contains(strings.ToLower(bucket.Name), "backup"):
		bucket.ObjectCount = int64(ageInDays * 50)
		bucket.SizeBytes = int64(ageInDays * 500 * 1024 * 1024) // 500MB per day
		bucket.StorageClass["STANDARD_IA"] = bucket.SizeBytes / 3
		bucket.StorageClass["GLACIER"] = bucket.SizeBytes * 2 / 3

	case strings.Contains(strings.ToLower(bucket.Name), "media"):
		bucket.ObjectCount = int64(ageInDays * 10)
		bucket.SizeBytes = int64(ageInDays * 1024 * 1024 * 1024) // 1GB per day
		bucket.StorageClass["STANDARD"] = bucket.SizeBytes

	default:
		bucket.ObjectCount = int64(ageInDays * 100)
		bucket.SizeBytes = int64(ageInDays * 10 * 1024 * 1024) // 10MB per day
		bucket.StorageClass["STANDARD"] = bucket.SizeBytes
	}

	// Calculate estimated monthly cost
	bucket.MonthlyCostEstimate = c.estimateS3Cost(bucket.SizeBytes, bucket.StorageClass)
}

// estimateS3Cost calculates monthly S3 storage cost
func (c *CustodianClient) estimateS3Cost(totalSize int64, storageClasses map[string]int64) float64 {
	// S3 pricing per GB per month (simplified, US East 1)
	pricing := map[string]float64{
		"STANDARD":     0.023,
		"STANDARD_IA":  0.0125,
		"GLACIER":      0.004,
		"DEEP_ARCHIVE": 0.00099,
	}

	totalCost := 0.0

	for class, size := range storageClasses {
		if price, exists := pricing[class]; exists {
			sizeGB := float64(size) / (1024 * 1024 * 1024) // Convert bytes to GB
			totalCost += sizeGB * price
		}
	}

	// Add request costs (simplified)
	totalCost += 0.50 // Rough estimate for requests

	return totalCost
}

// calculateSecurityScore assigns a security score to the bucket
func (c *CustodianClient) calculateSecurityScore(bucket S3Bucket) int {
	score := 100 // Start with perfect score

	// Deduct points for security issues
	if bucket.PublicReadACL {
		score -= 30
	}
	if bucket.PublicWriteACL {
		score -= 40
	}
	if bucket.PublicReadPolicy {
		score -= 25
	}
	if bucket.PublicWritePolicy {
		score -= 35
	}
	if !bucket.Encryption.Enabled {
		score -= 20
	}
	if bucket.Versioning == "Disabled" {
		score -= 10
	}
	if !bucket.BlockPublicACLs {
		score -= 15
	}
	if !bucket.BlockPublicPolicy {
		score -= 15
	}

	// Bonus points for good practices
	if bucket.Encryption.Algorithm == "aws:kms" {
		score += 5
	}
	if bucket.Versioning == "Enabled" {
		score += 5
	}

	if score < 0 {
		score = 0
	}

	return score
}

// identifyComplianceIssues finds compliance violations
func (c *CustodianClient) identifyComplianceIssues(bucket S3Bucket) []string {
	var issues []string

	if bucket.PublicReadACL || bucket.PublicReadPolicy {
		issues = append(issues, "Bucket allows public read access")
	}
	if bucket.PublicWriteACL || bucket.PublicWritePolicy {
		issues = append(issues, "Bucket allows public write access")
	}
	if !bucket.Encryption.Enabled {
		issues = append(issues, "Bucket encryption is not enabled")
	}
	if bucket.Versioning == "Disabled" {
		issues = append(issues, "Bucket versioning is disabled")
	}
	if !bucket.BlockPublicACLs || !bucket.BlockPublicPolicy {
		issues = append(issues, "Public access block is not fully configured")
	}

	// Check for required tags
	requiredTags := []string{"Environment", "Owner", "Project"}
	for _, tag := range requiredTags {
		if _, exists := bucket.Tags[tag]; !exists {
			issues = append(issues, fmt.Sprintf("Missing required tag: %s", tag))
		}
	}

	return issues
}

// matchesS3Filters checks if bucket matches the specified filters
func (c *CustodianClient) matchesS3Filters(bucket S3Bucket, filters S3Filter) bool {
	// Creation date filters
	if filters.CreatedAfter != nil && bucket.CreationDate.Before(*filters.CreatedAfter) {
		return false
	}
	if filters.CreatedBefore != nil && bucket.CreationDate.After(*filters.CreatedBefore) {
		return false
	}

	// Public access filter
	if filters.PublicAccessOnly {
		hasPublicAccess := bucket.PublicReadACL || bucket.PublicWriteACL ||
			bucket.PublicReadPolicy || bucket.PublicWritePolicy
		if !hasPublicAccess {
			return false
		}
	}

	// Encryption filter
	if filters.UnencryptedOnly && bucket.Encryption.Enabled {
		return false
	}

	// Size filter
	if filters.LargeSizeThreshold != nil && bucket.SizeBytes < *filters.LargeSizeThreshold {
		return false
	}

	// Security score filter
	if filters.MinSecurityScore != nil && bucket.SecurityScore < *filters.MinSecurityScore {
		return false
	}

	// Tag filters
	for key, value := range filters.Tags {
		bucketValue, exists := bucket.Tags[key]
		if !exists || (value != "*" && bucketValue != value) {
			return false
		}
	}

	return true
}

// BlockPublicAccess blocks public access on S3 buckets
func (c *CustodianClient) BlockPublicAccess(bucketNames []string) (*S3ActionResult, error) {
	c.LogAWSCall("S3", "PutPublicAccessBlock", c.DryRun)

	if len(bucketNames) == 0 {
		return &S3ActionResult{}, nil
	}

	fmt.Printf("🔒 Blocking public access on %d buckets...\n", len(bucketNames))

	ctx := context.Background()
	result := &S3ActionResult{
		Action:      "block-public-access",
		BucketNames: bucketNames,
		Success:     true,
		DryRun:      c.DryRun,
		Timestamp:   time.Now(),
		Results:     make(map[string]string),
	}

	for _, bucketName := range bucketNames {
		if c.DryRun {
			result.Results[bucketName] = "would block public access"
			continue
		}

		input := &s3.PutPublicAccessBlockInput{
			Bucket: aws.String(bucketName),
			PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
				BlockPublicAcls:       aws.Bool(true),
				BlockPublicPolicy:     aws.Bool(true),
				IgnorePublicAcls:      aws.Bool(true),
				RestrictPublicBuckets: aws.Bool(true),
			},
		}

		_, err := c.S3.PutPublicAccessBlock(ctx, input)
		if err != nil {
			result.Results[bucketName] = fmt.Sprintf("failed: %v", err)
			result.Success = false
		} else {
			result.Results[bucketName] = "public access blocked"
		}
	}

	fmt.Printf("✅ Public access block operation completed\n")
	return result, nil
}

// EnableEncryption enables server-side encryption on S3 buckets
func (c *CustodianClient) EnableEncryption(
	bucketNames []string,
	kmsKeyID string,
) (*S3ActionResult, error) {
	c.LogAWSCall("S3", "PutBucketEncryption", c.DryRun)

	if len(bucketNames) == 0 {
		return &S3ActionResult{}, nil
	}

	fmt.Printf("🔐 Enabling encryption on %d buckets...\n", len(bucketNames))

	ctx := context.Background()
	result := &S3ActionResult{
		Action:      "enable-encryption",
		BucketNames: bucketNames,
		Success:     true,
		DryRun:      c.DryRun,
		Timestamp:   time.Now(),
		Results:     make(map[string]string),
	}

	for _, bucketName := range bucketNames {
		if c.DryRun {
			result.Results[bucketName] = "would enable encryption"
			continue
		}

		// Configure encryption
		var encryptionRule types.ServerSideEncryptionRule
		if kmsKeyID != "" {
			encryptionRule = types.ServerSideEncryptionRule{
				ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{
					SSEAlgorithm:   types.ServerSideEncryptionAwsKms,
					KMSMasterKeyID: aws.String(kmsKeyID),
				},
				BucketKeyEnabled: aws.Bool(true),
			}
		} else {
			encryptionRule = types.ServerSideEncryptionRule{
				ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{
					SSEAlgorithm: types.ServerSideEncryptionAes256,
				},
			}
		}

		input := &s3.PutBucketEncryptionInput{
			Bucket: aws.String(bucketName),
			ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
				Rules: []types.ServerSideEncryptionRule{encryptionRule},
			},
		}

		_, err := c.S3.PutBucketEncryption(ctx, input)
		if err != nil {
			result.Results[bucketName] = fmt.Sprintf("failed: %v", err)
			result.Success = false
		} else {
			algorithm := "AES256"
			if kmsKeyID != "" {
				algorithm = "KMS"
			}
			result.Results[bucketName] = fmt.Sprintf("encryption enabled (%s)", algorithm)
		}
	}

	fmt.Printf("✅ Encryption operation completed\n")
	return result, nil
}

// EnableVersioning enables versioning on S3 buckets
func (c *CustodianClient) EnableVersioning(bucketNames []string) (*S3ActionResult, error) {
	c.LogAWSCall("S3", "PutBucketVersioning", c.DryRun)

	if len(bucketNames) == 0 {
		return &S3ActionResult{}, nil
	}

	fmt.Printf("📚 Enabling versioning on %d buckets...\n", len(bucketNames))

	ctx := context.Background()
	result := &S3ActionResult{
		Action:      "enable-versioning",
		BucketNames: bucketNames,
		Success:     true,
		DryRun:      c.DryRun,
		Timestamp:   time.Now(),
		Results:     make(map[string]string),
	}

	for _, bucketName := range bucketNames {
		if c.DryRun {
			result.Results[bucketName] = "would enable versioning"
			continue
		}

		input := &s3.PutBucketVersioningInput{
			Bucket: aws.String(bucketName),
			VersioningConfiguration: &types.VersioningConfiguration{
				Status: types.BucketVersioningStatusEnabled,
			},
		}

		_, err := c.S3.PutBucketVersioning(ctx, input)
		if err != nil {
			result.Results[bucketName] = fmt.Sprintf("failed: %v", err)
			result.Success = false
		} else {
			result.Results[bucketName] = "versioning enabled"
		}
	}

	fmt.Printf("✅ Versioning operation completed\n")
	return result, nil
}

// TagBuckets adds tags to S3 buckets
func (c *CustodianClient) TagBuckets(
	bucketNames []string,
	tags map[string]string,
) (*S3ActionResult, error) {
	c.LogAWSCall("S3", "PutBucketTagging", c.DryRun)

	if len(bucketNames) == 0 || len(tags) == 0 {
		return &S3ActionResult{}, nil
	}

	fmt.Printf("🏷️  Adding %d tags to %d buckets...\n", len(tags), len(bucketNames))

	ctx := context.Background()
	result := &S3ActionResult{
		Action:      "tag",
		BucketNames: bucketNames,
		Success:     true,
		DryRun:      c.DryRun,
		Timestamp:   time.Now(),
		Results:     make(map[string]string),
		Tags:        tags,
	}

	// Convert tags to S3 tag format
	var s3Tags []types.Tag
	for key, value := range tags {
		s3Tags = append(s3Tags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	for _, bucketName := range bucketNames {
		if c.DryRun {
			result.Results[bucketName] = "would add tags"
			continue
		}

		input := &s3.PutBucketTaggingInput{
			Bucket: aws.String(bucketName),
			Tagging: &types.Tagging{
				TagSet: s3Tags,
			},
		}

		_, err := c.S3.PutBucketTagging(ctx, input)
		if err != nil {
			result.Results[bucketName] = fmt.Sprintf("failed: %v", err)
			result.Success = false
		} else {
			result.Results[bucketName] = fmt.Sprintf("added %d tags", len(tags))
		}
	}

	fmt.Printf("✅ Tagging operation completed\n")
	return result, nil
}

// DeleteBuckets deletes S3 buckets (dangerous operation!)
func (c *CustodianClient) DeleteBuckets(bucketNames []string, force bool) (*S3ActionResult, error) {
	c.LogAWSCall("S3", "DeleteBucket", c.DryRun)

	if len(bucketNames) == 0 {
		return &S3ActionResult{}, nil
	}

	fmt.Printf("💀 Deleting %d buckets...\n", len(bucketNames))
	if !c.DryRun {
		fmt.Println("⚠️  WARNING: This action is IRREVERSIBLE!")
	}

	ctx := context.Background()
	result := &S3ActionResult{
		Action:      "delete",
		BucketNames: bucketNames,
		Success:     true,
		DryRun:      c.DryRun,
		Timestamp:   time.Now(),
		Results:     make(map[string]string),
	}

	for _, bucketName := range bucketNames {
		if c.DryRun {
			result.Results[bucketName] = "would delete bucket"
			continue
		}

		// First, try to empty the bucket if force is enabled
		if force {
			if err := c.emptyBucket(bucketName); err != nil {
				result.Results[bucketName] = fmt.Sprintf("failed to empty: %v", err)
				result.Success = false
				continue
			}
		}

		// Delete the bucket
		_, err := c.S3.DeleteBucket(ctx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			result.Results[bucketName] = fmt.Sprintf("failed: %v", err)
			result.Success = false
		} else {
			result.Results[bucketName] = "deleted successfully"
		}
	}

	fmt.Printf("✅ Delete operation completed\n")
	return result, nil
}

// emptyBucket removes all objects from a bucket
func (c *CustodianClient) emptyBucket(bucketName string) error {
	fmt.Printf("🗑️  Emptying bucket: %s\n", bucketName)

	ctx := context.Background()

	// List and delete objects in batches
	paginator := s3.NewListObjectsV2Paginator(c.S3, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to list objects: %v", err)
		}

		if len(page.Contents) == 0 {
			continue
		}

		// Prepare delete request
		var objectsToDelete []types.ObjectIdentifier
		for _, obj := range page.Contents {
			objectsToDelete = append(objectsToDelete, types.ObjectIdentifier{
				Key: obj.Key,
			})
		}

		// Delete objects in batch
		_, err = c.S3.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &types.Delete{
				Objects: objectsToDelete,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to delete objects: %v", err)
		}

		fmt.Printf("🗑️  Deleted %d objects from %s\n", len(objectsToDelete), bucketName)
	}

	return nil
}

// S3ActionResult represents the result of an S3 action
type S3ActionResult struct {
	Action      string            `json:"action"`
	BucketNames []string          `json:"bucket_names"`
	Success     bool              `json:"success"`
	DryRun      bool              `json:"dry_run"`
	Timestamp   time.Time         `json:"timestamp"`
	Results     map[string]string `json:"results"`        // bucket -> result message
	Tags        map[string]string `json:"tags,omitempty"` // for tagging operations
	Error       string            `json:"error,omitempty"`
}

// GetBucketCosts calculates total costs for buckets
func (c *CustodianClient) GetBucketCosts(buckets []S3Bucket) map[string]float64 {
	costs := map[string]float64{
		"total_monthly":     0,
		"standard_monthly":  0,
		"ia_monthly":        0,
		"glacier_monthly":   0,
		"potential_savings": 0,
	}

	for _, bucket := range buckets {
		costs["total_monthly"] += bucket.MonthlyCostEstimate

		// Break down by storage class
		for class, size := range bucket.StorageClass {
			sizeGB := float64(size) / (1024 * 1024 * 1024)
			switch class {
			case "STANDARD":
				costs["standard_monthly"] += sizeGB * 0.023
			case "STANDARD_IA":
				costs["ia_monthly"] += sizeGB * 0.0125
			case "GLACIER":
				costs["glacier_monthly"] += sizeGB * 0.004
			}
		}

		// Calculate potential savings for lifecycle optimization
		if standardSize, exists := bucket.StorageClass["STANDARD"]; exists {
			standardSizeGB := float64(standardSize) / (1024 * 1024 * 1024)
			// Estimate 30% of standard storage could move to IA
			potentialIA := standardSizeGB * 0.3
			savings := potentialIA * (0.023 - 0.0125) // Standard - IA price difference
			costs["potential_savings"] += savings
		}
	}

	return costs
}

// GetBucketSecuritySummary provides security overview
func (c *CustodianClient) GetBucketSecuritySummary(buckets []S3Bucket) map[string]interface{} {
	summary := map[string]interface{}{
		"total_buckets":         len(buckets),
		"public_read_buckets":   0,
		"public_write_buckets":  0,
		"unencrypted_buckets":   0,
		"no_versioning_buckets": 0,
		"avg_security_score":    0.0,
		"high_risk_buckets":     0,
		"compliance_violations": 0,
	}

	totalScore := 0
	complianceIssues := 0

	for _, bucket := range buckets {
		if bucket.PublicReadACL || bucket.PublicReadPolicy {
			summary["public_read_buckets"] = summary["public_read_buckets"].(int) + 1
		}
		if bucket.PublicWriteACL || bucket.PublicWritePolicy {
			summary["public_write_buckets"] = summary["public_write_buckets"].(int) + 1
		}
		if !bucket.Encryption.Enabled {
			summary["unencrypted_buckets"] = summary["unencrypted_buckets"].(int) + 1
		}
		if bucket.Versioning == "Disabled" {
			summary["no_versioning_buckets"] = summary["no_versioning_buckets"].(int) + 1
		}
		if bucket.SecurityScore < 50 {
			summary["high_risk_buckets"] = summary["high_risk_buckets"].(int) + 1
		}

		totalScore += bucket.SecurityScore
		complianceIssues += len(bucket.ComplianceIssues)
	}

	if len(buckets) > 0 {
		summary["avg_security_score"] = float64(totalScore) / float64(len(buckets))
	}
	summary["compliance_violations"] = complianceIssues

	return summary
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
