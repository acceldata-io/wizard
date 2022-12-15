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
	"embed"
	"fmt"
	"reflect"
	"testing"

	"github.com/acceldata-io/wizard/internal/parser"
	"github.com/acceldata-io/wizard/pkg/register"
)

//go:embed testdata
var files embed.FS

func TestCopyAction(t *testing.T) {
	PackageFiles = files

	tests := []struct {
		name    string
		input   *parser.Action
		wantErr error
	}{
		{
			name: "Copy embedded files with * to dest dir",
			input: &parser.Action{
				Action:      "copy",
				Name:        "copy file to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"src":        "testdata/*",
					"dest":       "/tmp/test_rec",
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
				},
			},
			wantErr: nil,
		},
		{
			name: "remove Dir /tmp/tesst",
			input: &parser.Action{
				Action:      "file",
				Name:        "touch dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/tmp/test_rec",
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
			name: "Copy embedded files with * to dest dir - recursive false",
			input: &parser.Action{
				Action:      "copy",
				Name:        "copy file to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"src":        "testdata/*/post*_test.sh",
					"dest":       "/tmp/test_rec",
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
					"recursive":  false,
				},
			},
			wantErr: nil,
		},
		{
			name: "remove Dir /tmp/tesst",
			input: &parser.Action{
				Action:      "file",
				Name:        "touch dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/tmp/test_rec",
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
			name: "Copy embedded files with * to dest dir - recursive true",
			input: &parser.Action{
				Action:      "copy",
				Name:        "copy file to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"src":        "testdata/*/post*_test.sh",
					"dest":       "/tmp/test_rec",
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
					"recursive":  true,
				},
			},
			wantErr: nil,
		},
		{
			name: "remove Dir /tmp/tesst",
			input: &parser.Action{
				Action:      "file",
				Name:        "touch dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/tmp/test_rec",
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
			name: "Copy embedded file to dest dir",
			input: &parser.Action{
				Action:      "copy",
				Name:        "copy file to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"src":        "testdata/postinstall_test.sh",
					"dest":       "/tmp",
					"permission": "0644",
					"owner":      "root",
					"group":      "root",
				},
			},
			wantErr: nil,
		},
		{
			name: "Copy embedded file to dest dir - same hash - overwrite",
			input: &parser.Action{
				Action:      "copy",
				Name:        "copy file to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"src":        "testdata/postinstall_test.sh",
					"dest":       "/tmp",
					"permission": "0644",
					"owner":      "root",
					"group":      "root",
					"over_write": true,
				},
			},
			wantErr: nil,
		},
		{
			name: "Copy embedded file to dest dir - same hash",
			input: &parser.Action{
				Action:      "copy",
				Name:        "copy file to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"src":        "testdata/postinstall_test.sh",
					"dest":       "/tmp",
					"permission": "0644",
					"owner":      "root",
					"group":      "root",
				},
			},
			wantErr: nil,
		},
		{
			name: "Copy embedded file to dest dir - different hash",
			input: &parser.Action{
				Action:      "copy",
				Name:        "copy file to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"src":        "testdata/test_dir/postinstall_test.sh",
					"dest":       "/tmp",
					"permission": "0644",
					"owner":      "root",
					"group":      "root",
				},
			},
			wantErr: nil,
		},
		{
			name: "remove file /tmp/postinstall_test.sh",
			input: &parser.Action{
				Action:      "file",
				Name:        "remove file",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/tmp/postinstall_test.sh",
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
			name: "Copy local file to dest dir",
			input: &parser.Action{
				Action:      "copy",
				Name:        "copy file to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "local",
					"src":        "testdata/postinstall_test.sh",
					"dest":       "/tmp",
					"permission": "0644",
					"owner":      "root",
					"group":      "root",
				},
			},
			wantErr: nil,
		},
		{
			name: "remove file /tmp/postinstall_test.sh",
			input: &parser.Action{
				Action:      "file",
				Name:        "remove file",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/tmp/postinstall_test.sh",
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
			name: "Copy embedded dir to dest dir",
			input: &parser.Action{
				Action:      "copy",
				Name:        "copy dir to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"src":        "testdata/test_dir",
					"dest":       "/tmp/test",
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
					"parents":    true,
				},
			},
			wantErr: nil,
		},
		{
			name: "remove Dir /tmp/test",
			input: &parser.Action{
				Action:      "file",
				Name:        "touch dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/tmp/test",
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
			name: "Copy local dir to dest dir",
			input: &parser.Action{
				Action:      "copy",
				Name:        "copy dir to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "local",
					"src":        "testdata/test_dir",
					"dest":       "/tmp/test",
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
					"parents":    true,
				},
			},
			wantErr: nil,
		},
		{
			name: "remove Dir /tmp/test",
			input: &parser.Action{
				Action:      "file",
				Name:        "touch dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/tmp/test",
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
			name: "Copy embedded file to dest file",
			input: &parser.Action{
				Action:      "copy",
				Name:        "copy file to file",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"src":        "testdata/postinstall_test.sh",
					"dest":       "/tmp/test.sh",
					"permission": "0644",
					"owner":      "root",
					"group":      "root",
				},
			},
			wantErr: nil,
		},
		{
			name: "Copy embedded file to dest file - same hash",
			input: &parser.Action{
				Action:      "copy",
				Name:        "copy file to file",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"src":        "testdata/postinstall_test.sh",
					"dest":       "/tmp/test.sh",
					"permission": "0644",
					"owner":      "root",
					"group":      "root",
				},
			},
			wantErr: nil,
		},
		{
			name: "Copy embedded file to dest file - same hash - overwrite",
			input: &parser.Action{
				Action:      "copy",
				Name:        "copy file to file",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"src":        "testdata/postinstall_test.sh",
					"dest":       "/tmp/test.sh",
					"permission": "0644",
					"owner":      "root",
					"group":      "root",
					"over_write": true,
				},
			},
			wantErr: nil,
		},
		{
			name: "Copy embedded file to dest file - different hash",
			input: &parser.Action{
				Action:      "copy",
				Name:        "copy file to file",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "embed",
					"src":        "testdata/preinstall_test.sh",
					"dest":       "/tmp/test.sh",
					"permission": "0644",
					"owner":      "root",
					"group":      "root",
				},
			},
			wantErr: nil,
		},
		{
			name: "remove file /tmp/test.sh",
			input: &parser.Action{
				Action:      "file",
				Name:        "remove file",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"files": []interface{}{
						map[string]interface{}{
							"dest": "/tmp/test.sh",
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
			name: "Copy local dir to dest dir - Recursive false - Error",
			input: &parser.Action{
				Action:      "copy",
				Name:        "copy dir to dir",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"src_type":   "local",
					"src":        "testdata/test_dir",
					"dest":       "/tmp/test/test1",
					"permission": "0655",
					"owner":      "root",
					"group":      "root",
					"parents":    false,
				},
			},
			wantErr: fmt.Errorf("destination dir - /tmp/test/test1 not found: stat /tmp/test/test1: no such file or directory"),
		},
	}

	for _, tc := range tests {
		var got error
		wLog := make(chan interface{})
		if tc.input.Action == "copy" {
			copyAction := NewCopyAction("test", 10, register.GetHash(tc.name))
			register.RMap[register.GetHash(tc.name)] = &register.Register{}
			go func() {
				got = copyAction.Do(tc.input, wLog)
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
