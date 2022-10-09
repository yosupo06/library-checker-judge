variable "env" {
  type = string
  default = "test"
}

source "googlecompute" "judge" {
  project_id = "library-checker-project"
  source_image_family = "v1-${var.env}-base-image"
  zone = "asia-northeast1-b"
  machine_type = "n1-standard-2"
  disk_size = 50
  ssh_username = "ubuntu"
  temporary_key_pair_type = "ed25519"
  image_name = "v1-${var.env}-judge-image-{{timestamp}}"
  image_family = "v1-${var.env}-judge-image"
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
  provisioner "shell" {
    inline = [
      "sudo cp /tmp/judge /root/judge",
      "sudo cp /tmp/langs.toml /root/langs.toml",
      "sudo cp /tmp/judge.service /usr/local/lib/systemd/system/judge.service",
    ]
  }

  provisioner "shell" {
    inline = [
      "sudo systemctl daemon-reload",
      "sudo systemctl enable judge",
    ]
  }
}
