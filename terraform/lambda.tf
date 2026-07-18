data "archive_file" "resource_alert" {
  type        = "zip"
  source_file = "../dist/resource-alert/bootstrap"
  output_path = "../dist/resource-alert.zip"
}

data "archive_file" "weekly_bill" {
  type        = "zip"
  source_file = "../dist/weekly-bill/bootstrap"
  output_path = "../dist/weekly-bill.zip"
}

data "archive_file" "mid_month_forecast" {
  type        = "zip"
  source_file = "../dist/mid-month-forecast/bootstrap"
  output_path = "../dist/mid-month-forecast.zip"
}

resource "aws_lambda_function" "resource_alert" {
  filename         = data.archive_file.resource_alert.output_path
  function_name    = "costgopher-resource-alert"
  role             = aws_iam_role.lambda.arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  source_code_hash = data.archive_file.resource_alert.output_base64sha256
  timeout          = 10
  memory_size      = 128

  environment {
    variables = {
      POWERTOOLS_SERVICE_NAME = "resource-alert"
    }
  }

  tags = {
    Name = "costgopher-resource-alert"
  }
}

resource "aws_lambda_function" "weekly_bill" {
  filename         = data.archive_file.weekly_bill.output_path
  function_name    = "costgopher-weekly-bill"
  role             = aws_iam_role.lambda.arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  source_code_hash = data.archive_file.weekly_bill.output_base64sha256
  timeout          = 30
  memory_size      = 128

  environment {
    variables = {
      POWERTOOLS_SERVICE_NAME = "weekly-bill"
    }
  }

  tags = {
    Name = "costgopher-weekly-bill"
  }
}

resource "aws_lambda_function" "mid_month_forecast" {
  filename         = data.archive_file.mid_month_forecast.output_path
  function_name    = "costgopher-mid-month-forecast"
  role             = aws_iam_role.lambda.arn
  handler          = "bootstrap"
  runtime          = "provided.al2023"
  source_code_hash = data.archive_file.mid_month_forecast.output_base64sha256
  timeout          = 30
  memory_size      = 128

  environment {
    variables = {
      POWERTOOLS_SERVICE_NAME = "mid-month-forecast"
    }
  }

  tags = {
    Name = "costgopher-mid-month-forecast"
  }
}

resource "aws_cloudwatch_log_group" "resource_alert" {
  name              = "/aws/lambda/costgopher-resource-alert"
  retention_in_days = var.log_retention_days

  tags = {
    Name = "costgopher-resource-alert-logs"
  }
}

resource "aws_cloudwatch_log_group" "weekly_bill" {
  name              = "/aws/lambda/costgopher-weekly-bill"
  retention_in_days = var.log_retention_days

  tags = {
    Name = "costgopher-weekly-bill-logs"
  }
}

resource "aws_cloudwatch_log_group" "mid_month_forecast" {
  name              = "/aws/lambda/costgopher-mid-month-forecast"
  retention_in_days = var.log_retention_days

  tags = {
    Name = "costgopher-mid-month-forecast-logs"
  }
}

resource "aws_lambda_permission" "resource_alert" {
  statement_id  = "AllowEventBridgeInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.resource_alert.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.resource_alert.arn
}

resource "aws_lambda_permission" "weekly_bill" {
  statement_id  = "AllowEventBridgeInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.weekly_bill.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.weekly_bill.arn
}

resource "aws_lambda_permission" "mid_month_forecast" {
  statement_id  = "AllowEventBridgeInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.mid_month_forecast.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.mid_month_forecast.arn
}

resource "aws_cloudwatch_event_target" "resource_alert" {
  rule      = aws_cloudwatch_event_rule.resource_alert.name
  arn       = aws_lambda_function.resource_alert.arn
  target_id = "costgopher-resource-alert-target"
}

resource "aws_cloudwatch_event_target" "weekly_bill" {
  rule      = aws_cloudwatch_event_rule.weekly_bill.name
  arn       = aws_lambda_function.weekly_bill.arn
  target_id = "costgopher-weekly-bill-target"
}

resource "aws_cloudwatch_event_target" "mid_month_forecast" {
  rule      = aws_cloudwatch_event_rule.mid_month_forecast.name
  arn       = aws_lambda_function.mid_month_forecast.arn
  target_id = "costgopher-mid-month-forecast-target"
}
