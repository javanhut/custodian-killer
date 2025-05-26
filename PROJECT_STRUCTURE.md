custodian-killer/
├── main.go                    # Main entry point with CLI setup and ASCII art
├── commands.go                # All the cobra command definitions  
├── policy.go                  # Policy engine, data structures, and core logic
├── wizard.go                  # Interactive policy creation wizard
├── aws/
│   ├── client.go             # AWS SDK client setup and configuration
│   ├── ec2.go                # EC2-specific operations
│   ├── s3.go                 # S3-specific operations  
│   ├── rds.go                # RDS-specific operations
│   └── iam.go                # IAM-specific operations
├── storage/
│   ├── file.go               # File-based policy storage
│   └── memory.go             # In-memory storage for testing
├── reports/
│   ├── html.go               # HTML report generation
│   ├── json.go               # JSON report output
│   └── csv.go                # CSV report output
├── templates/
│   └── policies.go           # Built-in policy templates
├── utils/
│   ├── config.go             # Configuration management
│   ├── colors.go             # Terminal colors and formatting
│   └── helpers.go            # Common utility functions
├── web/                      # Future web GUI
│   ├── server.go
│   └── static/
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums (auto-generated)
├── README.md                 # Epic README with usage examples
└── Makefile                  # Build automation
