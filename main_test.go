package main

import (
	"fmt"
	"strconv"
	"testing"
)

func TestRand(t *testing.T) {
	fmt.Println(strconv.FormatInt(1 << 6 -1, 2))
}
