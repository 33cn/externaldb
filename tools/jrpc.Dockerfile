FROM ubuntu:16.04

LABEL author="linj" 
LABEL module="externaldb/jrpc"

WORKDIR /root
COPY jrpc externaldb.toml ./

