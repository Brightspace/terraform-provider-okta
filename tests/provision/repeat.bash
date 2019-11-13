#!/bin/bash
counter=1
total=100
while [ $counter -le $total ]
do
    go run tests/provision/set/main.go "$@"
    go run tests/provision/revoke/main.go "$@"
    ((counter++))
done

echo "Done"