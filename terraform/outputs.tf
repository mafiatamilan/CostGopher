output "resource_alert_function" {
  description = "Resource Alert Lambda function name"
  value       = aws_lambda_function.resource_alert.function_name
}

output "weekly_bill_function" {
  description = "Weekly Bill Lambda function name"
  value       = aws_lambda_function.weekly_bill.function_name
}

output "mid_month_forecast_function" {
  description = "Mid-Month Forecast Lambda function name"
  value       = aws_lambda_function.mid_month_forecast.function_name
}

output "webhook_url_parameter" {
  description = "SSM Parameter path for Google Chat webhook URL"
  value       = aws_ssm_parameter.webhook_url.name
}

output "resource_registry_parameter" {
  description = "SSM Parameter path for resource registry"
  value       = aws_ssm_parameter.resource_registry.name
}
