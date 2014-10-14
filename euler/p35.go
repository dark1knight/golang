/*
The number, 197, is called a circular prime because all rotations of the digits: 197, 971, and 719, are themselves prime.

There are thirteen such primes below 100: 2, 3, 5, 7, 11, 13, 17, 31, 37, 71, 73, 79, and 97.

How many circular primes are there below one million?
*/
package main

import (
	"bytes"
	"fmt"
	"github.com/dark1knight/newmath"
	"strconv"
)

const (
	NumIters = 1000000
	Base     = 10
	BitSize  = 64
)

func rotateLeft(num_str *string, strlen int) string {
	var buffer bytes.Buffer // create a bytes buffer
	i := 0
	// move everything over to the left
	for ; i < strlen-1; i++ {
		buffer.WriteByte((*num_str)[i+1])
	}
	// move first character to the end
	buffer.WriteByte((*num_str)[0])
	return buffer.String()
}

func isCircularPrime(num int64) bool {
	var num_str string = strconv.FormatInt(num, Base)
	var strlen int = len(num_str)
	for i := 0; i < strlen; i++ {
		if num, err := strconv.ParseInt(num_str, Base, BitSize); err != nil {
			return false
		} else {
			if !newmath.IsPrime(num) {
				return false
			}
		}
		num_str = rotateLeft(&num_str, strlen)
	}
	return true
}

func main() {
	var total int = 0
	for i := int64(0); i < NumIters; i++ {
		if isCircularPrime(i) {
			total += 1
		}
	}
	fmt.Printf("%v\n", total)
}
