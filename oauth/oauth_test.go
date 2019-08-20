package oauth

import (
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	//model.InitView("redis://127.0.0.1:6379/1","root:@tcp(127.0.0.1)/ucenter",nil)
	StartServer("0.0.0.0:9090", "00000000", "00000001", time.Minute)
}
