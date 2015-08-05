package strings_test

import (
	"os"
	"strings"
	"testing"
)

func TestReader(t *testing.T) {
	r := strings.NewReader("0123456789")
	tests := []struct {
		off     int64
		seek    int
		n       int
		want    string
		wantpos int64
		seekerr string
	}{
		{seek: os.SEEK_SET, off: 0, n: 20, want: "0123456789"},
	}
}
