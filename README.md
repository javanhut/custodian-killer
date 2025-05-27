# Custodian Killer

```
 ██████╗██╗   ██╗███████╗████████╗ ██████╗ ██████╗ ██╗ █████╗ ███╗   ██╗    
██╔════╝██║   ██║██╔════╝╚══██╔══╝██╔═══██╗██╔══██╗██║██╔══██╗████╗  ██║    
██║     ██║   ██║███████╗   ██║   ██║   ██║██║  ██║██║███████║██╔██╗ ██║    
██║     ██║   ██║╚════██║   ██║   ██║   ██║██║  ██║██║██╔══██║██║╚██╗██║    
╚██████╗╚██████╔╝███████║   ██║   ╚██████╔╝██████╔╝██║██║  ██║██║ ╚████║    
 ╚═════╝ ╚═════╝ ╚══════╝   ╚═╝    ╚═════╝ ╚═════╝ ╚═╝╚═╝  ╚═╝╚═╝  ╚═══╝    
                                                                             
██╗  ██╗██╗██╗     ██╗     ███████╗██████╗                                  
██║ ██╔╝██║██║     ██║     ██╔════╝██╔══██╗                                 
█████╔╝ ██║██║     ██║     █████╗  ██████╔╝                                 
██╔═██╗ ██║██║     ██║     ██╔══╝  ██╔══██╗                                 
██║  ██╗██║███████╗███████╗███████╗██║  ██║                                 
╚═╝  ╚═╝╚═╝╚══════╝╚══════╝╚══════╝╚═╝  ╚═╝                                 
```

**Making AWS compliance fun again!** 🔥

