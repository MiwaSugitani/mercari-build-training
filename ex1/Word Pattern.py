def s_pattern(s, pattern):
    s_list = s.split() 
    l = list(pattern)
    dict = {}

    #patternが辞書になければ追加していく
    for i in range(len(l)):
        if l[i] not in dict:
            dict[l[i]] = s_list[i]

    if len(pattern) != len(s_list):
        return False

    #辞書と入力した文字列＆パターンが合うか確認
    for i in range(len(pattern)):
        if dict[pattern[i]] != s_list[i]:
            return False

    return True


