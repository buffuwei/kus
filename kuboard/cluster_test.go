package kuboard

import "testing"

func TestClus(tt *testing.T) {
	s, err := Clusters()
	if err != nil {
		tt.Error(err)
	}
	tt.Log(s)
	tt.Log("ending")
}

func TestNs(tt *testing.T) {
	s, err := Ns("GZ-DEV")
	if err != nil {
		tt.Error(err)
	}
	tt.Log(s)
	tt.Log("ending")
}
