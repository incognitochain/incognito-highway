#!/usr/bin/env bash

if [ "$1" == "all" ]; then
go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI= -proxy_port 9330 -support_shards all -admin_port 8330 --loglevel debug # QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv
fi

if [ "$1" == "beacon" ]; then
go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI= -proxy_port 9330 -support_shards beacon -admin_port 8330 --loglevel info # QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv
fi

# if [ "$1" == "s0" ]; then
# go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCCSIsG5e3ESqesSDqAcN3WIMUn9IoX9XTROixfYSQEMmKAKBggqhkjOPQMBB6FEA0IABDZNozvtHZvMkVVsks85svDPvvy0bw+/Zt7yCu7N66Cd1hchsR1CIe6U2q5E9uWke5Q+DZdjB5BPxq+ChnsDOeE= -proxy_port 9331 -support_shards 0 -bootstrap "/ip4/127.0.0.1/tcp/9330/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv,/ip4/127.0.0.1/tcp/9332/p2p/QmSZGxdbG1AfuDX1ia6kRyewrVbBNkjd6NU4AbZUCE2ZyG" --loglevel info # QmWoKWHPGjUNbjohYS2MySNDBiCXA2tDs3jT6P3MXDgb9D
# fi

# if [ "$1" == "s1" ]; then
# go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCDbZ14KJ1kqzxuGj1efoFnO+E2I14oURe4hmrez0IbbbaAKBggqhkjOPQMBB6FEA0IABId4AxM4O4hvBMv0voeENGphYll5E4VuKZ1RnXefBLIles0ay515b2QkF3B5BxGBMd7J9VYyfVGh2/NnJaOfvkI= -proxy_port 9332 -support_shards 1 --bootstrap "/ip4/127.0.0.1/tcp/9330/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv,/ip4/127.0.0.1/tcp/9331/p2p/QmWoKWHPGjUNbjohYS2MySNDBiCXA2tDs3jT6P3MXDgb9D" --loglevel info # QmSZGxdbG1AfuDX1ia6kRyewrVbBNkjd6NU4AbZUCE2ZyG
# fi

# QmRfp2xEWdRwagEfxVXTJee6xqaJnWTPogpKKdutq3FHNQ
if [ "$1" == "local" ]; then
go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI=  -support_shards all -proxy_port 9330 -bootstrap dummy
fi

if [ "$1" == "server" ]; then
# PUBLIC_IP=`dig -4 @resolver1.opendns.com ANY myip.opendns.com +short`;
go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI=  -support_shards all -proxy_port 9330 -bootstrap dummy -host "0.0.0.0"
fi