package logger

import (
	"os"
	"testing"
)

var sl = NewLogger(os.Stdout)

func TestNew(t *testing.T) {
	sl.Info("您好", "这是一个测试")
}

func BenchmarkInfo(b *testing.B) {
	b.ResetTimer()
	sl.Info("您好", "这是一个测试")
}

func BenchmarkInfof(b *testing.B) {
	b.ResetTimer()
	sl.Infof("您好-%s", "hhh")
}
