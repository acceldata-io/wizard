package actions

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/acceldata-io/wizard/internal/parser"
	"github.com/acceldata-io/wizard/pkg/register"
)

func TestNewUserVars(t *testing.T) {
	userVars := map[string]interface{}{
		"name":  "test",
		"force": false,
		"uid":   "test",
		"gid":   "test",
		"shell": "bash",
		"home":  "home",
		"state": "absent",
	}

	actualUserVars, _ := newUserVars(userVars)
	if actualUserVars.Name != "test" || actualUserVars.UserID != "test" || actualUserVars.Force != false || actualUserVars.Home != "home" || actualUserVars.GroupID != "test" || actualUserVars.Shell != "bash" || actualUserVars.State != "absent" {
		t.Fail()
	}
}

func TestUserAction(t *testing.T) {
	tests := []struct {
		name    string
		input   *parser.Action
		wantErr error
	}{
		{
			name: "add user adpulse",
			input: &parser.Action{
				Action:      "user",
				Name:        "user add",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"name":  "adpulse",
					"home":  "/opt/pulse",
					"shell": "/bin/sh",
					"uid":   "996",
					"gid":   "992",
					"force": false,
					"state": "present",
				},
			},
			wantErr: nil,
		},
		//{
		//	name: "add user adpulse should return user already exists error",
		//	input: &parser.Action{
		//		Action:      "user",
		//		Name:        "user add",
		//		IgnoreError: false,
		//		ActionVariables: map[string]interface{}{
		//			"name":  "adpulse",
		//			"home":  "/opt/pulse",
		//			"shell": "/bin/sh",
		//			"uid":   "996",
		//			"gid":   "992",
		//			"force": true,
		//			"state": "present",
		//		},
		//	},
		//	wantErr: fmt.Errorf("status code not 0 - useradd: user 'adpulse' already exists\n"),
		//},
		{
			name: "del user adpulse",
			input: &parser.Action{
				Action:      "user",
				Name:        "user add",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"name":  "adpulse",
					"state": "absent",
				},
			},
			wantErr: nil,
		},
		{
			name: "del user adpulse should return user not found error",
			input: &parser.Action{
				Action:      "user",
				Name:        "user add",
				IgnoreError: false,
				ActionVariables: map[string]interface{}{
					"name":  "adpulse",
					"state": "absent",
				},
			},
			wantErr: fmt.Errorf("USERDEL: User not found"),
		},
	}

	for _, tc := range tests {
		register.RMap[register.GetHash(tc.name)] = &register.Register{}
		userAction := NewUserAction(10, register.GetHash(tc.name))

		var got error
		wLog := make(chan interface{})
		go func() {
			got = userAction.Do(tc.input, wLog)
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
