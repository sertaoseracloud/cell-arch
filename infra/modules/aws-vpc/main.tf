terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.48.0"
    }
  }
}

# ── Hub VPC ─────────────────────────────────────────────────────
# CIDR 10.0.0.0/16 (65k IPs) — NAT Gateway, firewall, shared services

resource "aws_vpc" "hub" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags = merge(var.tags, {
    Name = "${var.project_name}-hub-vpc-${var.environment}"
  })
}

resource "aws_internet_gateway" "hub" {
  vpc_id = aws_vpc.hub.id
  tags = merge(var.tags, {
    Name = "${var.project_name}-hub-igw-${var.environment}"
  })
}

# Hub private subnets (for NAT GW, etc.)
resource "aws_subnet" "hub_private_a" {
  vpc_id            = aws_vpc.hub.id
  cidr_block        = "10.0.0.0/22"
  availability_zone = "${var.aws_region}a"
  map_public_ip_on_launch = false
  tags = merge(var.tags, {
    Name = "${var.project_name}-hub-private-a-${var.environment}"
  })
}

resource "aws_subnet" "hub_private_b" {
  vpc_id            = aws_vpc.hub.id
  cidr_block        = "10.0.4.0/22"
  availability_zone = "${var.aws_region}b"
  map_public_ip_on_launch = false
  tags = merge(var.tags, {
    Name = "${var.project_name}-hub-private-b-${var.environment}"
  })
}

# NAT Gateway subnet (needs EIP + path to IGW)
resource "aws_subnet" "hub_nat" {
  vpc_id            = aws_vpc.hub.id
  cidr_block        = "10.0.8.0/28"
  availability_zone = "${var.aws_region}a"
  map_public_ip_on_launch = false
  tags = merge(var.tags, {
    Name = "${var.project_name}-hub-nat-subnet-${var.environment}"
  })
}

resource "aws_eip" "nat" {
  domain = "vpc"
  tags = merge(var.tags, {
    Name = "${var.project_name}-hub-nat-eip-${var.environment}"
  })
}

resource "aws_nat_gateway" "hub" {
  allocation_id = aws_eip.nat.id
  subnet_id   = aws_subnet.hub_nat.id
  tags = merge(var.tags, {
    Name = "${var.project_name}-hub-nat-${var.environment}"
  })
}

# Route table for NAT subnet (0.0.0.0/0 → IGW)
resource "aws_route_table" "hub_nat_subnet" {
  vpc_id = aws_vpc.hub.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.hub.id
  }
  tags = merge(var.tags, {
    Name = "${var.project_name}-hub-nat-rt-${var.environment}"
  })
}

resource "aws_route_table_association" "hub_nat_subnet" {
  subnet_id      = aws_subnet.hub_nat.id
  route_table_id = aws_route_table.hub_nat_subnet.id
}

# Route table for Hub private subnets (0.0.0.0/0 → NAT GW)
resource "aws_route_table" "hub_private" {
  vpc_id = aws_vpc.hub.id
  route {
    cidr_block = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.hub.id
  }
  tags = merge(var.tags, {
    Name = "${var.project_name}-hub-private-rt-${var.environment}"
  })
}

resource "aws_route_table_association" "hub_private_a" {
  subnet_id      = aws_subnet.hub_private_a.id
  route_table_id = aws_route_table.hub_private.id
}

resource "aws_route_table_association" "hub_private_b" {
  subnet_id      = aws_subnet.hub_private_b.id
  route_table_id = aws_route_table.hub_private.id
}

# ── Spoke VPC ───────────────────────────────────────────────────
# CIDR 10.1.0.0/16 (65k IPs) — EKS nodes in private subnets

resource "aws_vpc" "spoke" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags = merge(var.tags, {
    Name = "${var.project_name}-spoke-vpc-${var.environment}"
  })
}

# Spoke private subnets (EKS nodes — no public IPs)
resource "aws_subnet" "spoke_private_a" {
  vpc_id            = aws_vpc.spoke.id
  cidr_block        = "10.1.0.0/22"
  availability_zone = "${var.aws_region}a"
  map_public_ip_on_launch = false
  tags = merge(var.tags, {
    Name = "${var.project_name}-spoke-private-a-${var.environment}"
  })
}

resource "aws_subnet" "spoke_private_b" {
  vpc_id            = aws_vpc.spoke.id
  cidr_block        = "10.1.4.0/22"
  availability_zone = "${var.aws_region}b"
  map_public_ip_on_launch = false
  tags = merge(var.tags, {
    Name = "${var.project_name}-spoke-private-b-${var.environment}"
  })
}

# Route table for Spoke private subnets (0.0.0.0/0 → VPC Peering → Hub NAT)
resource "aws_route_table" "spoke_private" {
  count   = 2
  vpc_id = aws_vpc.spoke.id
  route {
    cidr_block = "0.0.0.0/0"
    vpc_peering_connection_id = aws_vpc_peering_connection.hub_spoke.id
  }
  tags = merge(var.tags, {
    Name = "${var.project_name}-spoke-private-rt-${var.environment}-${count.index}"
  })
}

resource "aws_route_table_association" "spoke_private_a" {
  subnet_id      = aws_subnet.spoke_private_a.id
  route_table_id = aws_route_table.spoke_private[0].id
}

resource "aws_route_table_association" "spoke_private_b" {
  subnet_id      = aws_subnet.spoke_private_b.id
  route_table_id = aws_route_table.spoke_private[1].id
}

# ── VPC Peering (Hub ↔ Spoke) ────────────────────────────────

resource "aws_vpc_peering_connection" "hub_spoke" {
  vpc_id      = aws_vpc.hub.id
  peer_vpc_id = aws_vpc.spoke.id
  auto_accept = true
  tags = merge(var.tags, {
    Name = "${var.project_name}-hub-spoke-peering-${var.environment}"
  })
}

# Hub route to Spoke (10.1.0.0/16 → peering)
resource "aws_route" "hub_to_spoke" {
  route_table_id            = aws_route_table.hub_private.id
  destination_cidr_block = "10.1.0.0/16"
  vpc_peering_connection_id = aws_vpc_peering_connection.hub_spoke.id
}

# ── DynamoDB Gateway Endpoint (Pattern 6 from RESEARCH.md) ─────

resource "aws_vpc_endpoint" "dynamodb" {
  vpc_id            = aws_vpc.spoke.id
  service_name      = "com.amazonaws.${var.aws_region}.dynamodb"
  vpc_endpoint_type = "Gateway"
  route_table_ids   = aws_route_table.spoke_private[*].id
  tags = merge(var.tags, {
    Name = "${var.project_name}-dynamodb-ep-${var.environment}"
  })
}
