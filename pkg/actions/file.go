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
	"os"
	"os/user"
	"strconv"

	"github.com/acceldata-io/wizard/internal/parser"
	"github.com/acceldata-io/wizard/pkg/register"
	"github.com/acceldata-io/wizard/pkg/wlog"

	"github.com/go-playground/validator/v10"
)

type fileAction struct {
	agentName string
	timeout   int
	register  string
}

func NewFileAction(agentName string, timeout int, register string) Action {
	return &fileAction{agentName: agentName, timeout: timeout, register: register}
}

type fileVar struct {
	Files      []fileInfo `json:"files" validate:"required"`
	Dir        bool       `json:"dir"`
	State      string     `json:"state" validate:"required"`
	Permission string     `json:"permission" validate:"required"`
	Owner      string     `json:"owner" validate:"required"`
	Group      string     `json:"group" validate:"required"`
	Enforce    bool       `json:"force"`
}

type fileInfo struct {
	Source      string `json:"src"`
	Destination string `json:"dest"`
}

func NewFileVar(data map[string]interface{}) (*fileVar, error) {
	t := fileVar{}

	if dataB, err := json.Marshal(data); err == nil {
		if err := json.Unmarshal(dataB, &t); err != nil {
			return &t, err
		}
	} else {
		return &t, err
	}

	validate := validator.New()
	err := validate.Struct(t)
	if err != nil {
		return &t, err
	}

	return &t, nil
}

