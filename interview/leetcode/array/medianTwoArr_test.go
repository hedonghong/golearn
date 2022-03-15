package array

import (
	"fmt"
	"testing"
)

/*
[](https://blog.csdn.net/bjweimengshu/article/details/97717144)
假设两个数组混合后，A为绿色数组，B为橙色数组

假设数组A的长度是m，绿色和橙色元素的分界点是i，数组B的长度是n，绿色和橙色元素的分界点是j，那么为了让大数组的左右两部分长度相等，则i和j需要符合如下两个条件：

i + j = （m+n+1）/2

（之所以m+n后面要再加1，是为了应对大数组长度为奇数的情况）

Max(A[i-1],B[j-1]) < Min(A[i], B[j])

(直白的说，就是最大的绿色元素小于最小的橙色元素)

第一步，就像二分查找那样，把i设在数组A的正中位置，也就是让i=3
第二步，根据i的值来确定j的值，j=(m+n+1)/2 - i =3
第三步，验证i和j，分为下面三种情况：

重复步骤：

1.B[j−1]≤A[i] && A[i−1]≤B[j]

说明i和j左侧的元素都小于右侧，这一组i和j是我们想要的。

2.A[i]<B[j−1]

说明i对应的元素偏小了，i应该向右侧移动。

3.A[i−1]>B[j]

说明i-1对应的元素偏大了，i应该向左侧移动。





如果大数组长度是奇数，那么：

中位数 = Max(A[i-1],B[j-1])

(也就是大数组左半部分的最大值)

如果大数组长度是偶数，那么：

中位数 = （Max(A[i-1],B[j-1]) + Min（A[i], B[i]））/2

（也就是大数组左半部分的最大值和大数组右半部分的最小值取平均）


问题：
1.数组A的长度远大于数组B
也就是m远大于n，这时候会出现什么问题呢？
当我们设定了i的初值，也就是数组A正中间的元素，再计算j的时候有可能发生数组越界。
因此，我们可以提前把数组A和B进行交换，较短的数组放在前面，i从较短的数组中取。
这样做还有一个好处，由于数组A是较短数组，i的搜索次数减少了。

2.数组A的所有元素都小于数组B，或数组A的所有元素都大于数组B

这种情况下，最终确定的i值等于0，或最终确定的i值等于0。
如果按照Max(A[i-1],B[j-1])的公式来求中位数，就会出现下标为负数的情况。
此时求中位数的公式就简化为A[i-1]或B[i-1]（假设大数组长度为奇数）

*/
func findMedianSortedArrays(nums1 []int, nums2 []int) float64 {
	// 假设 nums1 的长度小
	if len(nums1) > len(nums2) {
		return findMedianSortedArrays(nums2, nums1)
	}
	low, high, k, i, j := 0, len(nums1), (len(nums1)+len(nums2)+1)>>1, 0, 0
	for low <= high {
		// i = m/2 => low + (high-low) >> 1 防止i一直不变，计算出low, high之间的一个数，并且是动态的
		i = low + (high-low)>>1 // 分界限右侧是 mid，分界线左侧是 mid - 1
		j = k - i
		if i > 0 && nums1[i-1] > nums2[j] { // nums1 中的分界线划多了，要向左边移动
			high = i - 1
		} else if i != len(nums1) && nums1[i] < nums2[j-1] { // nums1 中的分界线划少了，要向右边移动
			low = i + 1
		} else {
			// 找到合适的划分了，需要输出最终结果了
			// 分为奇数偶数 2 种情况
			break
		}
	}
	// j = 0
	// 1 2 3
	// 6 7 8 9 10
	// i = 0
	// 8 9 10
	// 1 2 3 4 5 6
	// 数组A的所有元素都小于数组B，或数组A的所有元素都大于数组B
	// 特殊情况处理，想象下把 b 数组固定下来，a数组放在b的前面还是后面的
	midLeft, midRight := 0, 0
	// 交叉取数 i-1 i          j-1 j
	// 又知B[j−1]≤A[i] && A[i−1]≤B[j]
	// A[i-1] < A[i]  B[j−1] < B[j]
	// ...a[i-1] | a[i]...
	// ...b[j-1] | b[j]...
	if i == 0 { //B[j−1]≤A[i]
		midLeft = nums2[j-1] //其实b数组的更小
	} else if j == 0 { // A[i−1]≤B[j]
		midLeft = nums1[i-1] //其实a组的更小
	} else {
		midLeft = max(nums1[i-1], nums2[j-1])
	}
	if (len(nums1)+len(nums2))&1 == 1 {
		return float64(midLeft)
	}
	if i == len(nums1) { // A[i−1]≤B[j]
		midRight = nums2[j] //其实b数组的更大
	} else if j == len(nums2) { //B[j−1]≤A[i]
		midRight = nums1[i] //其实a数组的更大
	} else {
		midRight = min(nums1[i], nums2[j])
	}
	return float64(midLeft+midRight) / 2
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func min(a, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}

func TestMedianTwoArr(t *testing.T) {

	nums1 := []int{1, 3, 7, 10, 15}         //3
	nums2 := []int{2, 6, 9, 13, 16, 17, 20} //3
	// 1,2,3,6,7, 9,10, 13,15,16,17,20
	// 9+10 = 19 / 2 = 9.xx
	fmt.Println(findMedianSortedArrays(nums1, nums2))
}
