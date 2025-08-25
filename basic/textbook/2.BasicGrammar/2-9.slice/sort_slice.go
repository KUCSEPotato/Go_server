package main

import (
	"fmt"
	"sort"
)

type myDataType struct {
	name string
	age  int
}

func main() {
	mySlice := make([]myDataType, 0)
	mySlice = append(mySlice, myDataType{"김형준", 42})
	mySlice = append(mySlice, myDataType{"홍길동", 28})
	mySlice = append(mySlice, myDataType{"임꺽정", 38})
	fmt.Println(mySlice)
	sort.Slice(mySlice, func(i, j int) bool {
		return mySlice[i].age < mySlice[j].age
	})
	fmt.Println(mySlice)
}
