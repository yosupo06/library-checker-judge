variable "env" {
  type = string
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
  source_image_family = "v2-${var.env}-base-image"
  zone = "asia-northeast1-b"
  machine_type = "c2-standard-4"
  disk_size = 50
  ssh_username = "ubuntu"
  temporary_key_pair_type = "ed25519"
  image_name = "v2-${var.env}-judge-image-{{timestamp}}"
  image_family = "v2-${var.env}-judge-image"
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
    source = "judge.service"
    destination = "/tmp/judge.service"
  }
  provisioner "file" {
    source = "judge.sh"
    destination = "/tmp/judge.sh"
  }
  provisioner "shell" {
    inline = [
      "sudo cp /tmp/judge /root/judge",
      "sudo cp /tmp/langs.toml /root/langs.toml",
      "sudo cp /tmp/judge.service /usr/local/lib/systemd/system/judge.service",
      "sudo cp /tmp/judge.sh /root/judge.sh",
    ]
  }

  provisioner "shell" {
    inline = [
      "sudo systemctl daemon-reload",
      "sudo systemctl enable judge",
    ]
  }
}
