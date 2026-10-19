variable "aws_region" {
  description = "AWS region for all TaskFlow resources."
  type        = string
  default     = "us-east-1"
}

variable "project_name" {
  description = "Prefix for resource names and the Project tag (e.g. taskflow-cluster)."
  type        = string
  default     = "taskflow"

  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{1,19}$", var.project_name))
    error_message = "project_name must be 2-20 chars: lowercase letters, digits, hyphens, starting with a letter."
  }
}

variable "environment" {
  description = "Deployment environment. Used in tags."
  type        = string
  default     = "dev"

  validation {
    condition     = contains(["dev", "staging", "production"], var.environment)
    error_message = "environment must be one of: dev, staging, prod."
  }
}
