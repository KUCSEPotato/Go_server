package main

/*

인터페이스는 메소드의 집합을 정의
구조체는 인터페이스를 구현하여 사용
*/

import (
	"fmt"
	"math"
)

// Shape interface
type Shape interface {
	area() float64
	perimeter() float64
}

// define Rect
type Rect struct {
	width, height float64
}

// define circle
type Circle struct {
	radius float64
}

// Rect type implements Shape interface
func (r Rect) area() float64 {
	return r.width * r.height
}

func (r Rect) perimeter() float64 {
	return 2 * (r.width + r.height)
}

// Circle type implements Shape interface
func (c Circle) area() float64 {
	return math.Pi * c.radius * c.radius
}

func (c Circle) perimeter() float64 {
	return 2 * math.Pi * c.radius
}

func showArea(shapes ...Shape) {
	for _, s := range shapes {
		a := s.area() // call interface method
		fmt.Println("Area:", a)

		b := s.perimeter() // call interface method
		fmt.Println("Perimeter:", b)
	}
}

func main() {
	r := Rect{width: 10, height: 20}
	c := Circle{radius: 5}
	showArea(r, c)
}
