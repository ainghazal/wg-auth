#!/bin/ash

set -e

/bin/wgauth server > /etc/wireguard/wg0.conf

default_route_ip=$(ip route | grep default | awk '{print $3}')
if [[ -z "$default_route_ip" ]]; then
	echo "No default route configured" >&2
	exit 1
fi

configs=`find /etc/wireguard -type f -printf "%f\n"`
if [[ -z "$configs" ]]; then
	echo "No configuration file found in /etc/wireguard" >&2
	exit 1
fi

config=`echo $configs | head -n 1`
interface="${config%.*}"

# The net.ipv4.conf.all.src_valid_mark sysctl is set when running the container, so don't have WireGuard also set it
sed -i "s:sysctl -q net.ipv4.conf.all.src_valid_mark=1:echo Skipping setting net.ipv4.conf.all.src_valid_mark:" /usr/bin/wg-quick

# Start WireGuard
wg-quick up $interface

shutdown () {
	wg-quick down $interface
	exit 0
}

trap shutdown SIGTERM SIGINT SIGQUIT

watch wg
