variable "env" {
  type = string
  default = "test"
}

source "googlecompute" "judge" {
  project_id = "library-checker-project"
  source_image_family = "ubuntu-2204-lts"
  zone = "asia-northeast1-b"
  machine_type = "c2-standard-4"
  disk_size = 50
  ssh_username = "ubuntu"
  temporary_key_pair_type = "ed25519"
  image_name = "v1-${var.env}-base-image-{{timestamp}}"
  image_family = "v1-${var.env}-base-image"
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

  # apt-get
  provisioner "shell" {
    inline = [
      "sudo apt-get update",
      "sudo apt-get upgrade -y",
      "sudo apt-get install -y cgroup-tools postgresql-client unzip git golang-go"
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
  
  # build our images
  provisioner "file" {
    source = "../../langs"
    destination = "/tmp"
  }
  provisioner "shell" {
    inline = [
      "sudo /tmp/langs/build.sh",
      "sudo docker image prune -f --all --filter=\"label!=library-checker-image=true\"",
      "sudo docker builder prune --force",
    ]
  }

  # fetch hello-world
  provisioner "shell" {
    inline = [
      "sudo docker pull hello-world",
    ]
  }

  # prepare docker-base
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
      "sudo mkdir -p /etc/docker",
      "sudo cp /tmp/daemon.json /etc/docker/daemon.json",
    ]
  }

  # prepare prepare-docker files
  provisioner "file" {
    source = "prepare-docker.service"
    destination = "/tmp/prepare-docker.service"
  }
  provisioner "file" {
    source = "prepare-docker.sh"
    destination = "/tmp/prepare-docker.sh"
  }
  provisioner "shell" {
    inline = [
      "sudo mkdir -p /usr/local/lib/systemd/system/",
      "sudo cp /tmp/prepare-docker.service /usr/local/lib/systemd/system/prepare-docker.service",
      "sudo cp /tmp/prepare-docker.sh /root/prepare-docker.sh",
    ]
  }

  # prepare prepare-cgroup files
  provisioner "file" {
    source = "prepare-cgroup.service"
    destination = "/tmp/prepare-cgroup.service"
  }
  provisioner "file" {
    source = "prepare-cgroup.sh"
    destination = "/tmp/prepare-cgroup.sh"
  }
  provisioner "shell" {
    inline = [
      "sudo cp /tmp/prepare-cgroup.service /usr/local/lib/systemd/system/prepare-cgroup.service",
      "sudo cp /tmp/prepare-cgroup.sh /root/prepare-cgroup.sh",
    ]
  }

  provisioner "file" {
    source = "docker-drop-in.conf"
    destination = "/tmp/docker-drop-in.conf"
  }
  provisioner "shell" {
    inline = [
      "sudo mkdir /etc/systemd/system/docker.service.d/",
      "sudo cp /tmp/docker-drop-in.conf /etc/systemd/system/docker.service.d/docker-drop-in.conf"
    ]
  }

  provisioner "shell" {
    inline = [
      "sudo systemctl daemon-reload"
    ]
  }
}
