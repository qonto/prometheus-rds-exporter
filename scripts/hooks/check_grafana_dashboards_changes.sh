#!/usr/bin/env bash

make dashboards
if [ $? -ne 0 ]; then
  echo ""
  echo "ERROR: Failed to make dashboards"
  exit 2
fi

git diff --exit-code configs/grafana/public
if [ $? -ne 0 ]; then
  echo ""
  echo "Grafana dashboards has been updated, please re-add it to your commit"
  exit 3
fi
