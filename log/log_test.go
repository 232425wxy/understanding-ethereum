package log

import (
	"fmt"
	"os"
	"testing"
)

func TestRootNew(t *testing.T) {
	l1 := New("app", "ethereum/coin")
	l1.SetHandler(StreamHandler(os.Stdout, TerminalFormat(true)))
	l1.Debug("start service")

	l2 := New("consensus", "POS")
	fmt.Println(l2.GetHandler())
	l2.SetHandler(StreamHandler(os.Stdout, TerminalFormat(true)))
	l2.Debug("start consensus")
}

func TestExample(t *testing.T) {
	l := New("blockchain", "ethereum")
	file, _ := os.OpenFile("text.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0664)
	l.SetHandler(StreamHandler(file, TerminalFormat(false)))
	l.Info("start service")
	l.Error("start service")

	l.SetHandler(LvlFilterHandler(LvlWarn, StreamHandler(os.Stdout, TerminalFormat(true))))
	l.Info("info logger")
	l.Warn("warn logger")
	l.Error("error logger")
}
