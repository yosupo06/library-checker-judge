variable "env" {
  type = string
  default = "test"
}

source "googlecompute" "judge" {
  project_id = "library-checker-project"
  source_image = "ubuntu-2204-jammy-v20220622"
  zone = "asia-northeast1-b"
  machine_type = "n1-standard-2"
  disk_size = 50
  ssh_username = "ubuntu"
  temporary_key_pair_type = "ed25519"
  image_name = "${var.env}-judge-image-{{timestamp}}"
  image_family = "${var.env}-judge-image-family"
}

build {
  sources = ["sources.googlecompute.judge"]

  # wait for cloud-init
  provisioner "shell" {
    inline = [
      "while [ ! -f /var/lib/cloud/instance/boot-finished ]; do echo 'Waiting for cloud-init...'; sleep 1; done"
    ]
  }

  # apt-get
  provisioner "shell" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get upgrade -y",
      "sudo apt-get install -y cgroup-tools postgresql-client unzip git golang-go"
    ]
  }

  # mount setting
  provisioner "file" {
    source = "create-ramdisk.sh"
    destination = "/tmp/create-ramdisk.sh"
  }
  provisioner "shell" {
    inline = [
      "sudo cp /tmp/create-ramdisk.sh /var/lib/cloud/scripts/per-boot/create-ramdisk.sh",
      "sudo chmod 755 /var/lib/cloud/scripts/per-boot/create-ramdisk.sh",
    ]
  }

  # install python, pip, pip-packages
  provisioner "shell" {
    inline = [
      "sudo apt-get install -y python3-pip python3 python3-dev",
      "sudo python3 -m pip install --upgrade pip",
      "sudo python3 -m pip install minio grpcio-tools",
    ]
  }

  # install docker
  provisioner "file" {
    source = "docker-install.sh"
    destination = "/tmp/docker-install.sh"
  }
  provisioner "shell" {
    inline = [
      "sudo sh /tmp/docker-install.sh"
    ]
  }
  # docker mount setting
  provisioner "file" {
    source = "prepare-docker.sh"
    destination = "/tmp/prepare-docker.sh"
  }
  provisioner "shell" {
    inline = [
      "sudo cp /tmp/prepare-docker.sh /var/lib/cloud/scripts/per-boot/prepare-docker.sh",
      "sudo chmod 755 /var/lib/cloud/scripts/per-boot/prepare-docker.sh",
    ]
  }
  

  # build our images
  provisioner "file" {
    source = "../langs"
    destination = "/tmp"
  }
  provisioner "shell" {
    inline = [
      "sudo /tmp/langs/build.sh",
    ]
  }

  # prepare docker
  provisioner "shell" {
    inline = [
      "sudo service docker stop",
      "sudo mv /var/lib/docker /var/lib/docker-base",
    ]
  }
  provisioner "file" {
    source = "docker-daemon.json"
    destination = "/tmp/daemon.json"
  }
  provisioner "shell" {
    inline = [
      "sudo cp /tmp/daemon.json /etc/docker/daemon.json"
    ]
  }
  provisioner "file" {
    source = "docker.service"
    destination = "/tmp/docker.service"
  }
  provisioner "shell" {
    inline = [
      "sudo cp /tmp/docker.service /lib/systemd/system/docker.service",
      "sudo systemctl daemon-reload",
    ]
  }

  # install crun
  provisioner "file" {
    source = "crun-install.sh"
    destination = "/tmp/crun-install.sh"
  }
  provisioner "shell" {
    inline = [
      "sudo sh /tmp/crun-install.sh"
    ]
  }
}
