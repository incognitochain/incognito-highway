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
./highway -privatekey CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI= -support_shards all -host "0.0.0.0" -admin_port 8080 -proxy_port 7337 -bootnode_port 9330 --bootstrap "0.0.0.0:9330" --loglevel debug --gdburl ${GDBURL} --version "0.1-local"
fi

if [ "$1" == "lc2" ]; then
./highway -privatekey CAMSeTB3AgEBBCDg9L4rFdng09R48KDyAEDzAiD0ckpqLsZOFmj6JNNWwqAKBggqhkjOPQMBB6FEA0IABK/dfR6Y+BQMIoBvPka6XXPIkTKFuzZxFbSbrz1PZbMcgAE0fMEvYiu7IAJ0NWKYyzzsg+hOFEIKBK+avbyna+k= -support_shards all -host "0.0.0.0" -admin_port 8081 -proxy_port 7338 -bootnode_port 9331 --bootstrap "0.0.0.0:9330" --loglevel debug --gdburl ${GDBURL} --version "0.1-local"
fi

if [ "$1" == "lc3" ]; then
./highway -privatekey CAMSeTB3AgEBBCAykkTQRzJaBV81t58HSyt6DUSRS68kvr0bnH8IGSnaqaAKBggqhkjOPQMBB6FEA0IABPZuAn3egjeNwZEC2hSEfY8LaPKKt63UBaQ8eL5AwUy6PnZjwgsuDl0dLzxAqm8eU9NuuADxLS9mwucAsfxOE+4= -support_shards all -host "0.0.0.0" -admin_port 8082 -proxy_port 7339 -bootnode_port 9332 --bootstrap "0.0.0.0:9330" --loglevel debug --gdburl ${GDBURL} --version "0.1-local"
fi

if [ "$1" == "lc4" ]; then
./highway -privatekey CAMSeTB3AgEBBCCSIsG5e3ESqesSDqAcN3WIMUn9IoX9XTROixfYSQEMmKAKBggqhkjOPQMBB6FEA0IABDZNozvtHZvMkVVsks85svDPvvy0bw+/Zt7yCu7N66Cd1hchsR1CIe6U2q5E9uWke5Q+DZdjB5BPxq+ChnsDOeE= -support_shards all -host "0.0.0.0" -admin_port 8083 -proxy_port 7340 -bootnode_port 9333 --bootstrap "0.0.0.0:9330" --loglevel debug --gdburl ${GDBURL} --version "0.1-local"
fi

if [ "$1" == "lc5" ]; then
./highway -privatekey CAMSeTB3AgEBBCDkUqzZ6d9FPTjaLfiCBg0I1erJ0PbX7kuuprAjoJsXvaAKBggqhkjOPQMBB6FEA0IABLMPt8du3AAG1swxlAJnvt05X9ri23yxB4KpRnpB9oHsr+tD8INDEnTC+21XlJyzlCmlpzHtg9Zrj3rHfqtS/GU= -support_shards all -host "0.0.0.0" -admin_port 8084 -proxy_port 7341 -bootnode_port 9334 --bootstrap "0.0.0.0:9330" --loglevel debug --gdburl ${GDBURL} --version "0.1-local"
fi

if [ "$1" == "lc6" ]; then
./highway -privatekey CAMSeTB3AgEBBCCdgGhuRb9pmgOjM0URcxfiBXgYA+42L2WSkDNQgyBegaAKBggqhkjOPQMBB6FEA0IABIZd6b7Gj9Li/xA9yzMVBmrDrej3m5AurkHpw0LN81Gmxro63DiWw6mGZabfA/FnW6+SP28+pZGqggp1A+GRwaY= -support_shards all -host "0.0.0.0" -admin_port 8085 -proxy_port 7342 -bootnode_port 9335 --bootstrap "0.0.0.0:9330" --loglevel debug --gdburl ${GDBURL} --version "0.1-local"
fi

if [ "$1" == "lc7" ]; then
./highway -privatekey CAMSeTB3AgEBBCAIb1+c/8k9kKyLGiKXMzFnZ8D6HPKH/T2PpLqWZ3DO+qAKBggqhkjOPQMBB6FEA0IABGa7HeVeM+qCvpLC33Qd3MMh4VUOkjhFkV4ry3YMemVzUHeIGB+pg+8NjGhs9iLz1ti76EBLGlyOxOozkZiVR38= -support_shards all -host "0.0.0.0" -admin_port 8086 -proxy_port 7343 -bootnode_port 9336 --bootstrap "0.0.0.0:9330" --loglevel debug --gdburl ${GDBURL} --version "0.1-local"
fi

if [ "$1" == "lc8" ]; then
./highway -privatekey CAMSeTB3AgEBBCB3/STCrATElSzpBR2MmUKk8q09xEnXislZUot/3cCgY6AKBggqhkjOPQMBB6FEA0IABHFjdQdWUWJZHF8qJT9DabY+ZDTdmEDGvf7Emjg7IijRS1c/8NO82NuoGP9a2C0E73oK/NQGMsQrzqSs9QqZoFk= -support_shards all -host "0.0.0.0" -admin_port 8087 -proxy_port 7344 -bootnode_port 9337 --bootstrap "0.0.0.0:9330" --loglevel debug --gdburl ${GDBURL} --version "0.1-local"
fi

if [ "$1" == "newlc1" ]; then
./highway -privateseed YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE= -index 0 -support_shards all  -host "0.0.0.0"  -admin_port 8080 -proxy_port 7337 -bootnode_port 9330  --loglevel debug --gdburl ${GDBURL} --version "0.1-local"
fi

if [ "$1" == "newlc2" ]; then
./highway -privateseed YWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWFhYWE= -index 1 -support_shards all -host "0.0.0.0" -admin_port 8081 -proxy_port 7338 -bootnode_port 9336  --bootstrap "0.0.0.0:9335" --loglevel debug --gdburl ${GDBURL} --version "0.1-local"
fi

if [ "$1" == "newhw1" ]; then
./highway -privateseed CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI= --index 0 -support_shards all -host "139.162.9.169" -admin_port 8080 -proxy_port 7337 -bootnode_port 9330  --loglevel debug --gdburl ${GDBURL} --version "0.1-devnet"
fi

if [ "$1" == "newhw2" ]; then
./highway -privateseed CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxbtxIU3Z5oIVVOsYvI= -support_shards all -host "139.162.9.169"  -admin_port 8081 -proxy_port 7338 -bootnode_port 9331  --bootstrap "139.162.9.169:9330" --index 1 --loglevel debug --gdburl ${GDBURL} --version "0.1-devnet"
fi

if [ "$1" == "newlc3" ]; then
./highway -privateseed CAMSeTB3AgEBBCDtIHJcnRKCWVtitn0gkRTHlKvJCvSO12XVtzHna3oSEqAKBggqhkjOPQMBB6FEA0IABKQXV3mHcxNSmL3n4mtWTO4vNP2IuPvizYngBGxf6Fx9cCJhYUYH8r+Plp40dVcz53iXFxb -support_shards all -host "0.0.0.0" -admin_port 8082 -proxy_port 7339 -bootnode_port 9332 --bootstrap "0.0.0.0:9330" --index 2 --loglevel debug --gdburl ${GDBURL} --version "0.1-local"
fi
