data "aws_caller_identity" "current" {}

locals {
  # Single place that decides resource names. To run several environments
  # in one account later, change this to "${var.project_name}-${var.environment}".
  name_prefix = var.project_name

  # Pinned on purpose: a data-source lookup can shift during an AZ incident
  # and force subnet (and RDS) replacement. Two AZs is the minimum for both
  # the ALB and an RDS subnet group.
  azs = ["${var.aws_region}a", "${var.aws_region}b"]
  container_port = 8080
  account_id = data.aws_caller_identity.current.account_id
}
