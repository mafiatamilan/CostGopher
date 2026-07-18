# CostGopher

A low-cost AWS resource-creation and billing alert bot that sends notifications to Google Chat.

## How it works

CostGopher monitors AWS CloudTrail management events for resource-creation API calls that incur cost (EC2 `RunInstances`, RDS `CreateDBInstance`, S3 `CreateBucket`, etc.). When a matching event occurs, it posts a plain-English alert to a Google Chat space. It also sends a weekly bill summary and a mid-month cost forecast using AWS Cost Explorer.

## Architecture

```
CloudTrail → EventBridge → resource-alert-fn → Google Chat webhook
                            weekly-bill-fn    → Google Chat webhook  (weekly cron)
                            mid-month-forecast-fn → Google Chat     (15th of month cron)
```

All state (resource registry) is stored in SSM Parameter Store — no databases needed.

## Estimated monthly cost

| Service | Cost |
|---------|------|
| Lambda (1M free requests, 400K GB-s) | $0.00 |
| CloudTrail (management events) | $0.00 |
| EventBridge (rules for CloudTrail events) | $0.00 |
| Cost Explorer (~5 calls/month @ $0.01) | ~$0.05 |
| SSM Parameter Store (2 parameters) | $0.00 |
| CloudWatch Logs (14-day retention) | ~$0.05 |
| **Total** | **~$0.10/month** |

## Prerequisites

- Go 1.22+
- Terraform 1.0+
- AWS CLI configured with credentials
- A Google Chat space with an [Incoming Webhook](https://developers.google.com/chat/how-tos/webhooks)

## Setup

### 1. Set the Google Chat webhook URL

Create the SSM parameter (or let Terraform do it — see step 3):

```bash
aws ssm put-parameter \
  --name "/costgopher/webhook-url" \
  --type SecureString \
  --value "https://chat.googleapis.com/v1/spaces/.../messages?key=..."
```

### 2. Build

```bash
make build
```

This cross-compiles the three Lambda binaries for `linux/amd64` into `dist/`.

### 3. Deploy

```bash
cd terraform
terraform init
terraform apply -var="google_chat_webhook_url=https://chat.googleapis.com/v1/spaces/.../messages?key=..."
```

Or from the project root:

```bash
make deploy GOOGLE_CHAT_WEBHOOK_URL="https://chat.googleapis.com/v1/spaces/.../messages?key=..."
```

### 4. Verify

Check the Lambda logs in CloudWatch to verify each function runs. You can also manually invoke a Lambda to test:

```bash
aws lambda invoke --function-name costgopher-weekly-bill output.txt
```

## Customizing the curated event list

Add or remove events in `terraform/eventbridge.tf` under the `locals` block:

```hcl
locals {
  costed_events = [
    "RunInstances",
    "CreateDBInstance",
    ...
  ]
  costed_sources = [
    "aws.ec2",
    "aws.rds",
    ...
  ]
}
```

And add the corresponding entry in `internal/format/format.go`:

```go
var eventActions = map[string]string{
    "NewApiCall": "description of what it does",
}
```

## Message examples

### Resource creation alert
```
🟢 New AWS resource created
Who: john.doe (IAM user)
What: launched a new server — EC2 server (t3.micro) in us-east-1
Name: web-server-prod-01 (i-0abc123)
When: 2026-07-21 14:32 UTC
⚠️  This service has a cost — keep an eye on it.
```

### Weekly bill
```
📊 Weekly AWS Bill Update
Total this week: $4.32

EC2 — $2.10
  • web-server-prod-01 (i-0abc123)

RDS — $0.90
  • orders-db-primary
```

### Mid-month forecast
```
🔮 Mid-Month Forecast (as of the 15th)
Projected total for July: $9.80
So far spent: $4.60
```

## Project structure

```
├── cmd/
│   ├── resource-alert/main.go       # Event-driven Lambda for creation alerts
│   ├── weekly-bill/main.go           # Scheduled Lambda for weekly summary
│   └── mid-month-forecast/main.go    # Scheduled Lambda for mid-month forecast
├── internal/
│   ├── notify/notify.go              # Google Chat webhook client
│   ├── format/format.go              # CloudTrail event → plain English
│   ├── cost/cost.go                  # Cost Explorer API wrapper
│   └── registry/registry.go          # SSM Parameter-based resource registry
├── terraform/
│   ├── providers.tf                  # Terraform/AWS provider config
│   ├── variables.tf                  # Input variables
│   ├── iam.tf                        # IAM role + least-privilege policy
│   ├── lambda.tf                     # Lambda functions + log groups
│   ├── eventbridge.tf                # EventBridge rules + targets
│   ├── cloudtrail.tf                 # Optional CloudTrail trail
│   ├── ssm.tf                        # SSM Parameter Store entries
│   └── outputs.tf                    # Outputs
├── Makefile                          # Build & deploy automation
└── README.md
```

## CloudTrail

If your account does not already have CloudTrail enabled, set `create_cloudtrail = true` in `terraform/variables.tf` or pass it as `-var`:

```bash
terraform apply -var="create_cloudtrail=true"
```

This creates a trail that captures management events only (free). If you already have a trail capturing management events, you do not need this.

## Notes

- The resource registry uses SSM Parameter Store (not DynamoDB). The parameter stores a JSON array of resource records created upon each `resource-alert-fn` invocation.
- Weekly billing matches Cost Explorer service-level totals against the registry's resource names for display context.
- Lambda functions run without a VPC (no NAT gateway needed) — they only call AWS APIs and outbound HTTPS to the Google Chat webhook.
