variable "env" {
  type = string
}

variable "image_family" {
  type = string
}

variable "minio_host" {
  type = string
}

locals {
  parsed_judge_service = templatefile("judge.service.pkrtpl", {
    minio_host = var.minio_host
  })
}

packer {
  required_plugins {
    googlecompute = {
      version = ">= 1.1.4"
      source  = "github.com/hashicorp/googlecompute"
    }
  }
}

source "googlecompute" "judge" {
  project_id = "${var.env}-library-checker-project"
  source_image_family = "v3-${var.env}-base-image"
  zone = "asia-northeast1-b"
  machine_type = "c2-standard-4"
  disk_size = 50
  ssh_username = "ubuntu"
  temporary_key_pair_type = "ed25519"
  image_name = "${var.image_family}-{{timestamp}}"
  image_family = "${var.image_family}"
  preemptible = true
}

build {
  sources = ["sources.googlecompute.judge"]

  # wait for cloud-init
  provisioner "shell" {
    inline = [
      "while [ ! -f /var/lib/cloud/instance/boot-finished ]; do echo 'Waiting for cloud-init...'; sleep 1; done"
    ]
  }

  # send judge
  provisioner "file" {
    source = "../../judge/judge"
    destination = "/tmp/judge"
  }
  provisioner "file" {
    source = "../../langs/langs.toml"
    destination = "/tmp/langs.toml"
  }
  provisioner "file" {
    content = local.parsed_judge_service
    destination = "/tmp/judge.service"
  }
  provisioner "shell" {
    inline = [
      "sudo cp /tmp/judge /root/judge",
      "sudo cp /tmp/langs.toml /root/langs.toml",
      "sudo cp /tmp/judge.service /usr/local/lib/systemd/system/judge.service",
    ]
  }
}
