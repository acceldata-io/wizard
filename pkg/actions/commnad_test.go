// Acceldata Inc. and its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// 	Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
