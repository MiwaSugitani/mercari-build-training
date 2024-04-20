class Solution(object):
    def findDisappearedNumbers(self, nums):
        n = len(nums)
        l = []
        for i in range(1,n+1):
            if i not in nums:
                l.append(i)
        
        return l