package io

import (
	"bytes"
	"testing"

)

func Test_teeReader(t *testing.T) {
	r := bytes.NewReader([]byte("hello world"))
	var buf []byte
	n, err := r.Read(buf)
	t.Logf("buf =%s, n = %d, err = %v", buf,n, err)
	/*
	w := &bytes.Buffer{}
	tr := TeeReader(r, w)
	n, err := tr.Read(buf)
	t.Logf("n = %d, err = %v, buf = %s, w = %s", n, err, buf, w.String())
	*/
}
