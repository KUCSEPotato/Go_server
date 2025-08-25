package main

/*
구조체는 필드의 집합
	Go는 객체 지향 언어가 가지는 클래스, 객체, 상속 개념이 없음
	일반적인 클래스가 필드, 메소드를 함께 가지는 것과 달리 Go의 구조체는 필드만 가질 수 있음
	메소드를 따로 선언해야 함.
Constructor: https://namu.wiki/w/%EC%83%9D%EC%84%B1%EC%9E%90
*/

import "fmt"

// define struct
type person struct {
	name string
	age  int
}
type dict struct {
	data map[int]string
}

// define constructor function
func newDict() *dict {
	d := dict{}
	d.data = map[int]string{}
	return &d // pointer 전달
}

func main() {
	// part of using person struct
	// create a new person
	p1 := person{name: "Alice", age: 30}
	fmt.Println(p1)

	// access fields
	fmt.Println("Name:", p1.name)
	fmt.Println("Age:", p1.age)

	// initialize a new person
	var p2 person
	p2.name = "Bob"
	p2.age = 25

	// initialize using new()
	// p3이 포인터라도 .을 사용한다
	p3 := new(person)
	p3.name = "Charlie"
	p3.age = 35
	fmt.Println(*p3)

	// part of using dict
	dic := newDict() // call constructor
	dic.data[1] = "one"
	fmt.Println(dic)
}
