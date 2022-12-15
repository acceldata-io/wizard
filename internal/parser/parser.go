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

package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/acceldata-io/goutils/netutils"
)

type TaskList struct {
	Tasks    map[string][]*Action `json:"tasks"`
	Priority []string             `json:"priority"`
}

type Action struct {
	Action          string                 `json:"action" validate:"required"`
	Name            string                 `json:"name" validate:"required"`
	When            *when                  `json:"when"`
	Command         []string               `json:"command"`
	ExitCode        float64                `json:"exit_code"`
	IgnoreError     bool                   `json:"ignore_error"`
	ActionVariables map[string]interface{} `json:"action_var"`
	Timeout         int                    `json:"timeout"`
	Register        string                 `json:"register"`
	BackupSrc       string
}

type when struct {
	Command  string `json:"cmd"`
	RVar     string `json:"rvar"`
	ExitCode int    `json:"exit_code"`
}

var env = make(map[string]string)

var wizardFacts = map[string]interface{}{
	"os_hostname":   getFactOSHostname,
	"fqdn_hostname": getFactFQDNHostname,
	"cmd_hostname":  getFactCMDHostname,
	"env":           func(envKey string) string { return env[envKey] },
}

// ParseConfig populates the json data into the Tasks structure
func ParseConfig(Config []byte) (config TaskList, err error) {
	if err = json.Unmarshal(Config, &config); err != nil {
		err = fmt.Errorf("ParseConfig: error unmarshalling config json - %s", err.Error())
	}
	return config, err
}

// ParseWizardFacts populated the yaml data into the TemplateConfig structure
func ParseWizardFacts() (facts map[string]interface{}, err error) {
	return GetWizardFacts(), err
}

// SetEnv gets all the env in the system and assign it to Env Global var
func SetEnv() {
	envList := os.Environ()
	for _, val := range envList {
		envStrings := strings.Split(val, "=")
		if len(envStrings) == 2 {
			env[strings.TrimSpace(envStrings[0])] = strings.TrimSpace(envStrings[1])
		}
	}
}

func GetWizardFacts() map[string]interface{} {
	return wizardFacts
}

func getFactOSHostname() string {
	hostname, _ := netutils.GetHostName("OS", 10)
	return hostname
}

func getFactFQDNHostname() string {
	hostname, _ := netutils.GetHostName("FQDN", 10)
	return hostname
}

func getFactCMDHostname() string {
	hostname, _ := netutils.GetHostName("CMD", 10)
	return hostname
}

func MergeFuncMap(funcMap, facts map[string]interface{}) map[string]interface{} {
	wizardFuncMap := make(map[string]interface{}, len(funcMap)+len(facts))
	for k, v := range funcMap {
		wizardFuncMap[k] = v
	}
	for k, v := range facts {
		wizardFuncMap[k] = v
	}
	return wizardFuncMap
}
