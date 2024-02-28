class Solution(object):
    def findDisappearedNumbers(self, nums):
        n = len(nums)
        l = []

        for i in range(1,n+1):
            for j in range(n):
                if nums[j] == i:
                    break
            else:
                l.append(i)

        return l