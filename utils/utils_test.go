package utils

import "testing"

func TestParse(t *testing.T) {
	z, x, y, ext, err := ParsePath("/tile/4/3/2.webp")
	if err != nil {
		t.Error(err)
	}
	if z != 4 || x != 3 || y != 2 || ext != "webp" {
		t.Fail()
	}
}

func TestParseError(t *testing.T) {
	invalid_paths := []string{
		"/tile/4/3/2",
		"/tile/4/3/2.jpg",
		"/tile/4/3/2.png/",
		"/tile/4/3/2.png/3",
		"/tile/4/3/2.png/3/4",
		"/tile/-1/3/2.png",
		"/tile/1.5/3/2.png",
		"/tile/10000/3.5/2.png",
		"/tile/100000000000000000000000000000000000000000000000000000000000000000/3/2.5.png",
		"/tile/abc/3/2.png",
	}
	for _, path := range invalid_paths {
		_, _, _, _, err := ParsePath(path)
		if err == nil {
			t.Errorf("expected error for path %s", path)
		}
	}
}
