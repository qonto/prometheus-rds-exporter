#!/usr/bin/env bats

load '/usr/lib/bats/bats-support/load'
load '/usr/lib/bats/bats-assert/load'
load '/usr/lib/bats/bats-file/load'

PACKAGE=/mnt/prometheus-rds-exporter.deb

setup() {
  run bash -c "DEBIAN_FRONTEND=noninteractive dpkg -i ${PACKAGE}"
  assert_success
}

remove_package() {
  run bash -c 'apt-get remove -y prometheus-rds-exporter'
  assert_success
}

purge_package() {
  run bash -c 'apt-get purge -y prometheus-rds-exporter'
  assert_success
}

@test "Test installation" {
  assert_file_exist /usr/bin/prometheus-rds-exporter
  assert_file_exist /usr/share/prometheus-rds-exporter/prometheus-rds-exporter.yaml.sample
  assert_file_exist /etc/systemd/system/prometheus-rds-exporter.service

  run bash -c 'prometheus-rds-exporter --version'
  assert_success

  run bash -c 'prometheus-rds-exporter --version'
  assert_output --regexp '^rds-exporter version'

  run bash -c "dpkg --info ${PACKAGE}"
  assert_output --regexp 'Package: prometheus-rds-exporter'
  assert_output --regexp 'Prometheus exporter for AWS RDS'
  assert_output --regexp 'Depends: adduser'
  assert_output --regexp 'Maintainer: SRE Team'
  assert_output --regexp 'Homepage: https://github.com/qonto/prometheus-rds-exporter'

  run bash -c 'id -u prometheus-rds-exporter'
  assert_success
}

@test 'Check removed package' {
  remove_package

  assert_file_not_exist /usr/bin/prometheus-rds-exporter
}

@test 'Check purged package' {
  purge_package

  assert_file_not_exist /usr/bin/prometheus-rds-exporter
  assert_file_not_exist /usr/share/prometheus-rds-exporter/prometheus-rds-exporter.yaml.sample
  assert_file_not_exist /etc/systemd/system/prometheus-rds-exporter.service

  run bash -c 'id -u prometheus-rds-exporter'
  assert_failure
}
