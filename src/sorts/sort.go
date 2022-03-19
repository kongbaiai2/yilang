package sorts

type SortInt struct{}

// 一般建议待排序数组为小规模情况下使用 直接插入排序，
// 在规模中等的情况下可以使用希尔排序，
// 但在大规模还是要使用快速排序，归并排序或堆排序。

func (s SortInt) BubblingSort(listInt []int) []int {
	diffSwap := false
	for i := len(listInt) - 1; i > 0; i-- {
		for j := 0; j < i; j++ {
			if listInt[j] > listInt[j+1] {
				listInt[j], listInt[j+1] = listInt[j+1], listInt[j]
				diffSwap = true
			}
		}
		if !diffSwap {
			return listInt
		}
	}
	return listInt
}

func (s SortInt) SelectSort(list []int) []int {
	n := len(list)

	// 只需循环一半
	for i := 0; i < n/2; i++ {
		minIndex := i // 最小值下标
		maxIndex := i // 最大值下标

		// 在这一轮迭代中要找到最大值和最小值的下标
		for j := i + 1; j < n-i; j++ {
			// 找到最大值下标
			if list[j] > list[maxIndex] {
				maxIndex = j // 这一轮这个是大的，直接 continue
				// continue
			}
			// 找到最小值下标
			if list[j] < list[minIndex] {
				minIndex = j
			}
		}

		if maxIndex == i && minIndex != n-i-1 {
			// 如果最大值是开头的元素，而最小值不是最尾的元素
			// 先将最大值和最尾的元素交换
			list[n-i-1], list[maxIndex] = list[maxIndex], list[n-i-1]
			// 然后最小的元素放在最开头
			list[i], list[minIndex] = list[minIndex], list[i]
		} else if maxIndex == i && minIndex == n-i-1 {
			// 如果最大值在开头，最小值在结尾，直接交换
			list[minIndex], list[maxIndex] = list[maxIndex], list[minIndex]
		} else {
			// 否则先将最小值放在开头，再将最大值放在结尾
			list[i], list[minIndex] = list[minIndex], list[i]
			list[n-i-1], list[maxIndex] = list[maxIndex], list[n-i-1]
		}
	}
	return list
}

func (s SortInt) InsertSort(list []int) []int {
	// 取当前下标值和前一项对比，小于前一项，则当前下标的值改为前一项的值。并减标继续对比
	// 直到当前下标值大于或下标-1，将值赋给下标前一项.
	// 8 4 2 6 7 3
	/*
	 */
	for i := 1; i < len(list); i++ {
		j := i - 1
		deal := list[i]
		// if deal < list[j] {
		for ; j >= 0 && deal < list[j]; j-- {
			list[j+1] = list[j]
		}
		list[j+1] = deal
		// continue
		// }
	}
	return list
}

func (s SortInt) ShellSort(list []int) []int {
	n := len(list)
	// 每次减半，直到步长为 1
	// 按step / 2，进行递归，取step位上的值进行插入排序
	for step := n / 2; step > 0; step /= 2 {
		for i := step; i < n; i += step {
			deal := list[i]
			j := i - step
			for ; j >= 0 && deal < list[j]; j -= step {
				// if list[j+step] < list[j] {
				list[j+step] = list[j]
			}

			list[j+step] = deal
			// continue
		}
	}
	return list
}

// 归并排序，分治法。
func (s SortInt) MergeSort(arr []int) []int {

	length := len(arr)
	if length < 2 {
		return arr
	}
	middle := length / 2
	left := arr[0:middle]
	right := arr[middle:]
	return s.Merge(s.MergeSort(left), s.MergeSort(right))
}

func (s SortInt) Merge(left []int, right []int) []int {
	var result []int
	for len(left) != 0 && len(right) != 0 {
		if left[0] <= right[0] {
			result = append(result, left[0])
			left = left[1:]
		} else {
			result = append(result, right[0])
			right = right[1:]
		}
	}

	for len(left) != 0 {
		result = append(result, left[0])
		left = left[1:]
	}

	for len(right) != 0 {
		result = append(result, right[0])
		right = right[1:]
	}

	return result
}
