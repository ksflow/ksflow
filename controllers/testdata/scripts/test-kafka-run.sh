#!/bin/bash

# copy files to where they should go
mkdir -p /opt/bitnami/kafka/config/certs/
cp /*.jks /opt/bitnami/kafka/config/certs/
cp /*.p12 /opt/bitnami/kafka/config/certs/

# wait to find out the real port
until [ -f /testcontainers_start.sh ]
do
  sleep 0.1
done
. /testcontainers_start.sh

# run original cmd
set -o errexit
set -o nounset
set -o pipefail
. /opt/bitnami/scripts/liblog.sh
. /opt/bitnami/scripts/libbitnami.sh
. /opt/bitnami/scripts/libkafka.sh
. /opt/bitnami/scripts/kafka-env.sh
info "** Starting Kafka setup **"
/opt/bitnami/scripts/kafka/setup.sh
info "** Kafka setup finished! **"
echo ""
exec "/opt/bitnami/scripts/kafka/run.sh"