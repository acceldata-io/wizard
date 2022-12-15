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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/acceldata-io/wizard/internal/parser"
	"github.com/acceldata-io/wizard/pkg/register"
	"github.com/acceldata-io/wizard/pkg/wlog"

	command "github.com/acceldata-io/goutils/shellutils/cmd"
	"github.com/go-playground/validator/v10"
)

type actionUser struct {
	timeout  int
	register string
}

func NewUserAction(timeout int, localRegister string) Action {
	return &actionUser{timeout: timeout, register: localRegister}
}

type userVars struct {
	Name    string `json:"name" validate:"required"`
	Home    string `json:"home" validate:"required_if=State present"`
	Shell   string `json:"shell" validate:"required_if=State present"`
	UserID  string `json:"uid" validate:"required_if=State present"`
	GroupID string `json:"gid" validate:"required_if=State present"`
	Force   bool   `json:"force"`
	State   string `json:"state" validate:"required"`
}

func newUserVars(data map[string]interface{}) (*userVars, error) {
	userVar := userVars{}

	if dataB, err := json.Marshal(data); err == nil {
		if err := json.Unmarshal(dataB, &userVar); err != nil {
			return &userVar, err
		}
	} else {
		return &userVar, err
	}

	validate := validator.New()
	err := validate.Struct(userVar)
	if err != nil {
		return &userVar, err
	}
	return &userVar, nil
}

func (u *actionUser) Do(actions *parser.Action, wizardLog chan interface{}) error {
	uRegister := register.RMap[u.register]

	wizardLog <- wlog.WLInfo("executing when condition")
	if actions.When != nil {
		when := NewWhen(actions.When.Command, actions.When.RVar, actions.When.ExitCode, u.timeout)
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
	userVars, err := newUserVars(actions.ActionVariables)
	if err != nil {
		wizardLog <- wlog.WLError("validation error: " + err.Error())
		return err
	}
	useUID, useGID, useHomeDir := false, false, false
	isUserDel := false

	// TODO - If no binaries are found, edit the /etc/passwd for user and /etc/group for group

	wizardLog <- wlog.WLInfo("checking user and group core binaries")
	userBins, err := CheckUserCoreBinaries(u.timeout)
	if err != nil {
		wizardLog <- wlog.WLError("unable to check for useradd/del OS binaries. Because: " + err.Error())
		return err
	}
	isUserAddPresent := userBins["useradd"]
	isAddUserPresent := userBins["adduser"]
	isUserDelPresent := userBins["userdel"]
	isDelUserPresent := userBins["deluser"]

	groupBins, err := CheckGroupCoreBinaries(u.timeout)
	if err != nil {
		wizardLog <- wlog.WLError("unable to check for groupadd/del OS binaries. Because: " + err.Error())
		return err
	}
	isAddGroupPresent := groupBins["addgroup"]
	isGroupAddPresent := groupBins["groupadd"]

	wizardLog <- wlog.WLInfo("checking if user is present with all the properties")
	// Check if the user is present with all the properties
	if userNameOK, err := isUserPresent(userVars.Name); !userNameOK {
		wizardLog <- wlog.WLInfo("user with name " + userVars.Name + " not found")
		if err != nil {
			return err
		}
		if userVars.State == "absent" {
			return fmt.Errorf("USERDEL: User not found")
		}

		uid, err := strconv.Atoi(userVars.UserID)
		if err != nil {
			return fmt.Errorf("unable to convert user id to int - %s", err)
		}
		if uidOK, err := isUserIDPresent(uid); !uidOK {
			if err != nil {
				return err
			}
			useUID = true
		}

		if homeDirOK, _, err := isHomeDirPresent(userVars.Home); !homeDirOK {
			if err != nil {
				return err
			}
			useHomeDir = true
			wizardLog <- wlog.WLInfo("home dir not present, will use: " + userVars.Home)
		} else {
			return fmt.Errorf("HomeDirFound")
		}
	} else {
		if userVars.State == "present" {
			if userVars.Force {
				userInfo, err := userLookup(userVars.Name)
				if err != nil {
					return err
				}
				wizardLog <- wlog.WLInfo("user with name " + userVars.Name + " found")
				if userInfo.homedir == userVars.Home {
					wizardLog <- wlog.WLInfo("home dir matches, returning nil")
					return nil
				} else {
					if homeDirOK, _, err := isHomeDirPresent(userVars.Home); !homeDirOK {
						if err != nil {
							return err
						}
						wizardLog <- wlog.WLInfo("home dir not present, will use: " + userVars.Home)
						// TODO - Modify the /etc/passwd with the userVars home
					} else {
						return fmt.Errorf("HomeDirFound")
					}
				}
			} else {
				wizardLog <- wlog.WLInfo("User already exists")
				return nil
			}
		} else if userVars.State == "absent" {
			isUserDel = true
		}
	}

	wizardLog <- wlog.WLInfo("checking if group " + userVars.GroupID + " is present")
	if groupNameOK, err := isGroupNamePresent(userVars.Name); !groupNameOK {
		if err != nil {
			return err
		}
		gid, err := strconv.Atoi(userVars.GroupID)
		if err != nil {
			return err
		}

		if gidOK, err := isGroupIDPresent(gid); !gidOK {
			if err != nil {
				return err
			}

			var groupCmd *command.Command
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(u.timeout)*time.Second)
			defer cancel()
			if isGroupAddPresent {
				groupCmd = command.New(ctx, "groupadd", []string{"-g", userVars.GroupID, userVars.Name})
			} else if isAddGroupPresent {
				groupCmd = command.New(ctx, "addgroup", []string{"-g", userVars.GroupID, userVars.Name})
			} else {
				return fmt.Errorf("addgroup and groupadd not present")
			}
			wizardLog <- wlog.WLInfo("running command: " + groupCmd.Command + " " + strings.Join(groupCmd.Args, " "))
			output, err := groupCmd.Run()
			if err != nil {
				return fmt.Errorf("unable to run the %q - %s", groupCmd.Command, err.Error())
			}
			if output.Status.ExitCode != 0 {
				return fmt.Errorf("status code not 0 - %s", output.Status.StdErr)
			}
			useGID = true
		}
	}

	cmdArgs := []string{"-s", userVars.Shell}
	if useUID {
		cmdArgs = append(cmdArgs, "-u", userVars.UserID)
	}
	if useGID {
		cmdArgs = append(cmdArgs, "-g", userVars.GroupID)
	} else {
		cmdArgs = append(cmdArgs, "-N")
	}
	if useHomeDir {
		cmdArgs = append(cmdArgs, "-d", userVars.Home)
	}
	cmdArgs = append(cmdArgs, userVars.Name)

	wizardLog <- wlog.WLInfo("executing user command")
	wizardLog <- wlog.WLInfo("command arguments: " + strings.Join(cmdArgs, " "))
	var output *command.Command
	if userVars.State == "absent" {
		if isUserDel {
			var userCmd *command.Command
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(u.timeout)*time.Second)
			defer cancel()
			if isUserDelPresent {
				userCmd = command.New(ctx, "userdel", []string{userVars.Name})
			} else if isDelUserPresent {
				userCmd = command.New(ctx, "deluser", []string{userVars.Name})
			} else {
				return fmt.Errorf("userdel and deluser not found")
			}
			wizardLog <- wlog.WLInfo("running command: " + userCmd.Command)
			output, err = userCmd.Run()
			if err != nil {
				wizardLog <- wlog.WLError("unable to run the command: '" + userCmd.Command + " " + strings.Join(userCmd.Args, " ") + "'")
				return err
			}
		}
	} else {
		var userCmd *command.Command
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(u.timeout)*time.Second)
		defer cancel()
		if isUserAddPresent {
			userCmd = command.New(ctx, "useradd", cmdArgs)
		} else if isAddUserPresent {
			userCmd = command.New(ctx, "adduser", cmdArgs)
		} else {
			return fmt.Errorf("useradd and adduser not present")
		}
		wizardLog <- wlog.WLInfo("running command: " + userCmd.Command)
		output, err = userCmd.Run()
		if err != nil {
			wizardLog <- wlog.WLError("unable to run the command: '" + userCmd.Command + " " + strings.Join(userCmd.Args, " ") + "'")
			return err
		}
	}
	if output.Status.ExitCode != 0 {
		return fmt.Errorf("status code not 0 - %s", output.Status.StdErr)
	}
	uRegister.Changed = true
	return nil
}

