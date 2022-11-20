#!/bin/sh
mkdir /etc/wireguard/"$1"
mkdir /etc/wireguard/"$1"/"$2"
wg genkey | tee /etc/wireguard/"$1"/"$2"/privatekey | wg pubkey | tee /etc/wireguard/"$1"/"$2"/publickey