The AWS policy management tool that actually works. Making Cloud Custodian weep since 2025.

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/your-username/custodian-killer)
[![AWS](https://img.shields.io/badge/AWS-Ready-orange.svg)](https://aws.amazon.com)

## 🚀 What is Custodian Killer?

Custodian Killer is a powerful, user-friendly CLI tool that helps you manage AWS resources with intelligent policies. Unlike other tools that make you cry, this one actually makes AWS compliance **enjoyable**.

### 🎯 Key Features

- **🤖 AI-Powered Policy Creation** - Describe what you want in plain English, get perfect policies
- **📊 Beautiful Reports** - HTML, JSON, and CSV reports that actually look good
- **🧪 Safe Dry-Run Mode** - Test everything before making changes
- **⚡ Lightning Fast** - Scan thousands of resources in seconds
- **🎨 Interactive CLI** - No more YAML hell, just intuitive wizards
- **💰 Cost Optimization** - Find and eliminate waste automatically
- **🔒 Security Enforcement** - Lock down public resources instantly
- **📋 Compliance Monitoring** - Stay compliant without the headache

## 🏗️ Architecture

```
custodian-killer/
├── aws/           # AWS SDK integrations & resource management
├── reports/       # HTML, JSON, CSV report generation
├── scanner/       # Policy scanning & resource discovery
├── storage/       # Policy storage & versioning
├── templates/     # Pre-built policy templates
└── main.go        # CLI interface & interactive wizard
```

## 🛠️ Installation

### Quick Install (Recommended)

```bash
# Clone the repo
git clone https://github.com/your-username/custodian-killer.git
cd custodian-killer

# Build and install
make install

# Or build locally
make build
```

### Prerequisites

- **Go 1.21+** (because we use the latest and greatest)
- **AWS CLI configured** (with your credentials)
- **Proper AWS permissions** (see [Permissions](#-aws-permissions) section)

## ⚡ Quick Start

### 1. Interactive Mode (Easiest)

```bash
# Start the interactive wizard
custodian-killer

# Or explicitly
custodian-killer interactive
```

### 2. Create Your First Policy

```bash
# Use the AI to create a policy
custodian-killer policy create

# Or use a template
custodian-killer policy create --template unused-ec2-killer
```

### 3. Test Before You Execute

```bash
# Scan to see what would happen (safe dry-run)
custodian-killer scan

# Execute for real (with confirmation)
custodian-killer execute
```

### 4. Generate Beautiful Reports

```bash
# Generate compliance report
custodian-killer report compliance --output html

# Generate cost analysis
custodian-killer report cost --output csv

# Generate everything
custodian-killer report inventory
```

## 🎮 Usage Examples

### AI-Powered Policy Creation

```bash
# Just describe what you want!
$ custodian-killer policy create
🤖 Describe what you want this policy to do: 
> Find EC2 instances that are wasting money and stop them

🧠 AI Analysis: Detected 'cost-optimization' intent from EC2 keywords
💡 AI Confidence: 89%

📋 Generated Policy: cost-optimizer-ec2-1201
🔍 Filters: instance-state=running, cpu-utilization<5%, running-days>=7
⚡ Actions: stop (dry-run enabled)
```

### Template-Based Creation

```bash
# Use pre-built templates
$ custodian-killer policy create --template
1. 🎯 Unused EC2 Instance Killer (high impact)
2. 🔒 Public S3 Bucket Security Locker (high impact)  
3. 🏷️ Untagged Resources Auto-Tagger (medium impact)
4. 🗄️ RDS Backup Policy Enforcer (high impact)
5. 🧹 Old EBS Snapshots Cleaner (medium impact)

Choose template: 2

✅ Selected: Public S3 Bucket Security Locker
📝 Configure variables...
```

### Policy Management

```bash
# List all policies
custodian-killer policy list

# Edit existing policy
custodian-killer policy edit my-policy-name

# Export/Import policies
custodian-killer policy export my-policy policy.json
custodian-killer policy import policy.json

# Delete policy
custodian-killer policy delete old-policy
```

### Scanning & Execution

```bash
# Scan specific policy
custodian-killer scan --policy unused-ec2-killer

# Scan all policies
custodian-killer scan

# Execute with confirmation
custodian-killer execute --policy cost-optimizer

# Force execution (skip confirmation)
custodian-killer execute --force

# Dry-run mode (same as scan)
custodian-killer execute --dry-run
```

### Report Generation

```bash
# HTML compliance report (opens in browser)
custodian-killer report compliance --output html

# CSV cost analysis (Excel-ready)
custodian-killer report cost --output csv --file monthly-costs.csv

# JSON data export
custodian-killer report compliance --output json

# Full executive summary (all formats)
custodian-killer report inventory
```

## 📊 Sample Reports

### Compliance Report (HTML)
![Compliance Report](docs/images/compliance-report.png)

**Features:**
- 🎯 Executive summary with key metrics
- 📈 Security score trending
- 🔍 Detailed findings by resource type
- 💰 Cost impact analysis
- 📋 Actionable recommendations

### Cost Analysis (CSV)
Perfect for Excel/Google Sheets analysis:
- Resource-level cost breakdown
- Potential savings identification
- Utilization metrics
- Risk assessment
- Recommended actions

## 🎯 Policy Templates

### Built-in Templates

| Template | Description | Impact | Difficulty |
|----------|-------------|--------|------------|
| 🎯 **Unused EC2 Killer** | Stop/terminate low-utilization instances | High | Beginner |
| 🔒 **S3 Security Locker** | Secure public S3 buckets | High | Intermediate |
| 🏷️ **Auto-Tagger** | Tag untagged resources | Medium | Beginner |
| 🗄️ **RDS Backup Enforcer** | Ensure proper backup policies | High | Intermediate |
| 🧹 **Snapshot Cleaner** | Clean up old EBS snapshots | Medium | Beginner |

### Custom Templates

Create your own templates and share them:

```bash
# Create custom template
custodian-killer template create --name "My Custom Policy"

# Share template
custodian-killer template export my-template template.json
```

## 🔧 Configuration

### AWS Credentials

Custodian Killer supports multiple authentication methods:

```bash
# Environment variables
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
export AWS_REGION="us-east-1"

# AWS Profile
export AWS_PROFILE="your-profile"

# Or use AWS SSO, instance roles, etc.
```

### Configuration File

Create `~/.custodian-killer/config.yaml`:

```yaml
aws:
  region: us-east-1
  profile: default
  timeout: 30s

policies:
  storage_path: ~/.custodian-killer/policies
  backup_enabled: true
  version_history: 10

reports:
  output_directory: ./reports
  default_format: html
  include_costs: true

execution:
  dry_run_default: true
  confirm_destructive: true
  max_concurrency: 5
```

## 🔐 AWS Permissions

### Minimum Required Permissions

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:Describe*",
        "s3:ListAllMyBuckets",
        "s3:GetBucketLocation",
        "s3:GetBucketTagging",
        "rds:Describe*",
        "lambda:List*",
        "iam:List*"
      ],
      "Resource": "*"
    }
  ]
}
```

### Full Permissions (For Execution)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:*",
        "s3:*",
        "rds:*",
        "lambda:*",
        "iam:*",
        "cloudwatch:GetMetricStatistics"
      ],
      "Resource": "*"
    }
  ]
}
```

## 🧪 Testing & Development

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package tests
go test -v ./aws/...

