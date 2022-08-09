FROM ubuntu:16.04

LABEL author="linj" 
LABEL module="externaldb/sync-convert"

RUN  apt-get update -y && apt-get install -y curl

WORKDIR /root
COPY sync_convert ci-start-sync-convert.sh ./
COPY externaldb_merge.toml  ./externaldb.toml
