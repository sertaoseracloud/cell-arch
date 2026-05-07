terraform {
  required_version = "~> 1.7.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.48.0"
    }
  }

  backend "s3" {
    bucket         = "cell-arch-tfstate-dev"
    key            = "dev/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "cell-arch-tfstate-lock-dev"
    encrypt        = true
  }
}

provider "aws" {
  region = "us-east-1"
}
