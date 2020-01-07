#!/usr/bin/env bash
mkdir -p /data

echo "/data/*.txt {
  rotate 3
  compress
  missingok
  delaycompress
  copytruncate
  size 1000M
}" > /tmp/logrotate
logrotate -fv /tmp/logrotate

if [ -z "$BOOTSTRAP" ]; then
    BOOTSTRAP=45.56.115.6:9330;
fi

if [ -z "$PUBLIC_IP" ]; then
    PUBLIC_IP=`dig -4 @resolver1.opendns.com ANY myip.opendns.com +short`;
fi

if [ -z "$PRIVATE_KEY" ]; then
    PRIVATE_KEY=CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI=
fi

echo ./highway -privatekey $PRIVATE_KEY -support_shards all -host $PUBLIC_IP --loglevel debug
./highway -privatekey $PRIVATE_KEY -support_shards all -host $PUBLIC_IP --bootstrap $BOOTSTRAP --loglevel debug  > /data/log.txt 2>&1