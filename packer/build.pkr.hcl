source "googlecompute" "judge" {
  project_id = "library-checker-project"
  source_image = "ubuntu-2004-focal-v20210315"
  zone = "asia-northeast1-c"
  disk_size = 25
  machine_type = "n1-standard-2"
  ssh_username = "ubuntu"
  image_name = "judge-image-{{timestamp}}"
  image_family = "judge-image-family"
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
      "sudo apt-get upgrade -y"
    ]
  }

  # add user: library-checker-user
  provisioner "shell" {
    inline = [
      "sudo useradd library-checker-user -u 2000 -m",
    ]
  }

  # apt packages
  provisioner "shell" {
    inline = [
      "sudo apt-get install -y cgroup-tools postgresql-client unzip git",
    ]
  }

  # grub setting
  provisioner "file" {
    source = "99-lib-judge.cfg"
    destination = "/tmp/99-lib-judge.cfg"
  }
  provisioner "shell" {
    inline = [      
      "sudo cp /tmp/99-lib-judge.cfg /etc/default/grub.d/99-lib-judge.cfg",
      "sudo update-grub",
    ]
  }

  # supervisor setting
  provisioner "shell" {
    inline = [
      "sudo apt-get install -y supervisor",
    ]
  }
  provisioner "file" {
    source = "judge.conf"
    destination = "/tmp/judge.conf"
  }
  provisioner "shell" {
    inline = [
      "sudo cp /tmp/judge.conf /etc/supervisor/conf.d/judge._conf",
      "sudo chmod 600 /etc/supervisor/conf.d/judge._conf",
    ]
  }

  # mount setting
  provisioner "file" {
    source = "judge-launch.sh"
    destination = "/tmp/judge-launch.sh"
  }
  provisioner "shell" {
    inline = [
      "sudo cp /tmp/judge-launch.sh /var/lib/cloud/scripts/per-boot/judge-launch.sh",
      "sudo chmod 755 /var/lib/cloud/scripts/per-boot/judge-launch.sh",
    ]
  }

  # install python, pip, pip-packages
  provisioner "shell" {
    inline = [
      "sudo apt-get install -y python3-pip python3.8 python3.8-dev",
      "sudo python3.8 -m pip install --upgrade pip",
      "sudo python3.8 -m pip install minio grpcio-tools",
    ]
  }

  # install docker
  provisioner "shell" {
    inline = [
      "curl -fsSL https://get.docker.com -o /tmp/get-docker.sh",
      "sudo sh /tmp/get-docker.sh",
      "sudo curl -L \"https://github.com/docker/compose/releases/download/1.25.0/docker-compose-$(uname -s)-$(uname -m)\" -o /usr/local/bin/docker-compose",
      "sudo chmod +x /usr/local/bin/docker-compose",
    ]
  }

  # install haskell
  provisioner "file" {
    source = "haskell_load.hs"
    destination = "/tmp/haskell_load.hs"
  }
  provisioner "shell" {
    script = "haskell_setup.sh"
  }

  # install C#
  provisioner "shell" {
    script = "c_sharp_setup.sh"
  }

  # install python (numpy, scipy)
  provisioner "shell" {
    inline = [
      "sudo python3.8 -m pip install --upgrade numpy scipy",
    ]
  }

  # install compilers
  provisioner "shell" {
    inline = [
      "sudo apt-get install -y g++ golang-go default-jdk default-jre pypy3 ldc rustc cargo sbcl",
    ]
  }


}
