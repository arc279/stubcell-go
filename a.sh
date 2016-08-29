#!/bin/bash

curl -k -s $* -XGET  https://localhost:1323/hello/bar

curl -k -s $* -XPOST https://localhost:1323/hello -d "name=foo"

curl -k -s $* -XGET  https://localhost:1323/
