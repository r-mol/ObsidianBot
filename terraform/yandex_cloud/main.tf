resource "yandex_compute_disk" "disk" {
  name = "disk"
  type = "network-hdd"
  size = 1000
  image_id = var.yandex_image_id
}

resource "yandex_compute_instance" "syncthing" {
  name        = "syncthing"
  description = "Syncthing instance"
  platform_id = "standard-v1"
  allow_stopping_for_update = true

  resources {
    cores  = 4
    memory = 8
  }

  # scheduling_policy {
  #   preemptible = true
  # }

  boot_disk {
    disk_id = yandex_compute_disk.disk.id
  }

  network_interface {
    subnet_id = yandex_vpc_subnet.subnet.id
    nat       = true
  }

  metadata = {
    ssh-keys = "roman_molochkov:${file("~/.ssh/id_ed25519.pub")}"
  }
}

resource "yandex_vpc_network" "network" {
  name = "network"
}

resource "yandex_vpc_subnet" "subnet" {
  name           = "subnet"
  zone           = "ru-central1-a"
  network_id     = yandex_vpc_network.network.id
  v4_cidr_blocks = ["192.168.10.0/24"]
}