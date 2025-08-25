package main

import (
	"fmt"
)

type Obj struct {
	Name string
	Age  int
}

func PrintObject(list []Obj) {
	for _, obj := range list {
		fmt.Printf("Name: %s, Age: %d\n", obj.Name, obj.Age)
	}
}

func main() {
	objs := []Obj{
		{"Alice", 30},
		{"Bob", 25},
		{"Charlie", 35},
	}

	for _, object := range objs {
		object.Age = object.Age * 2
	}

	// 출력: 30, 25, 35
	PrintObject(objs)

	for idx := range objs {
		object := &objs[idx]
		object.Age = object.Age * 2
	}

	// 출력: 60, 50, 70
	PrintObject(objs)
}

/*
왜 첫 번째 루프는 값이 안 바뀌나?
for _, object := range objs {
    object.Age = object.Age * 2
}


object는 objs의 각 요소를 복사한 값(struct)이라, 수정해도 원본 슬라이스는 그대로야.

그래서 이 루프 뒤 첫 출력은 30, 25, 35가 맞아.

두 번째 루프는 왜 바뀌나?
for idx := range objs {
    object := &objs[idx]   // 슬라이스 ‘원소의 주소’를 잡음
    object.Age = object.Age * 2
}


이번엔 원소의 주소(&objs[idx]) 를 잡아 원본을 직접 수정하니까 값이 두 배로 바뀐다.

최종 출력은 60, 50, 70이 맞아.


실무 팁 (실수 방지)

값 복사 주의

for _, v := range objs { v.Age *= 2 } → 원본 안 바뀜(v는 복사본)

원본을 바꾸고 싶으면 인덱스 사용

for i := range objs {
    objs[i].Age *= 2
}


포인터 슬라이스를 돌릴 수도 있음

ps := []*Obj{{"Alice",30},{"Bob",25},{"Charlie",35}}
for _, p := range ps {
    p.Age *= 2   // 포인터라 원본 수정
}


반드시 피해야 할 패턴

for _, obj := range objs {
    // &obj 를 어딘가에 저장하면 안 됨!
    // obj는 매 반복마다 ‘같은’ 루프 변수의 주소가 재사용됨
}

*/
