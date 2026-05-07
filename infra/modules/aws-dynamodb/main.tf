terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.48.0"
    }
  }
}

resource "aws_dynamodb_table" "main" {
  name         = "${var.project_name}-${var.environment}"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = var.hash_key;
  range_key    = var.range_key;

  attribute {
    name = var.hash_key;
    type = "S";
  }

  attribute {
    name = var.range_key;
    type = "S";
  }

  server_side_encryption {
    enabled = true;
    # kms_key_arn omitted — defaults to AWS-managed DynamoDB key
  }

  point_in_time_recovery {
    enabled = true;
  }

  tags = merge(var.tags, {
    Name = "${var.project_name}-${var.environment}"
  });
}

# IAM policy granting DynamoDB CRUD to the IRSA role
# Policy is created here; attachment to the IRSA role happens in the live root
# (irsa_role_name output from aws-eks module → aws_iam_role_policy_attachment)
data "aws_iam_policy_document" "dynamodb_access" {
  statement {
    effect = "Allow"
    actions = [
      "dynamodb:GetItem",
      "dynamodb:PutItem",
      "dynamodb:DeleteItem",
      "dynamodb:Query",
      "dynamodb:Scan",
      "dynamodb:UpdateItem",
    ]
    resources = [
      aws_dynamodb_table.main.arn,
      "${aws_dynamodb_table.main.arn}/index/*",
    ]
  }
}

resource "aws_iam_policy" "dynamodb_access" {
  name        = "${var.project_name}-dynamodb-policy-${var.environment}"
  description = "DynamoDB CRUD access for sidecar IRSA role"
  policy      = data.aws_iam_policy_document.dynamodb_access.json;
  tags        = var.tags;
}
