package main

import (
	"strings"
)

type leetcode struct{}

// 利用map没有重复的key。把所有值给到map，判断key值为target - nums[x] 的值存在时，说明值相等
// example: nums = []int{1, 4, 6, 8, 12, 54, 64, 12, 3, 2, 5, 7, 9, 10}, target = 8 ;
func (l leetcode) twoSum(nums []int, target int) []int {
	lenNums := len(nums)
	if lenNums < 2 {
		return []int{}
	}

	hashMap := make(map[int]int)
	for k, v := range nums {
		if p, ok := hashMap[target-v]; ok {
			return []int{p, k}
		}
		hashMap[v] = k
	}

	return []int{}
}

/**
 * Definition for singly-linked list.
 * type ListNode struct {
 *     Val int
 *     Next *ListNode
 * }
 */
//  l1 := &ListNode{Val: 2, Next: &ListNode{Val: 4, Next: &ListNode{Val: 3}}}
//  l2 := &ListNode{Val: 5, Next: &ListNode{Val: 6, Next: &ListNode{Val: 4}}}
// output: 7 0 8
// 遍历链表进行加法，分别取出个位数和十位数，个数放到Val，十位数参与下次加法。最后链表结束，十位数不为0时，指向此数。
func (l leetcode) addTwoNumbers(l1 *ListNode, l2 *ListNode) *ListNode {
	var tail, head *ListNode
	carry := 0

	for l1 != nil || l2 != nil {
		n1, n2 := 0, 0
		if l1 != nil {
			n1 = l1.Val
			l1 = l1.Next
		}
		if l2 != nil {
			n2 = l2.Val
			l2 = l2.Next
		}
		sum := n1 + n2 + carry
		sum, carry = sum%10, sum/10

		if head == nil {
			head = &ListNode{Val: sum}
			tail = head
			// head = head
		} else {
			head.Next = &ListNode{Val: sum}
			head = head.Next
		}

	}
	if carry > 0 {
		head.Next = &ListNode{Val: carry}
	}

	return tail
}

// 定义start, end的范围. 遍历s[i]，当i在范围内能找到下标cIndex时，-1未找到。
// 大于-1找到了i， 取max值，start后移1 + cIndex, 当i是最后一个值时，重新取max。
func (l leetcode) lengthOfLongestSubstring(s string) int {

	max := 0
	start := 0
	end := 0
	for i := 0; i < len(s); i++ {
		cIndex := strings.LastIndex(s[start:end], string(s[i]))
		if cIndex > -1 {
			if max < (end - start) {
				max = end - start
			}
			// if cIndex == 0 {
			// start++
			// } else {
			start += cIndex + 1
			// }
		}
		end++

		if i == len(s)-1 && max < end-start {
			max = end - start
		}
	}
	return max
}

// 先合并有序数组，后判断中位数是1个还是2个。速度慢，内存占用高，要改进。
func (l leetcode) findMedianSortedArrays(nums1 []int, nums2 []int) float64 {
	var fl64 float64
	sortList := []int{}
	m := len(nums1)
	n := len(nums2)
	var i, j int
	for i < m && j < n {
		if nums1[i] > nums2[j] {
			sortList = append(sortList, nums2[j])
			j++
		} else {
			sortList = append(sortList, nums1[i])
			i++
		}
	}

	if i < m {
		sortList = append(sortList, nums1[i:m]...)
	}
	if j < n {
		sortList = append(sortList, nums2[j:n]...)
	}

	q := len(sortList)
	// fmt.Println(sortList, q%2)
	if q%2 != 0 {
		fl64 = float64(sortList[q/2])
	} else {
		fl64 = (float64((sortList[q/2]) + sortList[q/2-1])) / 2
	}

	return fl64
}
