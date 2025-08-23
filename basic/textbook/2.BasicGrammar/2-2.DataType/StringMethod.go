package main

import (
	"fmt"
	"strings"
)

func main() {
	str1 := []string{"apple", "banana", "grape", "orange", "mango"}
	fmt.Println(strings.Join(str1, ":"))

	str2 := "a.b.c"
	r := strings.Replace(str2, ".", "-", -1)
	fmt.Println(r)
}
