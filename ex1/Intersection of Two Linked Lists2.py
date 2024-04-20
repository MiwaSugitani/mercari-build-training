from typing import Optional

class ListNode:
    def __init__(self, x):
        self.val = x
        self.next = None

class Solution:
    def getIntersectionNode(self, headA: ListNode, headB: ListNode) -> Optional[ListNode]:
        if not headA or not headB:
            return None

        # リストの長さを取得するヘルパー関数
        def get_length(node):
            length = 0
            while node:
                length += 1
                node = node.next
            return length

        # 各リストの長さを取得
        lenA = get_length(headA)
        lenB = get_length(headB)

        # 長さの差を計算
        diff = abs(lenA - lenB)

        # 長いリストのヘッドを短いリストと同じ長さの位置に移動
        long_head = headA if lenA > lenB else headB
        short_head = headB if lenA > lenB else headA

        # 長いリストを差分だけ進める
        for _ in range(diff):
            long_head = long_head.next

        # 交差点を探す
        while long_head and short_head:
            if long_head == short_head:
                return long_head
            long_head = long_head.next
            short_head = short_head.next

        return None