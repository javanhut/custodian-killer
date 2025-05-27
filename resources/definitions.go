// resources/definitions.go
package resources

// ResourceDefinition defines the structure and filterable fields for AWS resources
type ResourceDefinition struct {
	Name          string                     `json:"name"`
	Service       string                     `json:"service"`
	Description   string                     `json:"description"`
	Fields        map[string]FieldDefinition `json:"fields"`
	Relationships []RelationshipDefinition   `json:"relationships"`
	Actions       []string                   `json:"actions"`
	Tags          []string                   `json:"common_tags"`
}

// FieldDefinition defines a filterable field
type FieldDefinition struct {
	Type        string   `json:"type"` // string, int, float, bool, time, duration, array
	Description string   `json:"description"`
	Operators   []string `json:"operators"`             // Supported operators for this field
	Examples    []string `json:"examples"`              // Example values
	Required    bool     `json:"required"`              // Whether this field always exists
	Computed    bool     `json:"computed"`              // Whether this field is computed (like age, cost)
	EnumValues  []string `json:"enum_values,omitempty"` // For enum-like fields
}

// RelationshipDefinition defines relationships between resources
type RelationshipDefinition struct {
	Name        string `json:"name"`
	Type        string `json:"type"`      // one-to-one, one-to-many, many-to-many
	Target      string `json:"target"`    // Target resource type
	Direction   string `json:"direction"` // inbound, outbound, bidirectional
	Description string `json:"description"`
}

