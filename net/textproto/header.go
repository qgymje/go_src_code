package textproto

type MIMEHeader map[string][]string //也就是说header里的key对应的value可能有多个值, 十分细化

func (h MIMEHeader) Add(key, value string) {
	key = CanonicalMIMEHeaderKey(key) //验证一下key是否合法
	h[key] = append(h[key], value)
}

func (h MIMEHeader) Set(key, value string) {
	h[CanonicalMIMEHeaderKey(key)] = []string{value}
}

func (h MIMEHeader) Get(key string) string {
	if h == nil {
		return ""
	}
	v := h[CanonicalMIMEHeaderKey(key)]
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

func (h MIMEHeader) Del(key string) {
	delete(h, CanonicalMIMEHeaderKey(key))
}
