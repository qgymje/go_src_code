package http

type Header map[string][]string // Header是map的别名, 但是有map的可操作的方法, 比如v,ok := map[key], delete(map, key), 这里体现了"一人千面"的个性特征