func CheckUserCoreBinaries(timeout int) (map[string]bool, error) {
	//
	result := map[string]bool{
		"adduser": false,
		"useradd": false,
		"userdel": false,
		"deluser": false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	addUserCmd := command.New(ctx, "which", []string{"adduser"})
	output, err := addUserCmd.Run()
	if err != nil {
		return result, err
	}
	if output.Status.ExitCode == 0 {
		result["adduser"] = true
	}

	ctx, cancel2 := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel2()

	userAddCmd := command.New(ctx, "which", []string{"useradd"})
	output, err = userAddCmd.Run()
	if err != nil {
		return result, err
	}
	if output.Status.ExitCode == 0 {
		result["useradd"] = true
	}

	ctx, cancel3 := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel3()

	userDelCmd := command.New(ctx, "which", []string{"userdel"})
	output, err = userDelCmd.Run()
	if err != nil {
		return result, err
	}
	if output.Status.ExitCode == 0 {
		result["userdel"] = true
	}

	ctx, cancel4 := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel4()

	delUserCmd := command.New(ctx, "which", []string{"deluser"})
	output, err = delUserCmd.Run()
	if err != nil {
		return result, err
	}
	if output.Status.ExitCode == 0 {
		result["deluser"] = true
	}

	return result, nil
}

func CheckGroupCoreBinaries(timeout int) (map[string]bool, error) {
	//
	result := map[string]bool{
		"groupadd": false,
		"addgroup": false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	groupAddCmd := command.New(ctx, "which", []string{"groupadd"})
	output, err := groupAddCmd.Run()
	if err != nil {
		return result, err
	}

	if output.Status.ExitCode == 0 {
		result["groupadd"] = true
	}

	ctx, cancel2 := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel2()

	addGroupCmd := command.New(ctx, "which", []string{"addgroup"})
	output, err = addGroupCmd.Run()
	if err != nil {
		return result, err
	}

	if output.Status.ExitCode == 0 {
		result["addgroup"] = true
	}

	return result, nil
}
