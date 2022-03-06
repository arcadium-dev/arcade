#!/bin/sh 
/migrate -path /migrations -database ${SQL_URL} "$@"
