#!/bin/bash

./shitpool --worker 100 --chk-delay 2 --data shitpool.json mix \
    --apiaddr 0.0.0.0:3333 --apilog api.log \
    --httpaddr 0.0.0.0:4444 --socksaddr 0.0.0.0:5555 --pxylog proxy.log --pxytype all --pxyretry 5
