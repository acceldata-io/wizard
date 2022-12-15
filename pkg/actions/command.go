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
	"context"
	"fmt"
	"time"

	"github.com/acceldata-io/wizard/internal/parser"
	"github.com/acceldata-io/wizard/pkg/register"
	"github.com/acceldata-io/wizard/pkg/wlog"

	command "github.com/acceldata-io/goutils/shellutils/cmd"
)

type cmd struct {
	timeout  int
	register string
}

func NewCmdAction(timeout int, localRegister string) Action {
	return &cmd{timeout: timeout, register: localRegister}
}

func (s *cmd) Do(actions *parser.Action, wizardLog chan interface{}) error {
	sRegister := register.RMap[s.register]

	if len(actions.Command) < 1 {
		wizardLog <- wlog.WLError("wrong command found, length of the command is less than 1")
		return fmt.Errorf("wrong command found")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.timeout)*time.Second)
	defer cancel()

	execCmd := command.New(ctx, actions.Command[0], actions.Command[1:])
	wizardLog <- wlog.WLInfo("running command: " + execCmd.Command)
	cmd, err := execCmd.Run()
	if err != nil {
		wizardLog <- wlog.WLError(fmt.Sprintf("unable to execute the command: %q. Because: %s", actions.Command, err.Error()))
		return fmt.Errorf("unable to execute the command: %q. Because: %s", actions.Command, err.Error())
	}

	status := cmd.Status
	sRegister.StdOut = status.StdOut
	sRegister.ExitCode = status.ExitCode
	if status.ExitCode != int(actions.ExitCode) {
		wizardLog <- wlog.WLError(fmt.Sprintf("exit code not matched, expected: %d, Got: %d", int(actions.ExitCode), status.ExitCode))
		return fmt.Errorf("exit code not matched. expected: %d, Got: %d, stderr: %s", int(actions.ExitCode), status.ExitCode, status.StdErr)
	}
	return nil
}
