#!/bin/sh

service ssh start
service ssh status

# Keep the container running and available for operators' use.
while true; do
  sleep 30;
done