# Run integration tests (requires AWS credentials)
make test-integration
```

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Build Docker image
make docker-build
```

### Development Setup

```bash
# Clone and setup
git clone https://github.com/your-username/custodian-killer.git
cd custodian-killer

# Install dependencies
go mod download

# Run in development mode
go run main.go

# Or build and run
make build
./bin/custodian-killer
```

## 📚 Advanced Usage

### Batch Operations

```bash
# Process multiple policies
custodian-killer scan --policy "cost-*" --parallel

# Execute multiple policies with different settings
custodian-killer execute --config batch-config.yaml
```

### API Integration

```bash
# Generate JSON reports for API consumption
custodian-killer report compliance --output json --file /tmp/compliance.json

# Use in CI/CD pipelines
custodian-killer scan --policy security-check --exit-code
```

### Custom Filters & Actions

Create complex policies with custom logic:

```yaml
# Example: Advanced EC2 policy
name: "advanced-ec2-optimization"
resource_type: "ec2"
filters:
  - type: "instance-age"
    value: 30
    op: "gt"
  - type: "cpu-utilization-avg"
    value: 5
    op: "lt"
    days: 7
  - type: "tag"
    key: "Environment"
    value: "production"
    op: "ne"
actions:
  - type: "stop"
    dry_run: false
  - type: "tag"
    settings:
      OptimizedBy: "custodian-killer"
      OptimizedDate: "{{ .current_date }}"
```

## 🚨 Safety Features

### Dry-Run Mode
All operations default to dry-run mode. Real changes require explicit confirmation.

### Policy Versioning
All policy changes are versioned and can be rolled back.

### Backup & Recovery
Automatic backups before destructive operations.

### Confirmation Prompts
Interactive prompts for dangerous operations.

## 🤝 Contributing

We love contributions! Here's how to get started:

### Quick Contribution Guide

1. **Fork the repo** 🍴
2. **Create a feature branch** 🌿
   ```bash
   git checkout -b feature/amazing-feature
   ```
3. **Make your changes** ✨
4. **Add tests** 🧪
   ```bash
   make test
   ```
5. **Submit a PR** 🚀

### Development Guidelines

- **Code Style**: Follow Go conventions + our `.golangci.yml`
- **Testing**: Add tests for new features
- **Documentation**: Update README and docs
- **Emojis**: We love emojis! Use them liberally 🦍

### Areas We Need Help

- 🔌 **More AWS Services** (EKS, CloudFormation, etc.)
- 🎨 **UI Improvements** (Better reports, dashboard)
- 🧪 **Testing** (Integration tests, edge cases)
- 📚 **Documentation** (Tutorials, examples)
- 🌍 **Internationalization** (Multi-language support)

## 🐛 Issues & Support

### Reporting Bugs

Found a bug? We want to squash it! 🐛

1. **Check existing issues** first
2. **Create detailed bug report** with:
   - Steps to reproduce
   - Expected vs actual behavior
   - System info (OS, Go version, AWS region)
   - Logs (use `--verbose` flag)

### Getting Help

- 📖 **Documentation**: Check this README first
- 💬 **Discussions**: Use GitHub Discussions for questions
- 🐛 **Issues**: Use GitHub Issues for bugs
- 📧 **Email**: TODO

## 📈 Roadmap

### 🎯 Version 2.0 (Coming Soon)

- **🎛️ Web Dashboard** - Beautiful web UI for policy management
- **🔔 Real-time Notifications** - Slack, Teams, email integration
- **📊 Advanced Analytics** - Trend analysis, cost forecasting
- **🤖 Machine Learning** - Smarter recommendations
- **🔗 CI/CD Integration** - GitHub Actions, Jenkins plugins

### 🚀 Version 3.0 (Future)

- **☁️ Multi-Cloud Support** - Azure, GCP support
- **📱 Mobile App** - iOS/Android compliance monitoring
- **🧠 Advanced AI** - Natural language policy queries
- **🏢 Enterprise Features** - RBAC, audit trails, compliance frameworks

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👏 Acknowledgments

- **AWS SDK Team** - For the excellent Go SDK
- **Cloud Custodian** - For inspiration (and for being so hard to use that we built this)
- **The Go Community** - For building an amazing ecosystem
- **Our Contributors** - You make this project awesome!

## 🎉 Final Words

Custodian Killer isn't just a tool - it's a revolution in AWS management. We're making compliance fun, costs transparent, and AWS management actually enjoyable.

---

**Made with ☕ by developers who got tired of YAML**

*P.S. - Yes, we know the name is provocative. That's the point. Cloud Custodian made us do it.* 😈
