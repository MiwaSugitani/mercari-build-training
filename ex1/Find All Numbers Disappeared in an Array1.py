class Solution(object):
    def findDisappearedNumbers(self, nums):
        n = len(nums)

        for i in range(1, n + 1):
            judge = 0
            for j in range(n):
                if nums[j] == i:
                    judge = 1
                    break
            if judge == 0:
                nums.append(i)

        return nums[n:]