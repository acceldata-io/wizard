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
	"reflect"
	"testing"

	"github.com/acceldata-io/wizard/internal/parser"
	"github.com/acceldata-io/wizard/pkg/register"
)

func TestTemplateAction(t *testing.T) {
	PackageFiles = files

	tests := []struct {
		name    string
		input   *parser.Action
		wantErr error
	}{
		{
			name: "generate and copy yml file to dest",
			input: &parser.Action{
				Action:      "template",
				Name:        "copy file to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"parents":    true,
					"src":        "testdata/test.yml.tmpl",
					"dest":       "/tmp/test/test.yml",
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
				},
			},
			wantErr: nil,
		},
		{
			name: "remove yml file",
			input: &parser.Action{
				Action:      "file",
				Name:        "touch dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/tmp/test/test.yml",
						},
					},
					"dir":        false,
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
					"state":      "absent",
					"force":      true,
				},
			},
			wantErr: nil,
		},
		{
			name: "generate and copy yml file to dest",
			input: &parser.Action{
				Action:      "template",
				Name:        "copy file to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"parents":    true,
					"src":        "testdata/test.yml.tmpl",
					"dest":       "/tmp/test/test.yml",
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
				},
			},
			wantErr: nil,
		},
		{
			name: "generate and copy yml file to dest - same hash",
			input: &parser.Action{
				Action:      "template",
				Name:        "copy file to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"parents":    true,
					"src":        "testdata/test.yml.tmpl",
					"dest":       "/tmp/test/test.yml",
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
				},
			},
			wantErr: nil,
		},
		{
			name: "remove yml file",
			input: &parser.Action{
				Action:      "file",
				Name:        "touch dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/tmp/test/test.yml",
						},
					},
					"dir":        false,
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
					"state":      "absent",
					"force":      true,
				},
			},
			wantErr: nil,
		},
		{
			name: "generate and copy yml file to dest",
			input: &parser.Action{
				Action:      "template",
				Name:        "copy file to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"parents":    true,
					"src":        "testdata/test.yml.tmpl",
					"dest":       "/tmp/test/test.yml",
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
				},
			},
			wantErr: nil,
		},
		{
			name: "generate and copy yml file to dest - diff hash",
			input: &parser.Action{
				Action:      "template",
				Name:        "copy file to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"parents":    true,
					"src":        "testdata/test2.yml.tmpl",
					"dest":       "/tmp/test/test.yml",
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
				},
			},
			wantErr: nil,
		},
		{
			name: "remove yml file",
			input: &parser.Action{
				Action:      "file",
				Name:        "touch dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/tmp/test/test.yml",
						},
					},
					"dir":        false,
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
					"state":      "absent",
					"force":      true,
				},
			},
			wantErr: nil,
		},
		{
			name: "generate and copy hydra.service file to dest",
			input: &parser.Action{
				Action:      "template",
				Name:        "copy file to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"parents":    true,
					"src":        "testdata/hydra_test.service.tmpl",
					"dest":       "/tmp/test/hydra_test.service",
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
				},
			},
			wantErr: nil,
		},
		{
			name: "remove hydra.service file",
			input: &parser.Action{
				Action:      "file",
				Name:        "touch dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/tmp/test/hydra_test.service",
						},
					},
					"dir":        false,
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
					"state":      "absent",
					"force":      true,
				},
			},
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		var got error
		wLog := make(chan interface{})
		if tc.input.Action == "template" {
			templateAction := NewTemplateAction("test", nil, parser.GetWizardFacts(), 10, register.GetHash(tc.name))
			register.RMap[register.GetHash(tc.name)] = &register.Register{}
			go func() {
				got = templateAction.Do(tc.input, wLog)
				close(wLog)
			}()
		} else if tc.input.Action == "file" {
			fileAction := NewFileAction("test", 10, register.GetHash(tc.name))
			register.RMap[register.GetHash(tc.name)] = &register.Register{}
			go func() {
				got = fileAction.Do(tc.input, wLog)
				close(wLog)
			}()
		}

		// This is here to wait for the channel
		for range wLog {
		}
		if !reflect.DeepEqual(tc.wantErr, got) {
			t.Fatalf("%s: expected: %v, got: %v", tc.name, tc.wantErr, got)
		}
	}
}
