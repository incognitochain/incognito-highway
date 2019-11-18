#!/usr/bin/env bash

if [ "$1" == "beacon" ]; then
go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI= -proxy_port 9330 -support_shards beacon # QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv
fi

if [ "$1" == "s1" ]; then
go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCCSIsG5e3ESqesSDqAcN3WIMUn9IoX9XTROixfYSQEMmKAKBggqhkjOPQMBB6FEA0IABDZNozvtHZvMkVVsks85svDPvvy0bw+/Zt7yCu7N66Cd1hchsR1CIe6U2q5E9uWke5Q+DZdjB5BPxq+ChnsDOeE= -proxy_port 9331 -support_shards 0 -bootstrap "/ip4/127.0.0.1/tcp/9330/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv,/ip4/127.0.0.1/tcp/9332/p2p/QmSZGxdbG1AfuDX1ia6kRyewrVbBNkjd6NU4AbZUCE2ZyG" # QmWoKWHPGjUNbjohYS2MySNDBiCXA2tDs3jT6P3MXDgb9D
fi

if [ "$1" == "s2" ]; then
go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCDbZ14KJ1kqzxuGj1efoFnO+E2I14oURe4hmrez0IbbbaAKBggqhkjOPQMBB6FEA0IABId4AxM4O4hvBMv0voeENGphYll5E4VuKZ1RnXefBLIles0ay515b2QkF3B5BxGBMd7J9VYyfVGh2/NnJaOfvkI= -proxy_port 9332 -support_shards 1 --bootstrap "/ip4/127.0.0.1/tcp/9330/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv,/ip4/127.0.0.1/tcp/9331/p2p/QmWoKWHPGjUNbjohYS2MySNDBiCXA2tDs3jT6P3MXDgb9D" # QmSZGxdbG1AfuDX1ia6kRyewrVbBNkjd6NU4AbZUCE2ZyG
fi

# QmRfp2xEWdRwagEfxVXTJee6xqaJnWTPogpKKdutq3FHNQ
# go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCBPHYHjWDGEG4irMbsODtYwv4+PSrrpEA0mOCdymU5sKKAKBggqhkjOPQMBB6FEA0IABJxbC13BH4WFm3ILpGCn6xxjn8Fsqg6GOmufDqJ+o1kRN1cCeYT0Z3agP50iD3TOx55FNCinWhqEUDW3at2+vAU= -support_shards all -proxy_proxy_port 9331 -bootstrap dummy
