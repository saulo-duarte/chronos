variable "region" {
  type    = string
  description = "AWS region where resources will be created"
  default = "sa-east-1"
}

variable "lambda_function_name" {
  type        = string
  description = "Lambda function name"
  default     = "chronos-api"
}

variable "s3_bucket" {
  type        = string
  description = "S3 bucket where Lambda artifacts are uploaded"
}

variable "s3_key" {
  type        = string
  description = "S3 key of Lambda artifact (uploaded by CI/CD pipeline)"
}

variable "parameter_store_key" {
  type        = string
  description = "SSM Parameter Store key that stores the latest Lambda version"
  default     = "/chronos-api/latest-version"
}

variable "terraform_backend_bucket" {
  type        = string
  description = "S3 bucket for Terraform state backend"
  default     = ""
}
