#!/usr/bin/env bash

if [ "$1" == "all" ]; then
go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI= -support_shards all -admin_port 8330 --loglevel debug # QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv
fi

if [ "$1" == "beacon" ]; then
go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI= -support_shards beacon -admin_port 8330 --loglevel info # QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv
fi

# if [ "$1" == "s0" ]; then
# go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCCSIsG5e3ESqesSDqAcN3WIMUn9IoX9XTROixfYSQEMmKAKBggqhkjOPQMBB6FEA0IABDZNozvtHZvMkVVsks85svDPvvy0bw+/Zt7yCu7N66Cd1hchsR1CIe6U2q5E9uWke5Q+DZdjB5BPxq+ChnsDOeE= -proxy_port 9331 -support_shards 0 -bootstrap "/ip4/127.0.0.1/tcp/9330/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv,/ip4/127.0.0.1/tcp/9332/p2p/QmSZGxdbG1AfuDX1ia6kRyewrVbBNkjd6NU4AbZUCE2ZyG" --loglevel info # QmWoKWHPGjUNbjohYS2MySNDBiCXA2tDs3jT6P3MXDgb9D
# fi

# if [ "$1" == "s1" ]; then
# go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCDbZ14KJ1kqzxuGj1efoFnO+E2I14oURe4hmrez0IbbbaAKBggqhkjOPQMBB6FEA0IABId4AxM4O4hvBMv0voeENGphYll5E4VuKZ1RnXefBLIles0ay515b2QkF3B5BxGBMd7J9VYyfVGh2/NnJaOfvkI= -proxy_port 9332 -support_shards 1 --bootstrap "/ip4/127.0.0.1/tcp/9330/p2p/QmSPa4gxx6PRmoNRu6P2iFwEwmayaoLdR5By3i3MgM9gMv,/ip4/127.0.0.1/tcp/9331/p2p/QmWoKWHPGjUNbjohYS2MySNDBiCXA2tDs3jT6P3MXDgb9D" --loglevel info # QmSZGxdbG1AfuDX1ia6kRyewrVbBNkjd6NU4AbZUCE2ZyG
# fi

# QmRfp2xEWdRwagEfxVXTJee6xqaJnWTPogpKKdutq3FHNQ
if [ "$1" == "local" ]; then
go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI= -support_shards all -host "0.0.0.0" --loglevel debug
fi

if [ "$1" == "server" ]; then
# PUBLIC_IP=`dig -4 @resolver1.opendns.com ANY myip.opendns.com +short`;
go build -o highway && ./highway -privatekey CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI= -support_shards all -host "0.0.0.0"
fi

if [ "$1" == "hw1" ]; then
./highway -privatekey CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI= -support_shards all -host "139.162.9.169" -admin_port 8080 -proxy_port 7337 -bootnode_port 9330 --loglevel debug
fi

if [ "$1" == "hw2" ]; then
./highway -privatekey CAMSeTB3AgEBBCDg9L4rFdng09R48KDyAEDzAiD0ckpqLsZOFmj6JNNWwqAKBggqhkjOPQMBB6FEA0IABK/dfR6Y+BQMIoBvPka6XXPIkTKFuzZxFbSbrz1PZbMcgAE0fMEvYiu7IAJ0NWKYyzzsg+hOFEIKBK+avbyna+k= -support_shards all -host "139.162.9.169" -admin_port 8081 -proxy_port 7338 -bootnode_port 9331 --bootstrap "139.162.9.169:9330" --loglevel debug
fi

if [ "$1" == "hw3" ]; then
./highway -privatekey CAMSeTB3AgEBBCAykkTQRzJaBV81t58HSyt6DUSRS68kvr0bnH8IGSnaqaAKBggqhkjOPQMBB6FEA0IABPZuAn3egjeNwZEC2hSEfY8LaPKKt63UBaQ8eL5AwUy6PnZjwgsuDl0dLzxAqm8eU9NuuADxLS9mwucAsfxOE+4= -support_shards all -host "139.162.9.169" -admin_port 8082 -proxy_port 7339 -bootnode_port 9332 --bootstrap "139.162.9.169:9330" --loglevel debug
fi

if [ "$1" == "hw4" ]; then
./highway -privatekey CAMSeTB3AgEBBCCSIsG5e3ESqesSDqAcN3WIMUn9IoX9XTROixfYSQEMmKAKBggqhkjOPQMBB6FEA0IABDZNozvtHZvMkVVsks85svDPvvy0bw+/Zt7yCu7N66Cd1hchsR1CIe6U2q5E9uWke5Q+DZdjB5BPxq+ChnsDOeE= -support_shards all -host "139.162.9.169" -admin_port 8083 -proxy_port 7340 -bootnode_port 9333 --bootstrap "139.162.9.169:9330" --loglevel debug
fi

if [ "$1" == "lc1" ]; then
./highway -privatekey CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI= -support_shards all -host "0.0.0.0" -admin_port 8080 -proxy_port 7337 -bootnode_port 9330 --loglevel debug
fi

if [ "$1" == "lc2" ]; then
./highway -privatekey CAMSeTB3AgEBBCDg9L4rFdng09R48KDyAEDzAiD0ckpqLsZOFmj6JNNWwqAKBggqhkjOPQMBB6FEA0IABK/dfR6Y+BQMIoBvPka6XXPIkTKFuzZxFbSbrz1PZbMcgAE0fMEvYiu7IAJ0NWKYyzzsg+hOFEIKBK+avbyna+k= -support_shards all -host "0.0.0.0" -admin_port 8081 -proxy_port 7338 -bootnode_port 9331 --bootstrap "0.0.0.0:9330" --loglevel debug
fi

if [ "$1" == "lc3" ]; then
./highway -privatekey CAMSeTB3AgEBBCAykkTQRzJaBV81t58HSyt6DUSRS68kvr0bnH8IGSnaqaAKBggqhkjOPQMBB6FEA0IABPZuAn3egjeNwZEC2hSEfY8LaPKKt63UBaQ8eL5AwUy6PnZjwgsuDl0dLzxAqm8eU9NuuADxLS9mwucAsfxOE+4= -support_shards all -host "0.0.0.0" -admin_port 8082 -proxy_port 7339 -bootnode_port 9332 --bootstrap "0.0.0.0:9330" --loglevel debug
fi

if [ "$1" == "lc4" ]; then
./highway -privatekey CAMSeTB3AgEBBCCSIsG5e3ESqesSDqAcN3WIMUn9IoX9XTROixfYSQEMmKAKBggqhkjOPQMBB6FEA0IABDZNozvtHZvMkVVsks85svDPvvy0bw+/Zt7yCu7N66Cd1hchsR1CIe6U2q5E9uWke5Q+DZdjB5BPxq+ChnsDOeE= -support_shards all -host "0.0.0.0" -admin_port 8083 -proxy_port 7340 -bootnode_port 9333 --bootstrap "0.0.0.0:9330" --loglevel debug
fi
