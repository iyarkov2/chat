#!/bin/sh

# go get github.com/actgardner/gogen-avro/v10/cmd/gogen-avro
rm -f generated/*

gogen-avro -containers -package generated generated schema.avsc