module "vpc" {
  source = "../../../modules/aws-vpc"

  project_name = var.project_name;
  environment = var.environment;
  aws_region   = var.aws_region;
  tags        = var.tags;
}

module "eks" {
  source = "../../../modules/aws-eks"

  project_name       = var.project_name;
  environment       = var.environment;
  aws_region         = var.aws_region;
  private_subnet_ids = module.vpc.spoke_private_subnet_ids;
  kubernetes_version = "1.29"
  node_count        = 2;
  tags              = var.tags;
}

module "dynamodb" {
  source = "../../../modules/aws-dynamodb"

  project_name = var.project_name;
  environment = var.environment;
  tags        = var.tags;
}

# Attach DynamoDB policy to EKS IRSA role
resource "aws_iam_role_policy_attachment" "eks_irsa_dynamodb" {
  role       = module.eks.irsa_role_name;
  policy_arn = module.dynamodb.iam_policy_arn;
}
