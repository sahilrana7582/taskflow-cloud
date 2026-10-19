provider "aws" {
  region = var.aws_region

  # Applied to every taggable resource this provider creates.
  # Per-resource tags (like Name) are merged on top.
  default_tags {
    tags = {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "terraform"
      Repository  = "sahilrana7582/taskflow-cloud"
    }
  }
}
