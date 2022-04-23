#!/bin/sh

set -e

DOMAIN=$1
WORKDIR=$(dirname "$0")
WORKDIR=$(realpath "$WORKDIR")
WORKDIR=$(dirname "$WORKDIR")
cd $WORKDIR/server

docker build -t coolcar/$DOMAIN -f ../deployment/$DOMAIN/Dockerfile .
