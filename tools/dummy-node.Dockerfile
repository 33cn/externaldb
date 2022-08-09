FROM ubuntu:16.04

LABEL author="linj" 
LABEL module="externaldb/dummy-node"

WORKDIR /root
COPY ./dummy_node ./
