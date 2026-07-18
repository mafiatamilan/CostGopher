resource "aws_ssm_parameter" "webhook_url" {
  name      = "/costgopher/webhook-url"
  type      = "SecureString"
  value     = var.google_chat_webhook_url
  overwrite = true

  tags = {
    Name = "costgopher-webhook-url"
  }
}

resource "aws_ssm_parameter" "resource_registry" {
  name      = "/costgopher/resources"
  type      = "String"
  value     = "[]"
  overwrite = true

  tags = {
    Name = "costgopher-resource-registry"
  }
}
