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

func TestFileAction(t *testing.T) {
	tests := []struct {
		name    string
		input   *parser.Action
		wantErr error
	}{
		{
			name: "touch Dir /opt/pulse/bin",
			input: &parser.Action{
				Action:      "file",
				Name:        "touch dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/opt/pulse/bin",
						},
					},
					"dir":        true,
					"state":      "touch",
					"permission": "0755",
					"owner":      "root",
					"group":      "root",
					"force":      true,
				},
			},
			wantErr: nil,
		},
		{
			name: "remove Dir /opt/pulse/bin",
			input: &parser.Action{
				Action:      "file",
				Name:        "touch dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/opt/pulse/bin",
						},
					},
					"dir":        true,
					"permission": "0755",
					"owner":      "root",
					"group":      "root",
					"state":      "absent",
					"force":      true,
				},
			},
			wantErr: nil,
		},
		{
			name: "touch file /opt/pulse/post.tmp",
			input: &parser.Action{
				Action:      "file",
				Name:        "touch file",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/opt/pulse/post.tmp",
						},
					},
					"dir":        false,
					"permission": "0755",
					"owner":      "root",
					"group":      "root",
					"state":      "touch",
					"force":      true,
				},
			},
			wantErr: nil,
		},
		{
			name: "remove file /opt/pulse/post.tmp",
			input: &parser.Action{
				Action:      "file",
				Name:        "remove file",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/opt/pulse/post.tmp",
						},
					},
					"dir":        false,
					"permission": "0755",
					"owner":      "root",
					"group":      "root",
					"state":      "absent",
					"force":      true,
				},
			},
			wantErr: nil,
		},
		{
			name: "touch file /opt/pulse/post.tmp",
			input: &parser.Action{
				Action:      "file",
				Name:        "touch file",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/opt/pulse/post.tmp",
						},
					},
					"dir":        false,
					"permission": "0755",
					"owner":      "root",
					"group":      "root",
					"state":      "touch",
					"force":      true,
				},
			},
			wantErr: nil,
		},
		{
			name: "create symlink for file /opt/pulse/post.tmp",
			input: &parser.Action{
				Action:      "file",
				Name:        "remove file",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"src":  "/opt/pulse/post.tmp",
							"dest": "/opt/post.tmp",
						},
					},
					"dir":        false,
					"permission": "0755",
					"owner":      "root",
					"group":      "root",
					"state":      "link",
					"force":      true,
				},
			},
			wantErr: nil,
		},
		{
			name: "remove symlink for file /opt/pulse/post.tmp",
			input: &parser.Action{
				Action:      "file",
				Name:        "remove file",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/opt/post.tmp",
						},
					},
					"dir":        false,
					"permission": "0755",
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
		fileAction := NewFileAction("test", 10, register.GetHash(tc.name))
		register.RMap[register.GetHash(tc.name)] = &register.Register{}
		wLog := make(chan interface{})
		var got error

		go func() {
			got = fileAction.Do(tc.input, wLog)
			close(wLog)
		}()

		// This is here to wait for the channel
		for range wLog {
		}
		if !reflect.DeepEqual(tc.wantErr, got) {
			t.Fatalf("%s: expected: %v, got: %v", tc.name, tc.wantErr, got)
		}
	}
}
