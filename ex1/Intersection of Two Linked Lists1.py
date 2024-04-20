#リンクされたリストの定義
class ListNode:
    def __init__(self, x):
        self.val = x
        self.next = None

from typing import Optional  

class Solution:
    def getIntersectionNode(self, headA: ListNode, headB: ListNode) -> Optional[ListNode]:
        # ハッシュテーブルを作成
        nodes_in_b = set()
        
        # リストBのノードをハッシュテーブルに追加
        current = headB
        while current:
            nodes_in_b.add(current)
            current = current.next
        
        # リストAを走査し、リストBのノードが存在するかチェック
        current = headA
        while current:
            if current in nodes_in_b:
                return current  # 交差点を発見
            current = current.next
        
        return None  # 交差点なし