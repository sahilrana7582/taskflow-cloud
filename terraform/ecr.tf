# # This repository was created by hand before Terraform existed here. The
# # import block adopts it (and the image inside it) into state on the next
# # apply instead of creating a new one. After that apply it does nothing, and
# # you can delete it.
# import {
#   to = aws_ecr_repository.api
#   id = "sahilrana27582/taskflow"
# }

resource "aws_ecr_repository" "api" {
  # Must match the existing name exactly: a different name means a new
  # repository, and Terraform would plan to replace this one.
  name = "sahilrana27582/taskflow"

  # A pushed tag can never be overwritten, so the commit SHA in a task
  # definition always means exactly the bytes CI built from that commit.
  image_tag_mutability = "IMMUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }
}

resource "aws_ecr_lifecycle_policy" "api" {
  repository = aws_ecr_repository.api.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Remove untagged images after 1 day"
        selection = {
          tagStatus   = "untagged"
          countType   = "sinceImagePushed"
          countUnit   = "days"
          countNumber = 1
        }
        action = { type = "expire" }
      },
      {
        # Must be the last rule: a tagStatus "any" rule needs the highest
        # rulePriority. Keeps enough history to roll back.
        rulePriority = 2
        description  = "Keep only the 30 most recent images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 30
        }
        action = { type = "expire" }
      },
    ]
  })
}
