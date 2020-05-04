#!/usr/bin/env bash
mkdir -p /data
cron
if [ -z "$BOOTSTRAP" ]; then
    BOOTSTRAP=45.56.115.6:9330;
fi

if [ -z "$PUBLIC_IP" ]; then
    PUBLIC_IP=`dig -4 @resolver1.opendns.com ANY myip.opendns.com +short`;
fi

if [ -z "$PRIVATE_KEY" ]; then
    PRIVATE_KEY=CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI=
fi

if [ -z "$GDBURL" ]; then
    GDBURL=""
fi

if [ -z "$VERSION" ]; then
    VERSION="version"
fi

echo './highway -privatekey $PRIVATE_KEY -support_shards all -host $PUBLIC_IP --bootstrap $BOOTSTRAP --gdburl $GDBURL --version $VERSION --loglevel debug 2>&1 | cronolog /data/highway-$PUBLIC_IP-%Y-%m-%d.log'
./highway -privatekey $PRIVATE_KEY -support_shards all -host $PUBLIC_IP --bootstrap $BOOTSTRAP --gdburl $GDBURL --version $VERSION --loglevel debug 2>&1 | cronolog /data/highway-$PUBLIC_IP-%Y-%m-%d.log
