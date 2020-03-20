#!/bin/bash

test_cat() {
    diff <(printf "$1" | go run main.go cat 2>/dev/null) <(printf "$2") || 
        { echo "Failed for input $1"; exit 1; }
}

test_cat "" ""
test_cat "\n" ""
test_cat "\n\n" ""
test_cat "\r" ""
test_cat "\r\n" ""
test_cat "1" "1\n"
test_cat "1\n" "1\n"
test_cat "1\r" "1\n"
test_cat "1\r\n" "1\n"
test_cat "1\n2\n3\n" "1\n2\n3\n"
test_cat "1\n\n2\n3\n" "1\n2\n3\n"
