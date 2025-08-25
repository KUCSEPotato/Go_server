package main

/*
Go는 다른 객체 지향언어와 달리 클래스는 존재하지 않음.
구조체는 필드만을 가짐
메소드가 구조체의 메소드로 정의할 수 있음
*/

import "fmt"

// Rect - struct 정의
type Rect struct {
	width, height int
}

// RectMethod area() 메소드 정의
func (r Rect) area() int {
	return r.width * r.height
}

func main() {
	rect := Rect{width: 10, height: 20}
	fmt.Println("Area of rectangle:", rect.area()) // area 메소드 호출
}

// pointer Receiver
func (r *Rect) area2() int {
	return r.width * r.height
}

// value, pointer can be used for method parameters
/*
func test() {
	rect2 := Rect{width: 10, height: 20}
	fmt.Println("Area of rectangle:", rect2.area2()) // area2 메소드 호출
}
*/