// Comprehensive AWS Resource Definitions
var AWSResourceDefinitions = map[string]ResourceDefinition{
	"ec2": {
		Name:        "ec2",
		Service:     "EC2",
		Description: "Elastic Compute Cloud instances",
		Fields: map[string]FieldDefinition{
			// Basic instance information
			"instance_id": {
				Type:        "string",
				Description: "Unique instance identifier",
				Operators: []string{
					"eq",
					"ne",
					"in",
					"not-in",
					"starts-with",
					"ends-with",
					"regex",
				},
				Examples: []string{"i-1234567890abcdef0"},
				Required: true,
			},
			"name": {
				Type:        "string",
				Description: "Instance name (from Name tag or instance ID)",
				Operators: []string{
					"eq",
					"ne",
					"in",
					"not-in",
					"contains",
					"not-contains",
					"starts-with",
					"ends-with",
					"regex",
					"empty",
					"not-empty",
				},
				Examples: []string{"web-server-1", "database-prod"},
			},
			"instance_type": {
				Type:        "string",
				Description: "EC2 instance type",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"t3.micro", "m5.large", "c5.xlarge"},
				EnumValues: []string{
					"t3.nano",
					"t3.micro",
					"t3.small",
					"t3.medium",
					"t3.large",
					"t3.xlarge",
					"t3.2xlarge",
					"m5.large",
					"m5.xlarge",
					"m5.2xlarge",
					"m5.4xlarge",
					"c5.large",
					"c5.xlarge",
					"c5.2xlarge",
					"r5.large",
					"r5.xlarge",
				},
				Required: true,
			},
			"state": {
				Type:        "string",
				Description: "Current instance state",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"running", "stopped", "terminated"},
				EnumValues: []string{
					"pending",
					"running",
					"shutting-down",
					"terminated",
					"stopping",
					"stopped",
				},
				Required: true,
			},
			"launch_time": {
				Type:        "time",
				Description: "When the instance was launched",
				Operators: []string{
					"eq",
					"ne",
					"gt",
					"gte",
					"lt",
					"lte",
					"between",
					"age-gt",
					"age-lt",
				},
				Examples: []string{"2024-01-15T10:30:00Z", "2024-01-15"},
				Required: true,
			},

			// Network information
			"public_ip": {
				Type:        "string",
				Description: "Public IP address",
				Operators:   []string{"eq", "ne", "exists", "not-exists", "starts-with", "regex"},
				Examples:    []string{"54.123.45.67", "192.168.1.100"},
			},
			"private_ip": {
				Type:        "string",
				Description: "Private IP address",
				Operators:   []string{"eq", "ne", "starts-with", "regex"},
				Examples:    []string{"10.0.1.100", "172.16.0.50"},
			},
			"vpc_id": {
				Type:        "string",
				Description: "VPC identifier",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"vpc-12345678"},
			},
			"subnet_id": {
				Type:        "string",
				Description: "Subnet identifier",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"subnet-12345678"},
			},
			"availability_zone": {
				Type:        "string",
				Description: "Availability zone",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"us-east-1a", "us-west-2b"},
			},
			"security_groups": {
				Type:        "array",
				Description: "List of security group IDs",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{"sg-12345678", "sg-87654321"},
			},

			// Platform and architecture
			"platform": {
				Type:        "string",
				Description: "Platform (Linux/Windows)",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"linux", "windows"},
				EnumValues:  []string{"linux", "windows"},
			},
			"architecture": {
				Type:        "string",
				Description: "CPU architecture",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"x86_64", "arm64"},
				EnumValues:  []string{"i386", "x86_64", "arm64"},
			},
			"hypervisor": {
				Type:        "string",
				Description: "Hypervisor type",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"xen", "nitro"},
				EnumValues:  []string{"ovm", "xen", "nitro"},
			},

			// Performance and utilization (computed fields)
			"cpu_utilization": {
				Type:        "float",
				Description: "Average CPU utilization percentage",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"5.2", "85.0"},
				Computed:    true,
			},
			"network_in": {
				Type:        "float",
				Description: "Network bytes in",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"1024000", "5000000"},
				Computed:    true,
			},
			"network_out": {
				Type:        "float",
				Description: "Network bytes out",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"2048000", "10000000"},
				Computed:    true,
			},
			"running_days": {
				Type:        "int",
				Description: "Days since instance was launched",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"7", "30", "365"},
				Computed:    true,
			},
			"uptime_percentage": {
				Type:        "float",
				Description: "Percentage of time instance has been running",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"95.5", "99.9"},
				Computed:    true,
			},

			// Cost information (computed)
			"monthly_cost": {
				Type:        "float",
				Description: "Estimated monthly cost in USD",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"8.76", "175.68"},
				Computed:    true,
			},
			"hourly_cost": {
				Type:        "float",
				Description: "Hourly cost in USD",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"0.012", "0.24"},
				Computed:    true,
			},

			// Storage
			"root_device_type": {
				Type:        "string",
				Description: "Root device type",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"ebs", "instance-store"},
				EnumValues:  []string{"ebs", "instance-store"},
			},
			"ebs_optimized": {
				Type:        "bool",
				Description: "Whether EBS optimized",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},

			// Tags (dynamic)
			"tags": {
				Type:        "map",
				Description: "Instance tags",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{`{"Environment": "prod"}`, `{"Owner": "team-a"}`},
			},
			"tag_count": {
				Type:        "int",
				Description: "Number of tags",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte"},
				Examples:    []string{"0", "5", "10"},
				Computed:    true,
			},
		},
		Relationships: []RelationshipDefinition{
			{
				Name:        "volumes",
				Type:        "one-to-many",
				Target:      "ebs",
				Direction:   "outbound",
				Description: "EBS volumes attached to this instance",
			},
			{
				Name:        "security_groups",
				Type:        "many-to-many",
				Target:      "security-group",
				Direction:   "outbound",
				Description: "Security groups assigned to this instance",
			},
			{
				Name:        "subnet",
				Type:        "many-to-one",
				Target:      "subnet",
				Direction:   "outbound",
				Description: "Subnet where instance is located",
			},
			{
				Name:        "vpc",
				Type:        "many-to-one",
				Target:      "vpc",
				Direction:   "outbound",
				Description: "VPC where instance is located",
			},
			{
				Name:        "key_pair",
				Type:        "many-to-one",
				Target:      "key-pair",
				Direction:   "outbound",
				Description: "Key pair for SSH access",
			},
			{
				Name:        "snapshots",
				Type:        "one-to-many",
				Target:      "snapshot",
				Direction:   "outbound",
				Description: "Snapshots created from this instance",
			},
		},
		Actions: []string{
			"start",
			"stop",
			"terminate",
			"reboot",
			"tag",
			"untag",
			"modify",
			"create-image",
			"create-snapshot",
		},
		Tags: []string{"Name", "Environment", "Owner", "Project", "CostCenter", "Application"},
	},

	"s3": {
		Name:        "s3",
		Service:     "S3",
		Description: "Simple Storage Service buckets",
		Fields: map[string]FieldDefinition{
			// Basic bucket information
			"name": {
				Type:        "string",
				Description: "Bucket name",
				Operators: []string{
					"eq",
					"ne",
					"in",
					"not-in",
					"contains",
					"not-contains",
					"starts-with",
					"ends-with",
					"regex",
				},
				Examples: []string{"my-app-bucket", "backup-bucket-2024"},
				Required: true,
			},
			"region": {
				Type:        "string",
				Description: "AWS region where bucket is located",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"us-east-1", "eu-west-1"},
				Required:    true,
			},
			"creation_date": {
				Type:        "time",
				Description: "When the bucket was created",
				Operators: []string{
					"eq",
					"ne",
					"gt",
					"gte",
					"lt",
					"lte",
					"between",
					"age-gt",
					"age-lt",
				},
				Examples: []string{"2024-01-01T00:00:00Z"},
				Required: true,
			},

			// Access control
			"public_read_acl": {
				Type:        "bool",
				Description: "Whether bucket allows public read via ACL",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},
			"public_write_acl": {
				Type:        "bool",
				Description: "Whether bucket allows public write via ACL",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},
			"public_read_policy": {
				Type:        "bool",
				Description: "Whether bucket allows public read via policy",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},
			"public_write_policy": {
				Type:        "bool",
				Description: "Whether bucket allows public write via policy",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},
			"block_public_acls": {
				Type:        "bool",
				Description: "Whether public ACLs are blocked",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},
			"block_public_policy": {
				Type:        "bool",
				Description: "Whether public policies are blocked",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},
			"ignore_public_acls": {
				Type:        "bool",
				Description: "Whether public ACLs are ignored",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},
			"restrict_public_buckets": {
				Type:        "bool",
				Description: "Whether public bucket access is restricted",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},

			// Versioning and backup
			"versioning": {
				Type:        "string",
				Description: "Versioning status",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"Enabled", "Suspended", "Disabled"},
				EnumValues:  []string{"Enabled", "Suspended", "Disabled"},
			},
			"mfa_delete": {
				Type:        "bool",
				Description: "Whether MFA delete is enabled",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},

			// Encryption
			"encrypted": {
				Type:        "bool",
				Description: "Whether bucket has encryption enabled",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},
			"encryption_algorithm": {
				Type:        "string",
				Description: "Encryption algorithm used",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"AES256", "aws:kms"},
				EnumValues:  []string{"AES256", "aws:kms"},
			},
			"kms_key_id": {
				Type:        "string",
				Description: "KMS key ID for encryption",
				Operators:   []string{"eq", "ne", "exists", "not-exists"},
				Examples: []string{
					"arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
				},
			},
			"bucket_key_enabled": {
				Type:        "bool",
				Description: "Whether S3 bucket key is enabled",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},

			// Size and usage (computed)
			"object_count": {
				Type:        "int",
				Description: "Number of objects in bucket",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"0", "1000", "1000000"},
				Computed:    true,
			},
			"size_bytes": {
				Type:        "int",
				Description: "Total size in bytes",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"1024", "1073741824"},
				Computed:    true,
			},
			"size_gb": {
				Type:        "float",
				Description: "Total size in GB",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"1.5", "100.0"},
				Computed:    true,
			},

			// Storage classes
			"storage_class_standard": {
				Type:        "int",
				Description: "Bytes in STANDARD storage class",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"1073741824"},
				Computed:    true,
			},
			"storage_class_ia": {
				Type:        "int",
				Description: "Bytes in STANDARD_IA storage class",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"536870912"},
				Computed:    true,
			},
			"storage_class_glacier": {
				Type:        "int",
				Description: "Bytes in GLACIER storage class",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"268435456"},
				Computed:    true,
			},

			// Cost (computed)
			"monthly_cost": {
				Type:        "float",
				Description: "Estimated monthly cost in USD",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"5.50", "150.00"},
				Computed:    true,
			},

			// Security and compliance (computed)
			"security_score": {
				Type:        "int",
				Description: "Security score (0-100)",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"85", "95"},
				Computed:    true,
			},
			"compliance_issues": {
				Type:        "array",
				Description: "List of compliance issues",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{"public-access", "no-encryption"},
				Computed:    true,
			},

			// Age (computed)
			"age_days": {
				Type:        "int",
				Description: "Days since bucket was created",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"30", "365"},
				Computed:    true,
			},

			// Tags
			"tags": {
				Type:        "map",
				Description: "Bucket tags",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{`{"Environment": "prod"}`, `{"Purpose": "backup"}`},
			},
			"tag_count": {
				Type:        "int",
				Description: "Number of tags",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte"},
				Examples:    []string{"0", "3", "10"},
				Computed:    true,
			},
		},
		Relationships: []RelationshipDefinition{
			{
				Name:        "cloudfront_distributions",
				Type:        "one-to-many",
				Target:      "cloudfront",
				Direction:   "inbound",
				Description: "CloudFront distributions using this bucket",
			},
			{
				Name:        "access_logs",
				Type:        "one-to-one",
				Target:      "s3",
				Direction:   "outbound",
				Description: "Bucket for access logs",
			},
		},
		Actions: []string{
			"delete",
			"tag",
			"untag",
			"encrypt",
			"block-public-access",
			"enable-versioning",
			"disable-versioning",
			"set-lifecycle",
			"set-policy",
		},
		Tags: []string{"Environment", "Owner", "Project", "Purpose", "DataClassification"},
	},

	"rds": {
		Name:        "rds",
		Service:     "RDS",
		Description: "Relational Database Service instances",
		Fields: map[string]FieldDefinition{
			// Basic information
			"db_instance_identifier": {
				Type:        "string",
				Description: "Database instance identifier",
				Operators: []string{
					"eq",
					"ne",
					"in",
					"not-in",
					"contains",
					"starts-with",
					"ends-with",
					"regex",
				},
				Examples: []string{"prod-database", "test-mysql-01"},
				Required: true,
			},
			"db_name": {
				Type:        "string",
				Description: "Database name",
				Operators: []string{
					"eq",
					"ne",
					"contains",
					"starts-with",
					"ends-with",
					"exists",
					"not-exists",
				},
				Examples: []string{"production", "staging"},
			},
			"engine": {
				Type:        "string",
				Description: "Database engine",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"mysql", "postgres", "oracle-ee", "sqlserver-ex"},
				EnumValues: []string{
					"mysql",
					"postgres",
					"mariadb",
					"oracle-ee",
					"oracle-se2",
					"oracle-se1",
					"oracle-se",
					"sqlserver-ee",
					"sqlserver-se",
					"sqlserver-ex",
					"sqlserver-web",
				},
				Required: true,
			},
			"engine_version": {
				Type:        "string",
				Description: "Database engine version",
				Operators: []string{
					"eq",
					"ne",
					"starts-with",
					"contains",
					"gt",
					"gte",
					"lt",
					"lte",
				},
				Examples: []string{"8.0.35", "13.7", "19.0.0.0.ru-2022-01.rur-2022-01.r1"},
			},
			"db_instance_class": {
				Type:        "string",
				Description: "Database instance class",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"db.t3.micro", "db.r5.large", "db.m5.xlarge"},
				EnumValues: []string{
					"db.t3.nano",
					"db.t3.micro",
					"db.t3.small",
					"db.t3.medium",
					"db.t3.large",
					"db.t3.xlarge",
					"db.t3.2xlarge",
					"db.r5.large",
					"db.r5.xlarge",
					"db.r5.2xlarge",
					"db.m5.large",
					"db.m5.xlarge",
				},
				Required: true,
			},
			"db_instance_status": {
				Type:        "string",
				Description: "Current instance status",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"available", "stopped", "stopping", "starting"},
				EnumValues: []string{
					"available",
					"backing-up",
					"creating",
					"deleting",
					"failed",
					"inaccessible-encryption-credentials",
					"incompatible-network",
					"incompatible-option-group",
					"incompatible-parameters",
					"incompatible-restore",
					"maintenance",
					"modifying",
					"rebooting",
					"renaming",
					"resetting-master-credentials",
					"restore-error",
					"starting",
					"stopped",
					"stopping",
					"storage-full",
					"storage-optimization",
					"upgrading",
				},
				Required: true,
			},
			"instance_create_time": {
				Type:        "time",
				Description: "When the instance was created",
				Operators: []string{
					"eq",
					"ne",
					"gt",
					"gte",
					"lt",
					"lte",
					"between",
					"age-gt",
					"age-lt",
				},
				Examples: []string{"2024-01-01T12:00:00Z"},
				Required: true,
			},

			// Network and security
			"publicly_accessible": {
				Type:        "bool",
				Description: "Whether instance is publicly accessible",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},
			"vpc_id": {
				Type:        "string",
				Description: "VPC identifier",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"vpc-12345678"},
			},
			"subnet_group": {
				Type:        "string",
				Description: "DB subnet group name",
				Operators:   []string{"eq", "ne", "contains", "starts-with"},
				Examples:    []string{"default", "private-subnets"},
			},
			"availability_zone": {
				Type:        "string",
				Description: "Availability zone",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"us-east-1a", "us-west-2c"},
			},
			"multi_az": {
				Type:        "bool",
				Description: "Whether Multi-AZ deployment is enabled",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},
			"security_groups": {
				Type:        "array",
				Description: "List of security group IDs",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{"sg-12345678"},
			},

			// Storage
			"allocated_storage": {
				Type:        "int",
				Description: "Allocated storage in GB",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"20", "100", "1000"},
			},
			"storage_type": {
				Type:        "string",
				Description: "Storage type",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"gp2", "gp3", "io1", "io2"},
				EnumValues:  []string{"standard", "gp2", "gp3", "io1", "io2"},
			},
			"storage_encrypted": {
				Type:        "bool",
				Description: "Whether storage is encrypted",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},
			"kms_key_id": {
				Type:        "string",
				Description: "KMS key ID for encryption",
				Operators:   []string{"eq", "ne", "exists", "not-exists"},
				Examples: []string{
					"arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
				},
			},
			"iops": {
				Type:        "int",
				Description: "Provisioned IOPS",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"1000", "3000", "10000"},
			},

			// Backup and maintenance
			"backup_retention_period": {
				Type:        "int",
				Description: "Backup retention period in days",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"0", "7", "35"},
			},
			"backup_window": {
				Type:        "string",
				Description: "Daily backup window",
				Operators:   []string{"eq", "ne", "contains", "exists", "not-exists"},
				Examples:    []string{"03:00-04:00", "23:00-00:00"},
			},
			"maintenance_window": {
				Type:        "string",
				Description: "Weekly maintenance window",
				Operators:   []string{"eq", "ne", "contains", "exists", "not-exists"},
				Examples:    []string{"sun:05:00-sun:06:00", "tue:03:00-tue:04:00"},
			},
			"auto_minor_version_upgrade": {
				Type:        "bool",
				Description: "Whether auto minor version upgrade is enabled",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},
			"deletion_protection": {
				Type:        "bool",
				Description: "Whether deletion protection is enabled",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},

			// Performance and monitoring (computed)
			"cpu_utilization": {
				Type:        "float",
				Description: "Average CPU utilization percentage",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"25.5", "80.0"},
				Computed:    true,
			},
			"database_connections": {
				Type:        "int",
				Description: "Number of database connections",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"10", "50", "200"},
				Computed:    true,
			},
			"freeable_memory": {
				Type:        "int",
				Description: "Free memory in bytes",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"536870912", "1073741824"},
				Computed:    true,
			},
			"free_storage_space": {
				Type:        "int",
				Description: "Free storage space in bytes",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"1073741824", "10737418240"},
				Computed:    true,
			},

			// Cost (computed)
			"monthly_cost": {
				Type:        "float",
				Description: "Estimated monthly cost in USD",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"15.50", "500.00"},
				Computed:    true,
			},

			// Age (computed)
			"age_days": {
				Type:        "int",
				Description: "Days since instance was created",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"30", "90", "365"},
				Computed:    true,
			},

			// Tags
			"tags": {
				Type:        "map",
				Description: "Instance tags",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{`{"Environment": "prod"}`, `{"Application": "webapp"}`},
			},
		},
		Relationships: []RelationshipDefinition{
			{
				Name:        "snapshots",
				Type:        "one-to-many",
				Target:      "rds-snapshot",
				Direction:   "outbound",
				Description: "Manual snapshots of this instance",
			},
			{
				Name:        "subnet_group",
				Type:        "many-to-one",
				Target:      "db-subnet-group",
				Direction:   "outbound",
				Description: "DB subnet group",
			},
			{
				Name:        "parameter_group",
				Type:        "many-to-one",
				Target:      "db-parameter-group",
				Direction:   "outbound",
				Description: "DB parameter group",
			},
			{
				Name:        "option_group",
				Type:        "many-to-one",
				Target:      "db-option-group",
				Direction:   "outbound",
				Description: "DB option group",
			},
			{
				Name:        "security_groups",
				Type:        "many-to-many",
				Target:      "security-group",
				Direction:   "outbound",
				Description: "Security groups",
			},
		},
		Actions: []string{
			"start",
			"stop",
			"reboot",
			"delete",
			"create-snapshot",
			"restore",
			"modify",
			"tag",
			"untag",
		},
		Tags: []string{"Environment", "Owner", "Project", "Application", "Database"},
	},

	"lambda": {
		Name:        "lambda",
		Service:     "Lambda",
		Description: "Lambda functions",
		Fields: map[string]FieldDefinition{
			// Basic function information
			"function_name": {
				Type:        "string",
				Description: "Function name",
				Operators: []string{
					"eq",
					"ne",
					"in",
					"not-in",
					"contains",
					"starts-with",
					"ends-with",
					"regex",
				},
				Examples: []string{"my-function", "data-processor"},
				Required: true,
			},
			"function_arn": {
				Type:        "string",
				Description: "Function ARN",
				Operators:   []string{"eq", "ne", "contains", "starts-with", "ends-with"},
				Examples:    []string{"arn:aws:lambda:us-east-1:123456789012:function:my-function"},
				Required:    true,
			},
			"runtime": {
				Type:        "string",
				Description: "Function runtime",
				Operators:   []string{"eq", "ne", "in", "not-in", "starts-with"},
				Examples:    []string{"python3.9", "nodejs18.x", "java11"},
				EnumValues: []string{
					"nodejs18.x",
					"nodejs16.x",
					"nodejs14.x",
					"python3.11",
					"python3.10",
					"python3.9",
					"python3.8",
					"java17",
					"java11",
					"java8",
					"dotnet6",
					"go1.x",
					"ruby2.7",
					"provided.al2",
				},
				Required: true,
			},
			"handler": {
				Type:        "string",
				Description: "Function handler",
				Operators:   []string{"eq", "ne", "contains", "starts-with"},
				Examples:    []string{"index.handler", "lambda_function.lambda_handler"},
			},
			"description": {
				Type:        "string",
				Description: "Function description",
				Operators: []string{
					"contains",
					"not-contains",
					"starts-with",
					"ends-with",
					"empty",
					"not-empty",
					"regex",
				},
				Examples: []string{"Processes user data", "API Gateway backend"},
			},
			"timeout": {
				Type:        "int",
				Description: "Function timeout in seconds",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"3", "30", "900"},
			},
			"memory_size": {
				Type:        "int",
				Description: "Memory size in MB",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"128", "512", "3008"},
			},
			"ephemeral_storage": {
				Type:        "int",
				Description: "Ephemeral storage in MB",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"512", "1024", "10240"},
			},

			// Code and deployment
			"code_size": {
				Type:        "int",
				Description: "Code size in bytes",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"1048576", "52428800"},
			},
			"code_sha256": {
				Type:        "string",
				Description: "SHA256 hash of the code",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"abc123..."},
			},
			"last_modified": {
				Type:        "time",
				Description: "Last modification time",
				Operators:   []string{"gt", "gte", "lt", "lte", "between", "age-gt", "age-lt"},
				Examples:    []string{"2024-01-15T10:30:00Z"},
			},
			"version": {
				Type:        "string",
				Description: "Function version",
				Operators:   []string{"eq", "ne", "starts-with"},
				Examples:    []string{"$LATEST", "1", "2"},
			},
			"package_type": {
				Type:        "string",
				Description: "Package type",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"Zip", "Image"},
				EnumValues:  []string{"Zip", "Image"},
			},

			// Environment and configuration
			"environment_variables": {
				Type:        "map",
				Description: "Environment variables",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{`{"NODE_ENV": "production"}`, `{"DEBUG": "true"}`},
			},
			"env_var_count": {
				Type:        "int",
				Description: "Number of environment variables",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte"},
				Examples:    []string{"0", "5", "20"},
				Computed:    true,
			},

			// Network and security
			"vpc_config": {
				Type:        "bool",
				Description: "Whether function is in a VPC",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
				Computed:    true,
			},
			"vpc_id": {
				Type:        "string",
				Description: "VPC ID (if in VPC)",
				Operators:   []string{"eq", "ne", "exists", "not-exists"},
				Examples:    []string{"vpc-12345678"},
			},
			"subnet_ids": {
				Type:        "array",
				Description: "Subnet IDs (if in VPC)",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{"subnet-12345678"},
			},
			"security_group_ids": {
				Type:        "array",
				Description: "Security group IDs (if in VPC)",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{"sg-12345678"},
			},
			"role": {
				Type:        "string",
				Description: "IAM role ARN",
				Operators:   []string{"eq", "ne", "contains", "starts-with"},
				Examples:    []string{"arn:aws:iam::123456789012:role/lambda-execution-role"},
			},

			// Execution and performance (computed)
			"invocations_24h": {
				Type:        "int",
				Description: "Invocations in last 24 hours",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"0", "100", "10000"},
				Computed:    true,
			},
			"errors_24h": {
				Type:        "int",
				Description: "Errors in last 24 hours",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"0", "5", "50"},
				Computed:    true,
			},
			"duration_avg": {
				Type:        "float",
				Description: "Average duration in milliseconds",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"100.5", "5000.0"},
				Computed:    true,
			},
			"throttles_24h": {
				Type:        "int",
				Description: "Throttles in last 24 hours",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte"},
				Examples:    []string{"0", "10"},
				Computed:    true,
			},
			"concurrent_executions": {
				Type:        "int",
				Description: "Current concurrent executions",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"0", "10", "100"},
				Computed:    true,
			},
			"error_rate": {
				Type:        "float",
				Description: "Error rate percentage",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"0.0", "5.0", "25.0"},
				Computed:    true,
			},

			// Cost and usage (computed)
			"monthly_cost": {
				Type:        "float",
				Description: "Estimated monthly cost in USD",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"0.00", "50.00"},
				Computed:    true,
			},
			"gb_seconds_24h": {
				Type:        "float",
				Description: "GB-seconds consumed in last 24 hours",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"100.0", "10000.0"},
				Computed:    true,
			},

			// Age and lifecycle (computed)
			"age_days": {
				Type:        "int",
				Description: "Days since function was created",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"1", "30", "365"},
				Computed:    true,
			},
			"last_invoked": {
				Type:        "time",
				Description: "Time of last invocation",
				Operators: []string{
					"gt",
					"gte",
					"lt",
					"lte",
					"age-gt",
					"age-lt",
					"exists",
					"not-exists",
				},
				Examples: []string{"2024-01-15T10:30:00Z"},
				Computed: true,
			},
			"days_since_last_invocation": {
				Type:        "int",
				Description: "Days since last invocation",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"0", "7", "30"},
				Computed:    true,
			},

			// Tags
			"tags": {
				Type:        "map",
				Description: "Function tags",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{`{"Environment": "prod"}`, `{"Team": "backend"}`},
			},
		},
		Relationships: []RelationshipDefinition{
			{
				Name:        "triggers",
				Type:        "one-to-many",
				Target:      "event-source-mapping",
				Direction:   "inbound",
				Description: "Event sources that trigger this function",
			},
			{
				Name:        "role",
				Type:        "many-to-one",
				Target:      "iam-role",
				Direction:   "outbound",
				Description: "IAM execution role",
			},
			{
				Name:        "layers",
				Type:        "many-to-many",
				Target:      "lambda-layer",
				Direction:   "outbound",
				Description: "Lambda layers used by this function",
			},
			{
				Name:        "aliases",
				Type:        "one-to-many",
				Target:      "lambda-alias",
				Direction:   "outbound",
				Description: "Function aliases",
			},
			{
				Name:        "destinations",
				Type:        "one-to-many",
				Target:      "lambda-destination",
				Direction:   "outbound",
				Description: "Success/failure destinations",
			},
		},
		Actions: []string{
			"invoke",
			"update-code",
			"update-configuration",
			"delete",
			"create-alias",
			"publish-version",
			"tag",
			"untag",
		},
		Tags: []string{"Environment", "Owner", "Project", "Team", "Application"},
	},

	"ebs": {
		Name:        "ebs",
		Service:     "EC2",
		Description: "Elastic Block Store volumes",
		Fields: map[string]FieldDefinition{
			// Basic volume information
			"volume_id": {
				Type:        "string",
				Description: "Volume identifier",
				Operators:   []string{"eq", "ne", "in", "not-in", "starts-with", "ends-with"},
				Examples:    []string{"vol-1234567890abcdef0"},
				Required:    true,
			},
			"volume_type": {
				Type:        "string",
				Description: "Volume type",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"gp2", "gp3", "io1", "io2", "st1", "sc1"},
				EnumValues:  []string{"standard", "gp2", "gp3", "io1", "io2", "st1", "sc1"},
				Required:    true,
			},
			"size": {
				Type:        "int",
				Description: "Volume size in GB",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"8", "100", "1000"},
				Required:    true,
			},
			"state": {
				Type:        "string",
				Description: "Volume state",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"available", "in-use", "deleting"},
				EnumValues: []string{
					"creating",
					"available",
					"in-use",
					"deleting",
					"deleted",
					"error",
				},
				Required: true,
			},
			"create_time": {
				Type:        "time",
				Description: "Volume creation time",
				Operators:   []string{"gt", "gte", "lt", "lte", "between", "age-gt", "age-lt"},
				Examples:    []string{"2024-01-15T10:30:00Z"},
				Required:    true,
			},
			"availability_zone": {
				Type:        "string",
				Description: "Availability zone",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"us-east-1a", "us-west-2b"},
				Required:    true,
			},

			// Encryption
			"encrypted": {
				Type:        "bool",
				Description: "Whether volume is encrypted",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},
			"kms_key_id": {
				Type:        "string",
				Description: "KMS key ID for encryption",
				Operators:   []string{"eq", "ne", "exists", "not-exists"},
				Examples: []string{
					"arn:aws:kms:us-east-1:123456789012:key/12345678-1234-1234-1234-123456789012",
				},
			},

			// Performance
			"iops": {
				Type:        "int",
				Description: "Provisioned IOPS",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"100", "3000", "16000"},
			},
			"throughput": {
				Type:        "int",
				Description: "Throughput in MB/s (gp3 only)",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"125", "250", "1000"},
			},
			"multi_attach_enabled": {
				Type:        "bool",
				Description: "Whether multi-attach is enabled",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},

			// Attachment information
			"attachment_state": {
				Type:        "string",
				Description: "Attachment state",
				Operators:   []string{"eq", "ne", "in", "not-in"},
				Examples:    []string{"attached", "detached", "attaching", "detaching"},
				EnumValues:  []string{"attaching", "attached", "detaching", "detached"},
			},
			"instance_id": {
				Type:        "string",
				Description: "Attached instance ID",
				Operators:   []string{"eq", "ne", "exists", "not-exists"},
				Examples:    []string{"i-1234567890abcdef0"},
			},
			"device": {
				Type:        "string",
				Description: "Device name on instance",
				Operators:   []string{"eq", "ne", "starts-with"},
				Examples:    []string{"/dev/sdf", "/dev/xvdf"},
			},
			"delete_on_termination": {
				Type:        "bool",
				Description: "Whether volume is deleted on instance termination",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
			},
			"attach_time": {
				Type:        "time",
				Description: "When volume was attached",
				Operators: []string{
					"gt",
					"gte",
					"lt",
					"lte",
					"age-gt",
					"age-lt",
					"exists",
					"not-exists",
				},
				Examples: []string{"2024-01-15T11:00:00Z"},
			},

			// Snapshot information
			"snapshot_id": {
				Type:        "string",
				Description: "Source snapshot ID",
				Operators:   []string{"eq", "ne", "exists", "not-exists"},
				Examples:    []string{"snap-1234567890abcdef0"},
			},

			// Performance metrics (computed)
			"read_ops_24h": {
				Type:        "int",
				Description: "Read operations in last 24 hours",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"1000", "50000"},
				Computed:    true,
			},
			"write_ops_24h": {
				Type:        "int",
				Description: "Write operations in last 24 hours",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"500", "25000"},
				Computed:    true,
			},
			"read_bytes_24h": {
				Type:        "int",
				Description: "Bytes read in last 24 hours",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"1048576", "1073741824"},
				Computed:    true,
			},
			"write_bytes_24h": {
				Type:        "int",
				Description: "Bytes written in last 24 hours",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"524288", "536870912"},
				Computed:    true,
			},
			"queue_depth_avg": {
				Type:        "float",
				Description: "Average queue depth",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"1.0", "10.0"},
				Computed:    true,
			},
			"utilization_percentage": {
				Type:        "float",
				Description: "Volume utilization percentage",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"5.0", "95.0"},
				Computed:    true,
			},

			// Cost (computed)
			"monthly_cost": {
				Type:        "float",
				Description: "Monthly cost in USD",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"1.00", "100.00"},
				Computed:    true,
			},

			// Age and lifecycle (computed)
			"age_days": {
				Type:        "int",
				Description: "Days since volume was created",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"1", "30", "365"},
				Computed:    true,
			},
			"days_since_attachment": {
				Type:        "int",
				Description: "Days since last attachment",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"0", "7", "30"},
				Computed:    true,
			},
			"unused_days": {
				Type:        "int",
				Description: "Days since volume was detached (available volumes only)",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"1", "7", "30"},
				Computed:    true,
			},

			// Tags
			"tags": {
				Type:        "map",
				Description: "Volume tags",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{`{"Environment": "prod"}`, `{"Backup": "daily"}`},
			},
		},
		Relationships: []RelationshipDefinition{
			{
				Name:        "instance",
				Type:        "many-to-one",
				Target:      "ec2",
				Direction:   "outbound",
				Description: "Instance this volume is attached to",
			},
			{
				Name:        "snapshots",
				Type:        "one-to-many",
				Target:      "snapshot",
				Direction:   "outbound",
				Description: "Snapshots created from this volume",
			},
			{
				Name:        "source_snapshot",
				Type:        "many-to-one",
				Target:      "snapshot",
				Direction:   "outbound",
				Description: "Snapshot this volume was created from",
			},
		},
		Actions: []string{
			"attach",
			"detach",
			"create-snapshot",
			"delete",
			"modify",
			"encrypt",
			"tag",
			"untag",
		},
		Tags: []string{"Name", "Environment", "Owner", "Project", "Backup"},
	},

	"iam-role": {
		Name:        "iam-role",
		Service:     "IAM",
		Description: "IAM roles",
		Fields: map[string]FieldDefinition{
			// Basic role information
			"role_name": {
				Type:        "string",
				Description: "Role name",
				Operators: []string{
					"eq",
					"ne",
					"in",
					"not-in",
					"contains",
					"starts-with",
					"ends-with",
					"regex",
				},
				Examples: []string{"ec2-role", "lambda-execution-role"},
				Required: true,
			},
			"role_id": {
				Type:        "string",
				Description: "Role ID",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"AROA123456789EXAMPLE"},
				Required:    true,
			},
			"arn": {
				Type:        "string",
				Description: "Role ARN",
				Operators:   []string{"eq", "ne", "contains", "starts-with", "ends-with"},
				Examples:    []string{"arn:aws:iam::123456789012:role/MyRole"},
				Required:    true,
			},
			"path": {
				Type:        "string",
				Description: "Role path",
				Operators:   []string{"eq", "ne", "starts-with", "contains"},
				Examples:    []string{"/", "/service/", "/application/"},
			},
			"description": {
				Type:        "string",
				Description: "Role description",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty", "regex"},
				Examples:    []string{"Role for EC2 instances", "Lambda execution role"},
			},
			"create_date": {
				Type:        "time",
				Description: "When the role was created",
				Operators:   []string{"gt", "gte", "lt", "lte", "between", "age-gt", "age-lt"},
				Examples:    []string{"2024-01-01T12:00:00Z"},
				Required:    true,
			},
			"max_session_duration": {
				Type:        "int",
				Description: "Maximum session duration in seconds",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"3600", "43200"},
			},

			// Trust policy
			"assume_role_policy_document": {
				Type:        "string",
				Description: "Trust policy document (JSON)",
				Operators:   []string{"contains", "not-contains", "regex"},
				Examples:    []string{"ec2.amazonaws.com", "lambda.amazonaws.com"},
			},
			"trusted_services": {
				Type:        "array",
				Description: "AWS services that can assume this role",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{"ec2.amazonaws.com", "lambda.amazonaws.com"},
				Computed:    true,
			},
			"trusted_accounts": {
				Type:        "array",
				Description: "AWS accounts that can assume this role",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{"123456789012", "987654321098"},
				Computed:    true,
			},
			"allows_cross_account": {
				Type:        "bool",
				Description: "Whether role allows cross-account access",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
				Computed:    true,
			},

			// Attached policies
			"attached_policies": {
				Type:        "array",
				Description: "List of attached managed policy ARNs",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{"arn:aws:iam::aws:policy/ReadOnlyAccess"},
			},
			"attached_policy_count": {
				Type:        "int",
				Description: "Number of attached managed policies",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte"},
				Examples:    []string{"0", "5", "10"},
				Computed:    true,
			},
			"inline_policies": {
				Type:        "array",
				Description: "List of inline policy names",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{"S3Access", "DynamoDBRead"},
			},
			"inline_policy_count": {
				Type:        "int",
				Description: "Number of inline policies",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte"},
				Examples:    []string{"0", "2", "5"},
				Computed:    true,
			},
			"has_admin_access": {
				Type:        "bool",
				Description: "Whether role has administrator access",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
				Computed:    true,
			},

			// Usage tracking (computed)
			"last_used": {
				Type:        "time",
				Description: "When the role was last used",
				Operators: []string{
					"gt",
					"gte",
					"lt",
					"lte",
					"age-gt",
					"age-lt",
					"exists",
					"not-exists",
				},
				Examples: []string{"2024-01-15T10:30:00Z"},
				Computed: true,
			},
			"days_since_last_used": {
				Type:        "int",
				Description: "Days since role was last used",
				Operators:   []string{"gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"0", "30", "90"},
				Computed:    true,
			},
			"never_used": {
				Type:        "bool",
				Description: "Whether role has never been used",
				Operators:   []string{"eq", "ne"},
				Examples:    []string{"true", "false"},
				Computed:    true,
			},

			// Associated resources (computed)
			"instance_profile_count": {
				Type:        "int",
				Description: "Number of associated instance profiles",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte"},
				Examples:    []string{"0", "1", "5"},
				Computed:    true,
			},
			"lambda_function_count": {
				Type:        "int",
				Description: "Number of Lambda functions using this role",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte"},
				Examples:    []string{"0", "5", "20"},
				Computed:    true,
			},

			// Age (computed)
			"age_days": {
				Type:        "int",
				Description: "Days since role was created",
				Operators:   []string{"eq", "ne", "gt", "gte", "lt", "lte", "between"},
				Examples:    []string{"1", "90", "365"},
				Computed:    true,
			},

			// Tags
			"tags": {
				Type:        "map",
				Description: "Role tags",
				Operators:   []string{"contains", "not-contains", "empty", "not-empty"},
				Examples:    []string{`{"Environment": "prod"}`, `{"Team": "security"}`},
			},
		},
		Relationships: []RelationshipDefinition{
			{
				Name:        "policies",
				Type:        "many-to-many",
				Target:      "iam-policy",
				Direction:   "outbound",
				Description: "Attached managed policies",
			},
			{
				Name:        "instance_profiles",
				Type:        "one-to-many",
				Target:      "instance-profile",
				Direction:   "outbound",
				Description: "Associated instance profiles",
			},
			{
				Name:        "lambda_functions",
				Type:        "one-to-many",
				Target:      "lambda",
				Direction:   "inbound",
				Description: "Lambda functions using this role",
			},
		},
		Actions: []string{
			"delete",
			"attach-policy",
			"detach-policy",
			"put-inline-policy",
			"delete-inline-policy",
			"tag",
			"untag",
		},
		Tags: []string{"Environment", "Owner", "Project", "Team", "Purpose"},
	},
}

