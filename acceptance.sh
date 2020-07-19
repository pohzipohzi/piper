#!/bin/bash

test_cat() {
    diff <(printf "$1" | go run main.go -c cat) <(printf "$2") || 
        { echo "Failed for input $1"; exit 1; }
}

for input in "" "\n" "\n\n" "\r\n" "\r\n\r\n"; do
	test_cat "$input" ""
done
for input in "1" "1\n" "1\n\n" "1\r\n" "1\r\n\r\n" "\n1" "\n\n1" "\r\n1" "\r\n\r\n1"; do
	test_cat "$input" "(input)\n1\n(output)\n1\n\n"
done 
for input in "1\n2\n3" "1\n2\n3\n" "1\r\n2\r\n3" "1\r\n2\r\n3\r\n"; do
	test_cat "$input" "(input)\n1\n2\n3\n(output)\n1\n2\n3\n\n"
done
