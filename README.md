# Leaderboard Computer Lambda

This Lambda function computes leaderboards for clubs and users on a scheduled basis (every 5 minutes via EventBridge).

## Architecture

- **Trigger**: EventBridge scheduled event (every 5 minutes)
- **Runtime**: Go 1.23 (custom runtime)
- **Deployment**: GitHub Actions (automatic on push to main)
- **Database**: PostgreSQL (Aurora) via VPC

## What It Does

The leaderboard computer calculates and stores pre-computed rankings for:

1. **Club Leaderboards** (global and local/city-based)
   - Weekly rankings
   - Monthly rankings
   - Annual rankings

2. **User Contributor Leaderboards** (per club)
   - Weekly rankings
   - Monthly rankings
   - Annual rankings

Results are stored in snapshot tables:
- `club_leaderboard_snapshots`
- `user_club_leaderboard_snapshots`

## Local Development

### Prerequisites
- Go 1.23 or higher
- AWS CLI configured
- Access to the Aurora database

### Building

```bash
# Build for local testing
go build -o leaderboard-computer main.go

# Build for Lambda deployment
make build

# Create deployment package
make zip
```

### Testing Locally

```bash
# Run tests
make test

# Set up environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your-password
export DB_NAME=clubb

# Run locally
go run main.go
```

## Deployment

### Automatic Deployment (Recommended)

Deployment happens automatically via GitHub Actions when you push to the `main` branch:

1. Push your changes to `main`
2. GitHub Actions workflow builds the Go binary
3. Creates a ZIP package
4. Deploys to AWS Lambda
5. Publishes a new Lambda version

### Manual Deployment

If you need to deploy manually:

```bash
# Build and create ZIP
make zip

# Deploy using AWS CLI
aws lambda update-function-code \
  --function-name ClubbCdkStack-LeaderboardComputer \
  --zip-file fileb://bootstrap.zip \
  --region us-east-1
```

## Configuration

Environment variables (set by CDK):
- `DB_HOST` - Aurora cluster endpoint
- `DB_PORT` - Database port (5432)
- `DB_NAME` - Database name (clubb)
- `DB_SECRET_ARN` - ARN of Secrets Manager secret containing DB credentials

## Monitoring

- **CloudWatch Logs**: `/aws/lambda/ClubbCdkStack-LeaderboardComputer`
- **Metrics**: Lambda metrics in CloudWatch
- **Schedule**: EventBridge rule `LeaderboardComputeSchedule`

## Troubleshooting

### Lambda timeout errors
- Current timeout is 5 minutes
- Check CloudWatch Logs for performance issues
- Consider optimizing database queries

### Database connection issues
- Verify VPC security group allows Lambda access
- Check database secret is accessible
- Verify RDS cluster is available

### Failed deployments
- Check GitHub Actions logs
- Verify IAM role has Lambda update permissions
- Ensure function name matches in workflow

## Dependencies

See `go.mod` for full list. Key dependencies:
- `github.com/aws/aws-lambda-go` - Lambda runtime
- `gorm.io/gorm` - ORM for database access
- `gorm.io/driver/postgres` - PostgreSQL driver
