package main

import "fmt"

type ListNode struct {
	Val  int
	Next *ListNode
}

type leetcodeTest struct {
	leetcode
}

var lt leetcodeTest

func main() {

	// lt.leetcodeTest1()
	// lt.leetcodeTest2()
	lt.leetcodeTest3()

}

func (lt leetcodeTest) leetcodeTest1() {
	nums := []int{2, 5, 6, 2, 4, 2, 2, 8, 3, 11, 45, 34, 1, 7, 8, 9}
	target := 4
	fmt.Println(lt.twoSum(nums, target))
}
func (lt leetcodeTest) leetcodeTest2() {
	l1 := &ListNode{Val: 2, Next: &ListNode{Val: 4, Next: &ListNode{Val: 3}}}
	l2 := &ListNode{Val: 5, Next: &ListNode{Val: 6, Next: &ListNode{Val: 4}}}
	l3 := lt.addTwoNumbers(l1, l2)
	for l3 != nil {
		fmt.Println(l3.Val)
		l3 = l3.Next
	}
	// addTwoNumbers(l1, l2)

}

func (lt leetcodeTest) leetcodeTest3() {
	s := "abcabcbb"
	fmt.Println(lt.lengthOfLongestSubstring(s))
}
