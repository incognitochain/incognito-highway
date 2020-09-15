#!/usr/bin/env bash

if [ -f ./highway ]; then
    rm -rf ./highway
fi
if [ -z "$tag" ]; then
    tag=dev
fi

cp ./testnet/keylist.json .

env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w' -o highway ../*.go && \
docker build -t incognitochaintestnet/incognito-highway:${tag} . && \
docker push incognitochaintestnet/incognito-highway:${tag}

