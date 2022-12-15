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

package action

import (
	"github.com/acceldata-io/wizard/internal/parser"
	"github.com/acceldata-io/wizard/pkg/actions"
)

//go:generate mockgen -source actions_factory.go -destination ./mocks/actions_factory_mock.go -package actions_factory_mock

type ActionsFactory interface {
	NewActions(list *parser.Action, agentName string, config interface{}, wizardFacts map[string]interface{}, timeout int, register string) actions.Action
}

type actionsFactory struct{}

func NewActionsFactory() ActionsFactory {
	return &actionsFactory{}
}

func (a *actionsFactory) NewActions(action *parser.Action, agentName string, config interface{}, wizardFacts map[string]interface{}, timeout int, register string) actions.Action {
	var actionDo actions.Action
	if timeout == 0 {
		timeout = 10
	}
	switch action.Action {
	case "copy":
		actionDo = actions.NewCopyAction(agentName, timeout, register)
	case "template":
		actionDo = actions.NewTemplateAction(agentName, config, wizardFacts, timeout, register)
	case "file":
		actionDo = actions.NewFileAction(agentName, timeout, register)
	case "cmd":
		actionDo = actions.NewCmdAction(timeout, register)
	case "user":
		actionDo = actions.NewUserAction(timeout, register)
	case "systemd":
		actionDo = actions.NewSystemDAction(timeout, register)
	}
	return actionDo
}
