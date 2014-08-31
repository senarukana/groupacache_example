#!/bin/sh
cd protocol && go build
cd ../client && go build
cd ../dbclient && go build
cd ../dbserver && go build
cd ../cacheserver && go build
