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

	"github.com/acceldata-io/wizard/pkg/register"

	command "github.com/acceldata-io/goutils/shellutils/cmd"
)

type whenConditional struct {
	command  string
	rvar     string
	exitCode int
	timeout  int
}

func NewWhen(expression string, rvar string, expected int, timeout int) *whenConditional {
	return &whenConditional{
		command:  expression,
		exitCode: expected,
		timeout:  timeout,
		rvar:     rvar,
	}
}

func (w *whenConditional) Execute() (bool, error) {
	if w.rvar != "" {
		/*
				- Parse the input rvar
				- rvar contains registers variables
				- There can be logical operators with multiple registers (and, or, not, eq)
				- Comparison should happen with respective type
			 	- stdout and stderr are of type string, exit_code of type int and changed is of type bool

				logic to parse?
				- use a stack or queue? split line with a space?
				- will there be '(' ')' ? If so need to add priority in execution
		*/
		return register.ParseRegisterExp(w.rvar)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(w.timeout)*time.Second)
	defer cancel()

	cmd := command.New(ctx, "", []string{})
	cmd.WithExpression("bash", w.command)
	if _, err := cmd.Run(); err != nil {
		return false, err
	}
	if cmd.Status.ExitCode == w.exitCode {
		return true, nil
	}
	return false, fmt.Errorf("exit code %q doesn't match the expected exit code %q", cmd.Status.ExitCode, w.exitCode)
}
