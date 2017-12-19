provider "aws" {
  /*
  $ export AWS_ACCESS_KEY_ID=YOUR_AWS_ACCESS_KEY_ID
  $ export AWS_SECRET_ACCESS_KEY=YOUR_AWS_SECRET_ACCESS_KEY
  $ export AWS_DEFAULT_REGION=us-east-1
  */
  region = "${var.region}"
}

resource "aws_key_pair" "kismatic" {
  key_name   = "${var.cluster_name}"
  public_key = "${file("${var.public_ssh_key_path}")}"
}


