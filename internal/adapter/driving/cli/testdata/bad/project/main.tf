provider "aws" {
  region = "us-east-1"
}

data "aws_iam_policy_document" "bedrock" {
  statement {
    effect    = "Allow"
    actions   = ["bedrock:*"]
    resources = ["*"]
  }
}

variable "api_key" {
  default = "sk-ant-GEHEIM-TERRAFORM"
}

resource "aws_ssm_parameter" "model" {
  name  = "model"
  type  = "String"
  value = "us.anthropic.claude-3-5-sonnet-20241022-v2:0"
}
