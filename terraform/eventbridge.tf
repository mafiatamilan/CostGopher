locals {
  costed_events = [
    "RunInstances", "CreateInstance",
    "CreateDBInstance",
    "CreateBucket",
    "CreateFunction", "CreateFunction20150331",
    "CreateCacheCluster",
    "CreateElasticsearchDomain",
    "CreateCluster",
    "CreateStream",
    "CreateTable",
    "CreateQueue",
    "CreateTopic",
    "CreateService", "RunTask",
    "CreateAutoScalingGroup",
    "CreateLoadBalancer",
  ]

  costed_sources = [
    "aws.ec2",
    "aws.rds",
    "aws.s3",
    "aws.lambda",
    "aws.elasticache",
    "aws.es",
    "aws.redshift",
    "aws.kinesis",
    "aws.dynamodb",
    "aws.sqs",
    "aws.sns",
    "aws.ecs",
    "aws.eks",
    "aws.autoscaling",
    "aws.elasticloadbalancing",
  ]
}

resource "aws_cloudwatch_event_rule" "resource_alert" {
  name        = "costgopher-resource-alert"
  description = "Match AWS resource-creation API calls that have a cost"

  event_pattern = jsonencode({
    source      = local.costed_sources
    detail-type = ["AWS API Call via CloudTrail"]
    detail = {
      eventName = local.costed_events
    }
  })

  tags = {
    Name = "costgopher-resource-alert"
  }
}

resource "aws_cloudwatch_event_rule" "weekly_bill" {
  name                = "costgopher-weekly-bill"
  description         = "Trigger weekly AWS bill summary every Monday at 9am UTC"
  schedule_expression = "cron(0 9 ? * MON *)"

  tags = {
    Name = "costgopher-weekly-bill"
  }
}

resource "aws_cloudwatch_event_rule" "mid_month_forecast" {
  name                = "costgopher-mid-month-forecast"
  description         = "Trigger mid-month cost forecast on the 15th at 9am UTC"
  schedule_expression = "cron(0 9 15 * *)"

  tags = {
    Name = "costgopher-mid-month-forecast"
  }
}
