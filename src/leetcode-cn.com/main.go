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
	// lt.leetcodeTest3()
	lt.leetcodeTest4()

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
func (lt leetcodeTest) leetcodeTest4() {
	nums1 := []int{1, 3, 9, 14, 17, 19, 21}
	nums2 := []int{2, 6, 8, 9, 10, 12}
	fmt.Println(lt.findMedianSortedArrays(nums1, nums2))
}

// 1, 2, 3, 6, 8, 9, 9 ,11, 14, 17, 19 21 22 25
