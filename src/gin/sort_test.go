package main

import (
	"fmt"
	"sorts"
	"testing"
)

var (
	count2   int
	listInt  []int
	listInt2 []int
	listInt3 []int
	listInt4 []int
)

// func TestMain(m *testing.M) {
// 	count2 = 1

// 	count := 100000

// 	for i := 0; i < count; i++ {
// 		v := rand.Intn(count)
// 		listInt = append(listInt, v)
// 		listInt2 = append(listInt2, v)
// 		listInt3 = append(listInt3, v)
// 		listInt4 = append(listInt4, v)

// 	}
// 	// m.Run()
// }

func TestBubblingSort(t *testing.T) {

	// t.Log("2", listInt2)
	s := sorts.SortInt{}
	for i := 0; i < count2; i++ {
		s.BubblingSort(listInt)
		// s.SelectSort(listInt)
	}
	// t.Log("2", listInt2)
}
func TestSelectSort(t *testing.T) {
	// t.Log("1", listInt)
	s := sorts.SortInt{}
	for i := 0; i < count2; i++ {
		// s.BubblingSort(listInt)
		s.SelectSort(listInt2)
	}
	// t.Log("1", listInt)
}

func TestInsertSort(t *testing.T) {
	s := sorts.SortInt{}
	for i := 0; i < count2; i++ {
		// s.BubblingSort(listInt)
		s.InsertSort(listInt3)
	}
}
func TestShellSort(t *testing.T) {
	s := sorts.SortInt{}
	for i := 0; i < count2; i++ {
		// s.BubblingSort(listInt)
		s.ShellSort(listInt4)
	}
}

func Benchmark_Alloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fmt.Sprintf("i:%d", i)
	}
}

// go test -v -bench=Alloc -benchmem sort_test.go
// “16 B/op”表示每一次调用需要分配 16 个字节，“2 allocs/op”表示每一次调用有两次分配。
// -benchtime=5s 自定义测试时间, -benchmem 内存分配
func Benchmark_Add(b *testing.B) {
	var n int
	for i := 0; i < b.N; i++ {
		n++
	}
}
