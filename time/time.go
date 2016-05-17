package time

// A time 是纳秒级别的
// 保存与传递使用值类型store and pass them as values, 不要用pointer
// Time实现了MarshalBinary()([]byte, error) MarshalJSON([]byte, error) MarshalText()([]byte, error)
// 2006 01-02 15:04:05 -0700

// Date() 用于构建一个time.Time, 必须要有Location
// format?
// 如何计算每个月的最后一天, 或者第一天, 或者第一周?
// Since返回一个Duration
// Location?
// timer?
// Ticker?

type Time struct {
	// 从公元1年1月1号00:00:00 UTC开始算起
	// IsZero()判断上述这个值
	// 为什么不设置为1970年1月1号呢?
	sec int64

	nsec int32

	loc *Location
}

// 判断t是在u之后?
// After是针对t来说的
func (t Time) After(u Time) bool {
	return t.sec > u.sec || t.sec == u.sec && t.nsec > u.nsec
}

func (t Time) Before(u Time) bool {
	return t.sec < u.sec || t.sec == u.sec && t.nsec < t.nsec
}

func (t Time) Equal(u Time) bool {
	return t.sec == u.sec && t.nsec == t.nsec
}

type Month int

const (
	January Month = 1 + iota
	February
	March
	April
	May
	June
	July
	August
)

type Weekday int

const (
	Sunday Weekday = iota // 0
)

func (t Time) IsZero() bool {
	return t.sec == 0 && t.nsec == 0
}

type Duration int64

func isLeap(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}
