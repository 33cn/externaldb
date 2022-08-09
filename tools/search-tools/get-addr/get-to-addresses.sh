#!/usr/bin/env bash
xargs -n 1 -I {} ./coins-tx-addr.sh {} >data/to_addresses 2>/dev/null <data/keys