// GetResourceDefinition returns the definition for a specific resource type
func GetResourceDefinition(resourceType string) (ResourceDefinition, bool) {
	def, exists := AWSResourceDefinitions[resourceType]
	return def, exists
}

// GetAllResourceTypes returns all supported resource types
func GetAllResourceTypes() []string {
	var types []string
	for resourceType := range AWSResourceDefinitions {
		types = append(types, resourceType)
	}
	return types
}

// GetResourceServices returns all AWS services covered
func GetResourceServices() map[string][]string {
	services := make(map[string][]string)
	for resourceType, def := range AWSResourceDefinitions {
		services[def.Service] = append(services[def.Service], resourceType)
	}
	return services
}

// GetFieldDefinition returns the definition for a specific field of a resource
func GetFieldDefinition(resourceType, fieldName string) (FieldDefinition, bool) {
	if def, exists := AWSResourceDefinitions[resourceType]; exists {
		if field, fieldExists := def.Fields[fieldName]; fieldExists {
			return field, true
		}
	}
	return FieldDefinition{}, false
}

// ValidateOperatorForField checks if an operator is valid for a given field
func ValidateOperatorForField(resourceType, fieldName, operator string) bool {
	if field, exists := GetFieldDefinition(resourceType, fieldName); exists {
		for _, validOp := range field.Operators {
			if validOp == operator {
				return true
			}
		}
	}
	return false
}

// GetComputedFields returns all computed fields for a resource type
func GetComputedFields(resourceType string) []string {
	var computedFields []string
	if def, exists := AWSResourceDefinitions[resourceType]; exists {
		for fieldName, field := range def.Fields {
			if field.Computed {
				computedFields = append(computedFields, fieldName)
			}
		}
	}
	return computedFields
}

// GetRequiredFields returns all required fields for a resource type
func GetRequiredFields(resourceType string) []string {
	var requiredFields []string
	if def, exists := AWSResourceDefinitions[resourceType]; exists {
		for fieldName, field := range def.Fields {
			if field.Required {
				requiredFields = append(requiredFields, fieldName)
			}
		}
	}
	return requiredFields
}
