terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.13.1"
    }
  }

  backend "s3" {
    bucket = var.terraform_backend_bucket
    key    = "terraform.tfstate"
    region = var.region
  }
}

provider "aws" {
  region = var.region
}

data "aws_caller_identity" "current" {}

output "lambda_function_name" {
  value = aws_lambda_function.go_lambda.function_name
}

output "lambda_api_v2_endpoint" {
  value = aws_apigatewayv2_api.lambda_api_v2.api_endpoint
}
