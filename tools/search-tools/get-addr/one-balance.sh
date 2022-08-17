#!/usr/bin/env bash
~/guodun/chain33-cli --rpc_laddr http://192.168.3.8:8801 account balance -a "$1" -e coins | jq '.addr, .balance, .frozen, "coins"' | xargs
~/guodun/chain33-cli --rpc_laddr http://192.168.3.8:8801 account balance -a "$1" -e ticket | jq '.addr, .balance, .frozen, "ticket"' | xargs
