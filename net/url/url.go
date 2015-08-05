package url

type URL struct {
	Schema   string
	Opaque   string
	User     *Userinfo
	Host     string
	Path     string
	RawQuery string
	Fragment string
}

type Userinfo struct {
	username    string
	password    string
	passwordSet bool
}

type Values map[string][]string //又一个一人千面
