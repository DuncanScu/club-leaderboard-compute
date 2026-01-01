# Leaderboard Computer Lambda

This Lambda function computes leaderboards for clubs and users on a scheduled basis (every 5 minutes via EventBridge).

## Architecture

- **Trigger**: EventBridge scheduled event (every 5 minutes)
- **Runtime**: Go 1.23 (Docker container)
- **Deployment**: GitHub Actions → ECR → Lambda
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
- Docker
- AWS CLI configured
- Access to the Aurora database

### Building

```bash
# Build Go binary locally
make build

# Build Docker image
make docker-build

# Run Docker container locally
make docker-run
```

### Testing Locally

```bash
# Run tests
make test

# Run with Docker
docker run --rm \
  -e DB_HOST=localhost \
  -e DB_PORT=5432 \
  -e DB_USER=postgres \
  -e DB_PASSWORD=your-password \
  -e DB_NAME=clubb \
  leaderboard-computer:latest
```

## Deployment

### Automatic Deployment (Recommended)

Deployment happens automatically via GitHub Actions when you push to the `main` branch:

1. Push your changes to `main`
2. GitHub Actions workflow builds Docker image
3. Pushes image to ECR with tags:
   - `latest`
   - Git commit SHA
4. Lambda automatically pulls the new `latest` image on next invocation

**Workflow file**: `.github/workflows/publish-ecr.yml`

### Manual Deployment

If you need to deploy manually:

```bash
# Build Docker image
docker build -t leaderboard-computer:latest .

# Tag for ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com
docker tag leaderboard-computer:latest <account-id>.dkr.ecr.us-east-1.amazonaws.com/leaderboard-computer:latest

# Push to ECR
docker push <account-id>.dkr.ecr.us-east-1.amazonaws.com/leaderboard-computer:latest

# Update Lambda (optional - Lambda pulls automatically)
aws lambda update-function-code \
  --function-name ClubbCdkStack-LeaderboardComputer \
  --image-uri <account-id>.dkr.ecr.us-east-1.amazonaws.com/leaderboard-computer:latest
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
- **ECR Repository**: `leaderboard-computer`

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
- Verify IAM role has ECR push permissions
- Ensure ECR repository exists and Lambda has pull permissions

### Image not updating
- Lambda may cache the image
- Force update with `aws lambda update-function-code`
- Check ECR for latest image tag

## Dependencies

See `go.mod` for full list. Key dependencies:
- `github.com/aws/aws-lambda-go` - Lambda runtime
- `gorm.io/gorm` - ORM for database access
- `gorm.io/driver/postgres` - PostgreSQL driver
