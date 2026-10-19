terraform {
  # 1.11 is the first release where the S3 backend's native lockfile is GA,
  # which we rely on in the remote-state step instead of a DynamoDB table.
  required_version = ">= 1.11.0, < 2.0.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}
