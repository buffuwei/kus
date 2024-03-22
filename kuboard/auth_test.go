package kuboard

import (
	"testing"
)

func TestHost(tt *testing.T) {
	tt.Logf("host is : %s \n", Host())
}

func TestReqId(tt *testing.T) {
	tt.Logf("host is : %s \n", Host())
	tt.Log(reqId())
}

func TestNewAuthToken(tt *testing.T) {
	// reqId, _ := reqId()
	token, _ := NewToken("fuwei", "FU721312he!9")
	tt.Log(token)
}
