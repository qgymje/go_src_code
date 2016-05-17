package cookiejar

type PublicSuffixList interface {
	PublicSuffix(domain string) string

	String() string
}
