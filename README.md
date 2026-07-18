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

- **Go 1.22+** — for compiling Lambda binaries
- **Terraform 1.0+** — for infrastructure deployment
- **AWS CLI** — configured with credentials that have sufficient permissions (Admin or PowerUser equivalent)
- **A Google Chat space** with an [Incoming Webhook](https://developers.google.com/chat/how-tos/webhooks) URL

## Deployment guide

### Step 1: Clone the repo

```bash
git clone git@github.com:mafiatamilan/CostGopher.git
cd CostGopher
```

### Step 2: Get your Google Chat webhook URL

1. Go to your Google Chat space
2. Click the space name → **Manage webhooks**
3. Add an incoming webhook, give it a name (e.g. "CostGopher"), and copy the URL
4. The URL looks like:
   ```
   https://chat.googleapis.com/v1/spaces/AAAA.../messages?key=BBBB...&token=CCCC...
   ```

### Step 3: Build the Lambda binaries

```bash
make build
```

This cross-compiles all three Lambda handlers for `linux/amd64` (the Lambda `provided.al2023` runtime). The binaries are placed in `dist/`:

```
dist/
├── resource-alert/bootstrap
├── weekly-bill/bootstrap
└── mid-month-forecast/bootstrap
```

### Step 4: Deploy with Terraform

```bash
cd terraform
terraform init
terraform apply \
  -var="google_chat_webhook_url=https://chat.googleapis.com/v1/spaces/.../messages?key=..." \
  -var="aws_region=us-east-1"
```

You'll be prompted to confirm. Type `yes`.

**What Terraform creates:**
| Resource | Name | Purpose |
|----------|------|---------|
| IAM role | `costgopher-lambda-role` | Execution role for all three functions |
| IAM policy | `costgopher-lambda-policy` | Least-privilege: `ce:GetCost*`, `ssm:Get/PutParameter`, `logs:*` |
| Lambda | `costgopher-resource-alert` | Processes CloudTrail events, sends alert, writes to registry |
| Lambda | `costgopher-weekly-bill` | Runs weekly, queries Cost Explorer, formats + sends summary |
| Lambda | `costgopher-mid-month-forecast` | Runs on the 15th, forecasts monthly cost |
| EventBridge rule | `costgopher-resource-alert` | Matches `RunInstances`, `CreateDBInstance`, etc. |
| EventBridge rule | `costgopher-weekly-bill` | Cron: Monday 9am UTC |
| EventBridge rule | `costgopher-mid-month-forecast` | Cron: 15th of each month at 9am UTC |
| SSM param | `/costgopher/webhook-url` | Encrypted webhook URL (SecureString) |
| SSM param | `/costgopher/resources` | Resource registry JSON (String) |
| Log groups | `/aws/lambda/costgopher-*` | 14-day retention (configurable) |

You can also deploy from the project root with the Makefile:

```bash
make deploy GOOGLE_CHAT_WEBHOOK_URL="https://chat.googleapis.com/v1/spaces/.../messages?key=..."
```

### Step 5: Verify

**Check Lambda logs:**
```bash
# See recent invocations for the weekly bill function
aws logs filter-log-events \
  --log-group-name /aws/lambda/costgopher-weekly-bill \
  --limit 10
```

**Manually invoke a function (triggers a test run):**
```bash
aws lambda invoke \
  --function-name costgopher-weekly-bill \
  --invocation-type RequestResponse \
  output.txt && cat output.txt
```

If successful, you'll see a message in your Google Chat space within a few seconds.

### Step 6: Wait for CloudTrail events

The resource-alert function only fires when someone actually creates a resource. To test it, launch a small EC2 instance or create an S3 bucket, then check the logs:

```bash
aws logs filter-log-events \
  --log-group-name /aws/lambda/costgopher-resource-alert \
  --limit 10
```

You should see the alert message in Google Chat.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| Lambda timeout | Cost Explorer API slow | Increase `timeout` in `terraform/lambda.tf` for weekly/mid-month functions |
| "AccessDenied" in logs | IAM policy missing permissions | Verify `ce:*` and `ssm:*` actions are attached to the role |
| No alerts when creating resources | CloudTrail not enabled | Enable management events in CloudTrail, or set `create_cloudtrail=true` |
| Webhook returns 404 | Invalid Google Chat webhook URL | Re-run `terraform apply` with the correct URL |
| SSM parameter not found | Webhook URL not set | Run `terraform apply -var="google_chat_webhook_url=..."` |

## Cleaning up

To remove all resources and stop incurring any costs:

```bash
cd terraform
terraform destroy
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
