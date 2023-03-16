#!/bin/sh

cd ../frontend
./bindata.sh
cd ../example
go run main.go