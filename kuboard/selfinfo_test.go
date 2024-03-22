package kuboard

import "testing"

func TestSelfInfo(tt *testing.T) {
	n, e := GetSelfName()
	tt.Log(n, e)
}
