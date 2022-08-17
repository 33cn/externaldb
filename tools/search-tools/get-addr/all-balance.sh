#!/usr/bin/env bash
xargs -n 1 -I {} ./one-balance.sh {} >data/last-balance <data/to_addresses
