provider "aws" {
  region = "eu-central-1"
}

data "aws_iam_policy_document" "bedrock" {
  statement {
    effect    = "Allow"
    actions   = ["bedrock:InvokeModel"]
    resources = ["arn:aws:bedrock:eu-central-1::foundation-model/anthropic.claude-sonnet-4-6"]
  }
}

resource "aws_ssm_parameter" "model" {
  name  = "model"
  type  = "String"
  value = "eu.anthropic.claude-sonnet-4-6"
}
