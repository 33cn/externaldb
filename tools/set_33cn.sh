#!/usr/bin/env bash

rm "$GOPATH"/src/github.com/33cn/externaldb/vendor/github.com/33cn/* -rf
cp "$GOPATH"/src/github.com/33cn/chain33 "$GOPATH"/src/github.com/33cn/plugin "$GOPATH"/src/github.com/33cn/externaldb/vendor/github.com/33cn -r
rm "$GOPATH"/src/github.com/33cn/externaldb/vendor/github.com/33cn/chain33/vendor -rf
rm "$GOPATH"/src/github.com/33cn/externaldb/vendor/github.com/33cn/plugin/vendor -rf
rm "$GOPATH"/src/github.com/33cn/externaldb/vendor/github.com/33cn/chain33/.git -rf
rm "$GOPATH"/src/github.com/33cn/externaldb/vendor/github.com/33cn/plugin/.git -rf
