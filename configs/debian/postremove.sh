#!/bin/sh
# postrm script for prometheus-node-exporter
# Script executed after the package is removed.

set -e

case "$1" in
  purge)
    printf "\033[32mReload the service unit from disk\033[0m\n"
    systemctl daemon-reload ||:

    printf "\033[33mRemove system user\033[0m\n"
    userdel --remove prometheus-rds-exporter
	;;
esac

#DEBHELPER#
