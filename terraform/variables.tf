variable "aws_region" {
  description = "AWS region to deploy into"
  type        = string
  default     = "us-east-1"
}

variable "google_chat_webhook_url" {
  description = "Google Chat incoming webhook URL"
  type        = string
  sensitive   = true
}

variable "create_cloudtrail" {
  description = "Create a new CloudTrail trail for management events"
  type        = bool
  default     = false
}

variable "log_retention_days" {
  description = "CloudWatch Log Group retention in days"
  type        = number
  default     = 14
}