func (f *fileAction) Do(actions *parser.Action, wizardLog chan interface{}) error {
	// Create/Deleting a file and directory
	fRegister := register.RMap[f.register]

	wizardLog <- wlog.WLInfo("executing when condition")
	if actions.When != nil {
		when := NewWhen(actions.When.Command, actions.When.RVar, actions.When.ExitCode, f.timeout)
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
	touchVar, err := NewFileVar(actions.ActionVariables)
	if err != nil {
		wizardLog <- wlog.WLError("validation error: " + err.Error())
		return err
	}
	switch touchVar.State {
	case "touch":
		if touchVar.Dir {
			wizardLog <- wlog.WLInfo("touching a dir")
			err := touchVar.TouchDir()
			if err != nil {
				wizardLog <- wlog.WLError(err.Error())
				return err
			}
		} else {
			wizardLog <- wlog.WLInfo("touching a file")
			err := touchVar.TouchFile()
			if err != nil {
				wizardLog <- wlog.WLError(err.Error())
				return err
			}
		}
	case "link":
		wizardLog <- wlog.WLInfo("creating a symlink")
		err := touchVar.CreateSymlink()
		if err != nil {
			wizardLog <- wlog.WLError(err.Error())
			return err
		}
	case "absent":
		if touchVar.Dir {
			wizardLog <- wlog.WLInfo("removing a dir")
			err := touchVar.RemoveDir()
			if err != nil {
				wizardLog <- wlog.WLError(err.Error())
				return err
			}
		} else {
			wizardLog <- wlog.WLInfo("removing a file")
			err := touchVar.RemoveFile()
			if err != nil {
				wizardLog <- wlog.WLError(err.Error())
				return err
			}
		}
	default:
		wizardLog <- wlog.WLError(fmt.Sprintf("unknown state: %s", touchVar.State))
		return fmt.Errorf("unknown state: %s", touchVar.State)
	}
	fRegister.Changed = true
	return nil
}

func (t *fileVar) CreateSymlink() error {
	for _, file := range t.Files {
		if _, err := os.Stat(file.Source); os.IsNotExist(err) {
			return fmt.Errorf("CreateSymlink: source file does not exists - %s", err.Error())
		} else if err != nil {
			return fmt.Errorf("CreateSymLink: source file not accessible - %s", err.Error())
		}

		err := os.Symlink(file.Source, file.Destination)
		if err != nil {
			if os.IsExist(err) {
				// removing old symlink and creating new
				rf := fileVar{
					Files: []fileInfo{
						{
							Destination: file.Destination,
						},
					},
				}
				err = rf.RemoveFile()
				if err != nil {
					return fmt.Errorf("CreateSymlink: unable to remove old symlink %s", err)
				}

				err = os.Symlink(file.Source, file.Destination)
				if err != nil {
					return fmt.Errorf("CreateSymlink: unable to create symlink - %s", err)
				}
			} else {
				return fmt.Errorf("CreateSymlink: unable to create symlink - %s", err)
			}
		}
	}

	return nil
}

func (t *fileVar) TouchFile() error {
	// touch file
	for _, file := range t.Files {
		if _, err := os.Stat(file.Destination); os.IsNotExist(err) {
			f, err := os.Create(file.Destination)
			if err != nil {
				return fmt.Errorf("TouchFile: Unable to create file at dest - %s - %s", file.Destination, err)
			}
			err = f.Close()
			if err != nil {
				return fmt.Errorf("TouchFile: Unable to close file - %s", err)
			}
			permission, err := strconv.ParseInt(t.Permission, 8, 32)
			if err != nil {
				return fmt.Errorf("TouchFile: unable to parse permission: %s to int. Because: %s", t.Permission, err.Error())
			}
			err = os.Chmod(file.Destination, os.FileMode(permission))
			if err != nil {
				return fmt.Errorf("TouchFile: Unable to give permission to file - %s - %s", file.Destination, err)
			}
			fileUser, err := user.Lookup(t.Owner)
			if err != nil {
				return fmt.Errorf("TouchFile: %s", err)
			}
			uid, err := strconv.Atoi(fileUser.Uid)
			if err != nil {
				return fmt.Errorf("TouchFile: unable to convert the uid to int. Because: " + err.Error())
			}
			gid, err := strconv.Atoi(fileUser.Gid)
			if err != nil {
				return fmt.Errorf("unable to convert the gid to int. Because: " + err.Error())
			}
			err = os.Chown(file.Destination, uid, gid)
			if err != nil {
				return fmt.Errorf("TouchFile: Unable to change owner to file - %s, with uid - %d, gid - %d, - %s", file.Destination, uid, gid, err)
			}
		} else if err != nil {
			return fmt.Errorf("TouchFile: %s", err)
		} else {
			if t.Enforce {
				permission, err := strconv.ParseInt(t.Permission, 8, 32)
				if err != nil {
					return fmt.Errorf("TouchFile: unable to parse permission: %s to int. Because: %s", t.Permission, err.Error())
				}
				err = os.Chmod(file.Destination, os.FileMode(permission))
				if err != nil {
					return fmt.Errorf("TouchFile: Unable to give permission to file - %s - %s", file.Destination, err)
				}
				fileUser, err := user.Lookup(t.Owner)
				if err != nil {
					return fmt.Errorf("TouchFile: %s", err)
				}
				uid, err := strconv.Atoi(fileUser.Uid)
				if err != nil {
					return fmt.Errorf("TouchFile: unable to convert the uid to int. Because: " + err.Error())
				}
				gid, err := strconv.Atoi(fileUser.Gid)
				if err != nil {
					return fmt.Errorf("TouchFile: unable to convert the gid to int. Because: " + err.Error())
				}
				err = os.Chown(file.Destination, uid, gid)
				if err != nil {
					return fmt.Errorf("TouchFile: Unable to change owner to file - %s, with uid - %d, gid - %d, - %s", file.Destination, uid, gid, err)
				}
			}
		}
	}

	return nil
}

func (t *fileVar) RemoveFile() error {
	// remove file
	for _, file := range t.Files {
		if _, err := os.Stat(file.Destination); err == nil {
			if err := os.RemoveAll(file.Destination); err != nil {
				return fmt.Errorf("RemoveFile: Unable to remove file - %s", err)
			}
		}
	}

	return nil
}

func (t *fileVar) TouchDir() error {
	permission, err := strconv.ParseInt(t.Permission, 8, 32)
	if err != nil {
		return fmt.Errorf("TouchDir: unable to parse permission: %s to int. Because: %s", t.Permission, err.Error())
	}

	for _, dir := range t.Files {
		err := os.MkdirAll(dir.Destination, os.FileMode(permission))
		if err != nil {
			return err
		}
		permission, err := strconv.ParseInt(t.Permission, 8, 32)
		if err != nil {
			return fmt.Errorf("TouchDir: unable to parse permission: %s to int. Because: %s", t.Permission, err.Error())
		}
		err = os.Chmod(dir.Destination, os.FileMode(permission))
		if err != nil {
			return fmt.Errorf("TouchDir: Unable to give permission to dir - %s - %s", dir.Destination, err)
		}
		fileUser, err := user.Lookup(t.Owner)
		if err != nil {
			return fmt.Errorf("TouchDir: %s", err)
		}
		uid, err := strconv.Atoi(fileUser.Uid)
		if err != nil {
			return fmt.Errorf("TouchDir: unable to convert the uid to int. Because: " + err.Error())
		}
		gid, err := strconv.Atoi(fileUser.Gid)
		if err != nil {
			return fmt.Errorf("TouchDir: unable to convert the gid to int. Because: " + err.Error())
		}
		err = os.Chown(dir.Destination, uid, gid)
		if err != nil {
			return fmt.Errorf("TouchDir: Unable to change owner to dir - %s, with uid - %d, gid - %d, - %s", dir.Destination, uid, gid, err)
		}
	}

	return nil
}

func (t *fileVar) RemoveDir() error {
	for _, dir := range t.Files {
		if _, err := os.Stat(dir.Destination); err == nil {
			err := os.RemoveAll(dir.Destination)
			if err != nil {
				return fmt.Errorf("RemoveDir: Unable to remove dir - %s", err)
			}
		}
	}

	return nil
}
