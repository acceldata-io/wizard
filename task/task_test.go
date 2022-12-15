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

package task

import (
	"embed"
	"fmt"
	"os"
	"testing"

	actions_factory_mock "github.com/acceldata-io/wizard/factory/action/mocks"
	mock_actions "github.com/acceldata-io/wizard/pkg/actions/mocks"
	"github.com/golang/mock/gomock"
)

func TestNewFail(t *testing.T) {
	file, _ := os.ReadFile("../testdata/parser_config_fail.yaml")
	_, err := New(file, embed.FS{}, TemplateOptions{
		EnableWizardFacts: false,
	})
	if err == nil {
		t.Fail()
	}
}

func TestNewPass(t *testing.T) {
	file, _ := os.ReadFile("../testdata/parser_config_pass.json")
	task, err := New(file, embed.FS{}, TemplateOptions{
		EnableWizardFacts: false,
	})
	if err != nil {
		t.Fail()
	}
	if len(task.taskList.Tasks) == 0 {
		t.Fail()
	}
}

func TestPerformPass(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	actionsMock := mock_actions.NewMockAction(ctrl)
	actionsMock.EXPECT().Do(gomock.Any(), gomock.Any()).Times(2).Return(nil)

	actionsFactoryMock := actions_factory_mock.NewMockActionsFactory(ctrl)
	actionsFactoryMock.EXPECT().NewActions(gomock.Any(), "hydra", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(actionsMock)
	actionsFactoryMock.EXPECT().NewActions(gomock.Any(), "hydra2", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(actionsMock)

	file, _ := os.ReadFile("../testdata/task_pass.json")
	task, err := New(file, embed.FS{}, TemplateOptions{
		EnableWizardFacts: false,
	})
	if err != nil {
		t.Fail()
	}
	if len(task.taskList.Tasks) == 0 {
		t.Fail()
	}

	task.actionFactory = actionsFactoryMock
	_, err = task.Execute()

	if err != nil {
		t.Fail()
	}
}

func TestPerformWhenError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	actionsMock := mock_actions.NewMockAction(ctrl)
	actionsMock.EXPECT().Do(gomock.Any(), gomock.Any()).Times(2).Return(fmt.Errorf("whenNotSatisfied"))

	actionsFactoryMock := actions_factory_mock.NewMockActionsFactory(ctrl)
	actionsFactoryMock.EXPECT().NewActions(gomock.Any(), "hydra", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(actionsMock)
	actionsFactoryMock.EXPECT().NewActions(gomock.Any(), "hydra2", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(actionsMock)

	file, _ := os.ReadFile("../testdata/task_pass.json")
	task, err := New(file, embed.FS{}, TemplateOptions{
		EnableWizardFacts: false,
	})
	if err != nil {
		t.Fail()
	}
	if len(task.taskList.Tasks) == 0 {
		t.Fail()
	}

	task.actionFactory = actionsFactoryMock
	_, err = task.Execute()
	if err != nil {
		t.Fail()
	}
}

func TestPerformIgnoreErrorFalse(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	actionsMock := mock_actions.NewMockAction(ctrl)
	actionsMock.EXPECT().Do(gomock.Any(), gomock.Any()).Times(1).Return(fmt.Errorf("random error"))

	actionsFactoryMock := actions_factory_mock.NewMockActionsFactory(ctrl)
	actionsFactoryMock.EXPECT().NewActions(gomock.Any(), "hydra", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(actionsMock)

	file, _ := os.ReadFile("../testdata/task_pass.json")
	task, err := New(file, embed.FS{}, TemplateOptions{
		EnableWizardFacts: false,
	})
	if err != nil {
		t.Fail()
	}
	if len(task.taskList.Tasks) == 0 {
		t.Fail()
	}

	task.actionFactory = actionsFactoryMock
	_, err = task.Execute()
	if err == nil {
		t.Fail()
	}
}

func TestPerformIgnoreErrorTrue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	actionsMock := mock_actions.NewMockAction(ctrl)
	actionsMock.EXPECT().Do(gomock.Any(), gomock.Any()).Times(1).Return(fmt.Errorf("random error"))

	actionsFactoryMock := actions_factory_mock.NewMockActionsFactory(ctrl)
	actionsFactoryMock.EXPECT().NewActions(gomock.Any(), "hydra2", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(actionsMock)

	file, _ := os.ReadFile("../testdata/task_ignore_error_true.json")
	task, err := New(file, embed.FS{}, TemplateOptions{
		EnableWizardFacts: false,
	})
	if err != nil {
		t.Fail()
	}
	if len(task.taskList.Tasks) == 0 {
		t.Fail()
	}

	task.actionFactory = actionsFactoryMock
	_, err = task.Execute()
	if err != nil {
		t.Fail()
	}
}
