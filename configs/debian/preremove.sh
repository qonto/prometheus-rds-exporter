#!/bin/sh
# prerm script for prometheus-node-exporter
# Script executed before the package is removed.

printf "\033[32mMask the service\033[0m\n"
systemctl mask prometheus-rds-exporter ||:

printf "\033[32mSet the enabled flag for the service unit\033[0m\n"
systemctl disable prometheus-rds-exporter ||:

printf "\033[33mStop the service unit\033[0m\n"
systemctl stop prometheus-rds-exporter ||:
