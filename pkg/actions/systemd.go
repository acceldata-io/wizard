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
	"encoding/json"
	"fmt"

	"github.com/acceldata-io/wizard/internal/parser"
	"github.com/acceldata-io/wizard/pkg/wlog"

	"github.com/acceldata-io/goutils/libsysd"

	"github.com/go-playground/validator/v10"
)

type systemD struct {
	timeout  int
	register string
}

type systemDVar struct {
	Name         string `json:"name" validate:"required"`
	DaemonReload bool   `json:"daemon_reload"`
	Force        bool   `json:"force"`
	State        string `json:"state" validate:"required"`
}

func NewSystemDAction(timeout int, register string) Action {
	return &systemD{timeout: timeout, register: register}
}

func NewSystemDVar(data map[string]interface{}) (*systemDVar, error) {
	s := systemDVar{}

	if dataB, err := json.Marshal(data); err == nil {
		if err := json.Unmarshal(dataB, &s); err != nil {
			return &s, err
		}
	} else {
		return &s, err
	}

	validate := validator.New()
	err := validate.Struct(s)
	if err != nil {
		return &s, err
	}

	return &s, nil
}

func (s *systemD) Do(actions *parser.Action, wizardLog chan interface{}) error {
	wizardLog <- wlog.WLInfo("executing when condition")
	if actions.When != nil {
		when := NewWhen(actions.When.Command, actions.When.RVar, actions.When.ExitCode, s.timeout)
		successfulExec, err := when.Execute()
		if err != nil {
			if actions.When.RVar != "" {
				wizardLog <- wlog.WLError("register field validation error: " + err.Error())
				return fmt.Errorf("whenNotSatisfied")
			}
			wizardLog <- wlog.WLError("when condition not satisfied: " + err.Error())
			return fmt.Errorf("whenNotSatisfied")
		}
		if !successfulExec {
			return fmt.Errorf("whenNotSatisfied")
		}
	}
	vars, err := NewSystemDVar(actions.ActionVariables)
	if err != nil {
		wizardLog <- wlog.WLError("validation error: " + err.Error())
		return err
	}
	systemD := libsysd.NewSystemDAdapter()

	if vars.DaemonReload {
		wizardLog <- wlog.WLInfo("reloading the systemd daemon")
		if err := systemD.ReloadDaemon(); err != nil {
			return err
		}
	}

	switch vars.State {
	case "restart":
		wizardLog <- wlog.WLInfo("restarting systemd service: " + vars.Name)
		if _, err := systemD.RestartService(vars.Name); err != nil {
			return err
		}
	case "start":
		wizardLog <- wlog.WLInfo("starting systemd service: " + vars.Name)
		if err := systemD.StartService(vars.Name); err != nil {
			return err
		}
	case "stop":
		wizardLog <- wlog.WLInfo("stopping systemd service: " + vars.Name)
		if err := systemD.StopService(vars.Name); err != nil {
			return err
		}
	case "reload":
		wizardLog <- wlog.WLInfo("reloading systemd service: " + vars.Name)
		if err := systemD.ReloadService(vars.Name); err != nil {
			return err
		}
	default:
		wizardLog <- wlog.WLError(fmt.Sprintf("unknown state: %s", vars.State))
		return fmt.Errorf("unknown state: %s", vars.State)
	}

	return nil
}
