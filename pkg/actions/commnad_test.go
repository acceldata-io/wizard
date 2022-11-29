package actions

import (
	"testing"

	"github.com/acceldata-io/wizard/internal/parser"
	"github.com/acceldata-io/wizard/pkg/register"
)

func TestCommandLengthFail(t *testing.T) {
	register.RMap["test"] = &register.Register{}
	command := NewCmdAction(10, "test")

	var err error
	wLog := make(chan interface{})
	go func() {
		err = command.Do(&parser.Action{
			Action:   "cmd",
			Name:     "no command",
			Command:  nil,
			Register: register.GetHash("test"),
		}, wLog)
		close(wLog)
	}()

	// This is here to wait for the channel
	for range wLog {
	}

	if err.Error() != "wrong command found" {
		t.Fail()
	}
}

func TestCommandExitCodeMatch(t *testing.T) {
	register.RMap["test"] = &register.Register{}
	command := NewCmdAction(10, "test")

	var err error
	wLog := make(chan interface{})
	go func() {
		err = command.Do(&parser.Action{
			Action:   "cmd",
			Name:     "ls",
			Command:  []string{"ls", "-lha"},
			ExitCode: 0,
			Register: register.GetHash("test"),
		}, wLog)
		close(wLog)
	}()

	// This is here to wait for the channel
	for range wLog {
	}

	if err != nil {
		t.Fail()
	}
}

func TestCommandExitCodeNotMatch(t *testing.T) {
	register.RMap["test"] = &register.Register{}
	command := NewCmdAction(10, "test")
	var err error
	wLog := make(chan interface{})
	go func() {
		err = command.Do(&parser.Action{
			Action:   "cmd",
			Name:     "ls",
			Command:  []string{"ls", "-lha"},
			ExitCode: 1,
			Register: register.GetHash("test"),
		}, wLog)
		close(wLog)
	}()

	// This is here to wait for the channel
	for range wLog {
	}

	if err == nil {
		t.Fail()
	}
}
