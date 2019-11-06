#!/usr/bin/env bash
protoc -I . highway.proto --go_out=plugins=grpc:./
