package time

type Location struct {
	name string
	zone []zone
	tx   []zoneTrans

	cacheStart int64
	cacheEnd   int64
	cacheZone  *zone
}
