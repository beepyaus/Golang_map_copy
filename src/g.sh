#!/bin/sh
#rm ./main
go build main.go & ./main $@
