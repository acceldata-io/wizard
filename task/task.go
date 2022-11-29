package task

import (
	"embed"
	"fmt"

	"github.com/acceldata-io/wizard/factory/action"
	"github.com/acceldata-io/wizard/internal/parser"
	"github.com/acceldata-io/wizard/pkg/actions"
	"github.com/acceldata-io/wizard/pkg/register"
	"github.com/acceldata-io/wizard/pkg/wlog"
)

// Task is used for performing tasks/actions
type Task struct {
	taskList       parser.TaskList
	actionFactory  action.ActionsFactory
	templateConfig interface{}
	wizardFacts    map[string]interface{}
}

// TemplateOptions are used for the template action
// If EnableWizardFacts is set to 'true' then the wizard can use all the ENV variables and some predefined facts in the template
// TemplateConfig is the user defined structure to use in the template
type TemplateOptions struct {
	EnableWizardFacts bool
	TemplateConfig    interface{}
}

// New parses the input config and returns a Task, log chan, error if any
// function parameters:
// config -> is the user defined task list JSON
// packageFiles -> embedded file system that the wizard will use to perform actions on
// tmplOptions -> for template action
func New(config []byte, packageFiles embed.FS, tmplOptions TemplateOptions) (*Task, error) {
	taskList, err := parser.ParseConfig(config)
	if err != nil {
		return nil, fmt.Errorf("new: %s", err.Error())
	}

	var wizardFacts map[string]interface{}
	if tmplOptions.EnableWizardFacts {
		wizardFacts, err = parser.ParseWizardFacts()
		if err != nil {
			return nil, fmt.Errorf("new: %s", err.Error())
		}
	}

	parser.SetEnv()

	actions.PackageFiles = packageFiles
	return &Task{
		taskList:       taskList,
		actionFactory:  action.NewActionsFactory(),
		templateConfig: tmplOptions.TemplateConfig,
		wizardFacts:    wizardFacts,
	}, nil
}

var wizardLog chan interface{}

// NewWithLog parses the config and returns task struct
func NewWithLog(config []byte, packageFiles embed.FS, tmplOptions TemplateOptions) (*Task, chan interface{}, error) {
	wizardLog = make(chan interface{})
	taskList, err := parser.ParseConfig(config)
	if err != nil {
		return nil, wizardLog, fmt.Errorf("NewWithLog: %s", err.Error())
	}

	var wizardFacts map[string]interface{}
	if tmplOptions.EnableWizardFacts {
		wizardFacts, err = parser.ParseWizardFacts()
		if err != nil {
			return nil, wizardLog, fmt.Errorf("NewWithLog: %s", err.Error())
		}
	}

	parser.SetEnv()

	actions.PackageFiles = packageFiles
	return &Task{
		taskList:       taskList,
		actionFactory:  action.NewActionsFactory(),
		templateConfig: tmplOptions.TemplateConfig,
		wizardFacts:    wizardFacts,
	}, wizardLog, nil
}

// Perform iterates through each task and performs actions based on the priority list
// Takes the log chan as input parameter to input logs
func (t *Task) Perform(logCh chan interface{}) error {
	defer close(logCh)
	for _, priorityName := range t.taskList.Priority {
		taskName := priorityName
		taskActions := t.taskList.Tasks[taskName]
		for _, play := range taskActions {
			logCh <- wlog.WLInfo(fmt.Sprintf("Perform: Task: %s Action: %s, Name: %s", taskName, play.Action, play.Name))
			if play.Register == "" {
				play.Register = register.GetHash(play.Name)
			}
			register.RMap[play.Register] = &register.Register{}
			newAction := t.actionFactory.NewActions(play, taskName, t.templateConfig, t.wizardFacts, play.Timeout, play.Register)
			err := newAction.Do(play, logCh)
			if err != nil {
				aRegister := register.RMap[play.Register]
				aRegister.StdErr = err.Error()

				if err.Error() == "whenNotSatisfied" {
					logCh <- wlog.WLWarn(fmt.Sprintf("Perform: Task: %s Action: %s, Name: %s, Err: %s", taskName, play.Action, play.Name, err.Error()))
					continue
				} else if !play.IgnoreError {
					logCh <- wlog.WLError(fmt.Sprintf("Perform: Task: %s Action: %s, Name: %s, Err: %s", taskName, play.Action, play.Name, err.Error()))
					return fmt.Errorf("perform: Task: %s Action: %s, Name: %s, Error: %s", taskName, play.Action, play.Name, err.Error())
				}
			}
		}
	}
	return nil
}

// Execute iterates through each task and performs actions based on the priority list
// Returns an []interface, error
// []interface are logs of wlog pkg types
func (t *Task) Execute() ([]interface{}, error) {
	var err error
	logs := []interface{}{}
	logCh := make(chan interface{})

	// spawns the main logic
	go func() {
		err = t.Perform(logCh)
	}()

	// Waits for the go routine to complete and collects logs
	for v := range logCh {
		logs = append(logs, v)
	}

	return logs, err
}
