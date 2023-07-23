package utils

import "testing"

func TestParse(t *testing.T) {
	z, x, y, err := utils.ParsePath("/tile/4/3/2.png")
	if err != nil {
		t.Error(err)
	}
	if z != 4 || x != 3 || y != 2 {
		t.Fail()
	}
}
