package main

import "fmt"

func main() {
	arr := [3]int{1, 2, 3}
	s := make([]int, 1)
	copy(s, arr[0:2])
	s[0] = 99

	fmt.Println(arr)
}
