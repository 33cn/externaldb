FROM ubuntu:16.04

LABEL author="linj" 
LABEL module="externaldb/convert"

RUN  apt-get update -y && apt-get install -y curl

WORKDIR /root
COPY convert externaldb.toml ci-start-convert.sh ./

