#======================================================
# variables
variable "redis-count" {
  description = "Number of redis instances to run"
  default     = "1"
}

variable "redis-version" {
  description = "Number of redis instances to run"
  default     = "5.0.3"
}

variable "redis-size" {
  description = "Size of the redis cache"
  default = "cache.t2.micro"
}

variable "redis-port" {
  description = "Port to run Redis on - recommend non-standard"
  default = "6379"
}

variable "allowed-cidrs" {
  description = "CIDRS that can connect to redis"
  default = ["76.14.80.208/32"]
}

resource "aws_elasticache_parameter_group" "testredis" {
  name   = "testredis"
  family = "redis${join(".", slice(split(".", var.redis-version), 0, 2))}"
}

#======================================================
# network

provider "aws" {
  region  = "us-west-1"
  profile = "redislabs"
}


resource "aws_vpc" "redistest" {
  cidr_block           = "10.0.10.0/24"
  tags {
    Name = "redis-test-VPC"
  }
}

resource "aws_subnet" "internal" {
  vpc_id            = "${aws_vpc.redistest.id}"
  cidr_block        = "10.0.10.0/25"
  availability_zone = "us-west-1a"
  tags {
    Name = "testredis-subnet"
  }
}


resource "aws_elasticache_cluster" "testredis" {
  count                = "${var.redis-count}"
  cluster_id           = "testredis-${count.index}"
  engine               = "redis"
  engine_version       = "${var.redis-version}"
  node_type            = "${var.redis-size}"
  port                 = "${var.redis-port}"
  num_cache_nodes      = 1
  subnet_group_name    = "${aws_elasticache_subnet_group.testredis.name}"
  parameter_group_name = "${aws_elasticache_parameter_group.testredis.name}"
  security_group_ids   = ["${aws_security_group.testredis.id}"]
  depends_on           = ["aws_security_group.testredis"]
}

resource "aws_elasticache_subnet_group" "testredis" {
  name       = "testredis-cache-subnet"
  subnet_ids = ["${aws_subnet.internal.id}"]
}

resource "aws_security_group" "testredis" {
  name        = "testredis"
  description = "Test Redis"
  vpc_id      = "${aws_vpc.redistest.id}"

  ###############################################################################
  ingress {
    from_port   = "${var.redis-port}"
    to_port     = "${var.redis-port}"
    protocol    = "tcp"
    cidr_blocks = ["${var.allowed-cidrs}"]
  }

  ###############################################################################
  # Allow everything going out
  egress {
    from_port   = "${var.redis-port}"
    to_port     = "${var.redis-port}"
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
