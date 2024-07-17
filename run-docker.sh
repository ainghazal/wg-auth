#!/bin/sh
docker run -p 51820:51820/udp \
        --cap-add NET_ADMIN \
        --cap-add SYS_MODULE \
        --sysctl="net.ipv4.conf.all.src_valid_mark=1" \
        --sysctl="net.ipv4.ip_forward=1" --rm -it \
        wgauth
