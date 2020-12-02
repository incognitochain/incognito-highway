#!/usr/bin/env bash
mkdir -p /data
cron
if [ -z "$BOOTSTRAP" ]; then
    echo "NO BOOTSTRAP"
    exit -1;
fi

if [ -z "$NAME" ]; then
    echo "NO NAME"
    exit -1;
fi

if [ -z "$PUBLIC_IP" ]; then
    PUBLIC_IP=`dig -4 TXT +short o-o.myaddr.l.google.com @ns1.google.com | tr -d '"'`;
fi

if [ -z "$PUBLIC_IP" ]; then
    PUBLIC_IP=`dig -4 @resolver1.opendns.com ANY myip.opendns.com +short`;
fi

if [ -z "$PRIVATE_SEED" ]; then
    echo "NO PRIVATE SEED"
    exit -1;
fi

if [ -z "$INDEX" ]; then
    echo "NO INDEX"
    exit -1;
fi


if [ -z "$GDBURL" ]; then
    GDBURL=""
fi

if [ -z "$VERSION" ]; then
    VERSION="version"
fi

./highway -privateseed $PRIVATE_SEED -index $INDEX -support_shards all -host $PUBLIC_IP --bootstrap $BOOTSTRAP --gdburl $GDBURL --version $VERSION --loglevel debug 2>&1 | cronolog /data/$NAME/highway-$PUBLIC_IP-%Y-%m-%d.log -S /data/$NAME/$PUBLIC_IP.cur.log
