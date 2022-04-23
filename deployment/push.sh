#!/bin/sh

set -e

DOMAIN=$1
VERSION=$2
WORKDIR=$(dirname "$0")
WORKDIR=$(realpath "$WORKDIR")
WORKDIR=$(dirname "$WORKDIR")
cd $WORKDIR/server

docker tag coolcar/$DOMAIN clivezhang/coolcar_$DOMAIN:$VERSION
docker push clivezhang/coolcar_$DOMAIN:$VERSION
