#!/usr/bin/env bash

while true
do
  echo "do"
  date
  ./coordinate.py
  date
  echo "done"

  echo "sleep..."
  sleep 30
done
