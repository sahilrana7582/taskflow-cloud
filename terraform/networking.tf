locals {
  vpc_cidr = "10.0.0.0/16"

  # Third octet encodes the tier: 1-9 public, 10-19 database, 20-29 app.
  # Maps are keyed by AZ name, so each subnet is tracked by its AZ rather
  # than its position in a list.
  public_subnet_cidrs = zipmap(local.azs, ["10.0.1.0/24", "10.0.2.0/24"])
  db_subnet_cidrs     = zipmap(local.azs, ["10.0.10.0/24", "10.0.11.0/24"])
  app_subnet_cidrs    = zipmap(local.azs, ["10.0.21.0/24", "10.0.22.0/24"])
}

resource "aws_vpc" "main" {
  cidr_block           = local.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = { Name = "${local.name_prefix}-vpc" }
}

resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = { Name = "${local.name_prefix}-igw" }
}

# ---------- Subnets ----------

resource "aws_subnet" "public" {
  for_each = local.public_subnet_cidrs

  vpc_id            = aws_vpc.main.id
  availability_zone = each.key
  cidr_block        = each.value

  # A subnet is public because its route table points at the IGW, not
  # because instances get public IPs. The ALB and NAT get theirs explicitly.
  map_public_ip_on_launch = false

  tags = {
    Name = "${local.name_prefix}-public-${each.key}"
    Tier = "public"
  }
}

resource "aws_subnet" "app" {
  for_each = local.app_subnet_cidrs

  vpc_id            = aws_vpc.main.id
  availability_zone = each.key
  cidr_block        = each.value

  tags = {
    Name = "${local.name_prefix}-app-${each.key}"
    Tier = "app"
  }
}

resource "aws_subnet" "db" {
  for_each = local.db_subnet_cidrs

  vpc_id            = aws_vpc.main.id
  availability_zone = each.key
  cidr_block        = each.value

  tags = {
    Name = "${local.name_prefix}-db-${each.key}"
    Tier = "db"
  }
}

# ---------- NAT (single, in the first AZ) ----------

resource "aws_eip" "nat" {
  domain = "vpc"

  tags = { Name = "${local.name_prefix}-nat-eip" }
}

resource "aws_nat_gateway" "main" {
  allocation_id = aws_eip.nat.id
  subnet_id     = aws_subnet.public[local.azs[0]].id

  tags = { Name = "${local.name_prefix}-nat" }

  # Nothing in the config links the NAT to the IGW. Without this, Terraform
  # may delete both in parallel on destroy, and AWS refuses to detach an IGW
  # while a NAT's public IP is still mapped in the VPC.
  depends_on = [aws_internet_gateway.main]
}

# ---------- Route tables ----------

# Public: default route to the IGW.
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  tags = { Name = "${local.name_prefix}-public-rt" }
}

resource "aws_route" "public_internet" {
  route_table_id         = aws_route_table.public.id
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.main.id
}

resource "aws_route_table_association" "public" {
  for_each = aws_subnet.public

  subnet_id      = each.value.id
  route_table_id = aws_route_table.public.id
}

# App: one table per AZ, all pointing at the single NAT for now. Moving to a
# NAT per AZ later only changes the route target, not the associations.
resource "aws_route_table" "app" {
  for_each = toset(local.azs)

  vpc_id = aws_vpc.main.id

  tags = { Name = "${local.name_prefix}-app-rt-${each.key}" }
}

resource "aws_route" "app_nat" {
  for_each = aws_route_table.app

  route_table_id         = each.value.id
  destination_cidr_block = "0.0.0.0/0"
  nat_gateway_id         = aws_nat_gateway.main.id
}

resource "aws_route_table_association" "app" {
  for_each = aws_subnet.app

  subnet_id      = each.value.id
  route_table_id = aws_route_table.app[each.key].id
}

# Database: explicit table with only the implicit local route. No path to
# or from the internet. Explicit so the DB tier never inherits whatever
# someone later adds to the VPC's main route table.
resource "aws_route_table" "db" {
  vpc_id = aws_vpc.main.id

  tags = { Name = "${local.name_prefix}-db-rt" }
}

resource "aws_route_table_association" "db" {
  for_each = aws_subnet.db

  subnet_id      = each.value.id
  route_table_id = aws_route_table.db.id
}
