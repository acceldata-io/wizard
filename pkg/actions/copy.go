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
	"crypto/sha256"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"github.com/acceldata-io/wizard/internal/parser"
	"github.com/acceldata-io/wizard/pkg/register"
	"github.com/acceldata-io/wizard/pkg/wlog"

	"github.com/go-playground/validator/v10"
)

const BackupDir = "/tmp/backup"

// A Hash is a directory hash function.
// It accepts a list of files along with a function that opens the content of each file.
// It opens, reads, hashes, and closes each file and returns the overall directory hash.
type Hash func(files []string, open func(string) (io.ReadCloser, error)) (string, error)

type copyAction struct {
	agentName string
	timeout   int
	register  string
}

type copyVars struct {
	SourceType  string `json:"src_type" validate:"required"`
	Source      string `json:"src" validate:"required"`
	SourceFiles []string
	Destination string `json:"dest" validate:"required"`
	Permission  string `json:"permission" validate:"required"`
	Owner       string `json:"owner" validate:"required"`
	Group       string `json:"group" validate:"required"`
	Force       bool   `json:"force"`
	Backup      bool   `json:"backup"`
	Parents     bool   `json:"parents"`
	Recursive   bool   `json:"recursive"`
	IsDestDir   bool
}

func NewCopyAction(agentName string, timeout int, localRegister string) Action {
	return &copyAction{agentName: agentName, timeout: timeout, register: localRegister}
}

func newCopyVars(data map[string]interface{}) (*copyVars, error) {
	copyConfig := copyVars{}

	if dataB, err := json.Marshal(data); err == nil {
		if err := json.Unmarshal(dataB, &copyConfig); err != nil {
			return &copyConfig, err
		}
	} else {
		return &copyConfig, err
	}

	validate := validator.New()
	err := validate.Struct(copyConfig)
	if err != nil {
		return &copyConfig, err
	}

	if copyConfig.SourceType == "embed" {
		copyConfig.SourceFiles, err = fs.Glob(PackageFiles, copyConfig.Source)
	} else if copyConfig.SourceType == "local" {
		copyConfig.SourceFiles, err = filepath.Glob(copyConfig.Source)
	} else {
		return nil, fmt.Errorf("wrong source type")
	}
	if err != nil {
		return nil, fmt.Errorf("unable to get source file(s) for %s - %s", copyConfig.Source, err)
	}

	if strings.Contains(copyConfig.Source, "*") {
		copyConfig.IsDestDir = true
	}

	return &copyConfig, nil
}

func (c *copyAction) BackupConfigFile(filePath, srcType string) (string, error) {
	err := os.MkdirAll(BackupDir+"/"+c.agentName, 0o755)
	if err != nil {
		return "", fmt.Errorf("BackupConfigFile: unable to create backup dir - %s", err)
	}
	_, fileName := filepath.Split(filePath)
	err = CopyFile(filePath, BackupDir+"/"+c.agentName+"/"+fileName, "0644", srcType)
	if err != nil {
		return "", fmt.Errorf("BackupConfigFile: %s", err)
	}
	return BackupDir + "/" + c.agentName + "/" + fileName, nil
}

func (c *copyAction) Do(actions *parser.Action, wizardLog chan interface{}) error {
	cRegister := register.RMap[c.register]

	wizardLog <- wlog.WLInfo("running when condition")
	if actions.When != nil {
		when := NewWhen(actions.When.Command, actions.When.RVar, actions.When.ExitCode, c.timeout)
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

	copyConfig, err := newCopyVars(actions.ActionVariables)
	if err != nil {
		wizardLog <- wlog.WLError("validation error: " + err.Error())
		return err
	}

	isSrcDir := false
	isDestDir := false
	var uid, gid int

	isDestDir = copyConfig.IsDestDir
	masterDest := copyConfig.Destination

	for _, src := range copyConfig.SourceFiles {
		copyConfig.Destination = masterDest

		if copyConfig.SourceType == "embed" {
			wizardLog <- wlog.WLInfo("Identified source type as embed")
			if stat, err := fs.Stat(PackageFiles, src); err == nil && stat.IsDir() {
				isSrcDir = true
			} else if os.IsNotExist(err) {
				return fmt.Errorf("source directory - %s does not exists - %s", src, copyConfig.SourceType)
			}
		} else if copyConfig.SourceType == "local" {
			wizardLog <- wlog.WLInfo("Identified source type as local")
			if stat, err := os.Stat(src); err == nil && stat.IsDir() {
				isSrcDir = true
			} else if os.IsNotExist(err) {
				return fmt.Errorf("source directory - %s does not exists - %s", src, copyConfig.SourceType)
			}
		}

		if isSrcDir && copyConfig.IsDestDir { // if master src has * and looped src is dir
			copyConfig.Destination = copyConfig.Destination + "/" + filepath.Base(src)
		} else if copyConfig.IsDestDir { // if looped src is not dir but master src has *
			if copyConfig.Recursive {
				srcDir := strings.Split(filepath.Dir(src), "/")
				dir := strings.Join(srcDir[1:], "/")
				copyConfig.Destination = copyConfig.Destination + "/" + dir
			}
		}

		if stat, err := os.Stat(copyConfig.Destination); err == nil && stat.IsDir() {
			isDestDir = true
		} else if os.IsNotExist(err) {
			if copyConfig.Parents || isDestDir {
				var path string
				if isSrcDir || isDestDir {
					path = copyConfig.Destination
					isDestDir = true
				} else {
					path = filepath.Dir(copyConfig.Destination)
				}
				err := CreateIfNotExists(path, copyConfig.Permission)
				if err != nil {
					if isSrcDir {
						return fmt.Errorf("unable to create dir - %s for src - %s - %s", path, src, err)
					}
					return fmt.Errorf("unable to create parent dir - %s for src - %s - %s", path, src, err)
				}
			} else if isSrcDir {
				return fmt.Errorf("destination dir - %s not found: %s", copyConfig.Destination, err)
			} else if _, err := os.Stat(filepath.Dir(copyConfig.Destination)); err != nil && os.IsNotExist(err) {
				return fmt.Errorf("destination parent dir - %s not found: %s", filepath.Dir(copyConfig.Destination), err)
			}
		} else if err != nil {
			return fmt.Errorf("unknown error occured: %s", err)
		}

		if isSrcDir && isDestDir {
			// COPY DIRECTORY TO DIRECTORY
			wizardLog <- wlog.WLInfo("Identified copy dir to dir: " + src + " to" + copyConfig.Destination)
			srcDirHash, err := hashDir(src, "", hash1, copyConfig.SourceType)
			if err != nil {
				return fmt.Errorf("src dir hashing error - %s", err.Error())
			}

			// TODO -Ignoring any error for now - But should only ignore file not found error
			destDirHash, _ := hashDir(copyConfig.Destination, "", hash1, copyConfig.SourceType)

			if destDirHash != srcDirHash {
				wizardLog <- wlog.WLInfo("hash not matched, copying dir to dir: " + src + " to" + copyConfig.Destination)
				err := CopyDirectory(src, copyConfig.Destination, copyConfig.Permission, copyConfig.Owner, copyConfig.Group, copyConfig.SourceType)
				if err != nil {
					return err
				}
				cRegister.Changed = true
			} else {
				wizardLog <- wlog.WLInfo("dir hash matched, not copying")
			}
		}

		if isSrcDir && !isDestDir {
			// CANNOT COPY
			wizardLog <- wlog.WLError("cannot copy a directory into a file")
			return fmt.Errorf("cannot copy a directory into a file")
		}

		if !isSrcDir && isDestDir {
			// Copy file into a directory with same name
			wizardLog <- wlog.WLInfo("Identified copy file to dir: " + src + " to" + copyConfig.Destination)
			backUp, err := c.BackupConfigFile(src, copyConfig.SourceType)
			if err != nil {
				return err
			}
			actions.BackupSrc = backUp
			_, fileName := filepath.Split(src)
			copyConfig.Destination = copyConfig.Destination + "/" + fileName
			if _, err := os.Stat(copyConfig.Destination); os.IsNotExist(err) {
				wizardLog <- wlog.WLInfo("file not found at destination, copying file to dir: " + src + " to" + copyConfig.Destination)
				// Normal Copy, no hash to be checked and update the changed status
				// Directory exists but file not exist
				if err := CopyFile(src, copyConfig.Destination, copyConfig.Permission, copyConfig.SourceType); err != nil {
					return err
				}
				// Change the group and owner of the file
				fileUser, err := user.Lookup(copyConfig.Owner)
				if err != nil {
					return err
				}
				uid, err = strconv.Atoi(fileUser.Uid)
				if err != nil {
					wizardLog <- wlog.WLInfo("unable to convert the uid to int. Because: " + err.Error())
					return err
				}
				if strings.TrimSpace(copyConfig.Group) == "" {
					gid, err = strconv.Atoi(fileUser.Gid)
					if err != nil {
						wizardLog <- wlog.WLInfo("unable to convert the gid to int. Because: " + err.Error())
						return err
					}
				} else {
					fileGroup, err := user.LookupGroup(copyConfig.Group)
					if err != nil {
						return err
					}
					gid, err = strconv.Atoi(fileGroup.Gid)
					if err != nil {
						wizardLog <- wlog.WLInfo("unable to convert the gid to int. Because: " + err.Error())
						return err
					}
				}
				err = os.Chown(copyConfig.Destination, uid, gid)
				if err != nil {
					return fmt.Errorf("unable to change owner to file - %s, with uid - %d, gid - %d, - %s", copyConfig.Destination, uid, gid, err)
				}
				cRegister.Changed = true
			} else if err == nil {
				// Check overwrite func and the hash and update in condition the changed status
				wizardLog <- wlog.WLInfo("file found at destination, checking hash")
				if GetHashOfFile(src) == GetHashOfFile(copyConfig.Destination) {
					// No need to change the file
					if copyConfig.Force {
						wizardLog <- wlog.WLInfo("hash match, but force is true, copying file to dir: " + src + " to" + copyConfig.Destination)
						if err = CopyFile(src, copyConfig.Destination, copyConfig.Permission, copyConfig.SourceType); err != nil {
							return err
						}
						// Change the group and owner of the file
						fileUser, err := user.Lookup(copyConfig.Owner)
						if err != nil {
							return err
						}
						uid, err = strconv.Atoi(fileUser.Uid)
						if err != nil {
							wizardLog <- wlog.WLInfo("unable to convert the uid to int. Because: " + err.Error())
							return err
						}
						if strings.TrimSpace(copyConfig.Group) == "" {
							gid, err = strconv.Atoi(fileUser.Gid)
							if err != nil {
								wizardLog <- wlog.WLInfo("unable to convert the gid to int. Because: " + err.Error())
								return err
							}
						} else {
							fileGroup, err := user.LookupGroup(copyConfig.Group)
							if err != nil {
								return err
							}
							gid, err = strconv.Atoi(fileGroup.Gid)
							if err != nil {
								wizardLog <- wlog.WLInfo("unable to convert the gid to int. Because: " + err.Error())
								return err
							}
						}
						err = os.Chown(copyConfig.Destination, uid, gid)
						if err != nil {
							return fmt.Errorf("unable to change owner to file - %s, with uid - %d, gid - %d, - %s", copyConfig.Destination, uid, gid, err)
						}
						cRegister.Changed = true
					}
				} else {
					wizardLog <- wlog.WLInfo("hash not match, copying file to dir: " + src + " to" + copyConfig.Destination)
					if err = CopyFile(src, copyConfig.Destination, copyConfig.Permission, copyConfig.SourceType); err != nil {
						return err
					}
					// Change the group and owner of the file
					fileUser, err := user.Lookup(copyConfig.Owner)
					if err != nil {
						return err
					}
					uid, err = strconv.Atoi(fileUser.Uid)
					if err != nil {
						wizardLog <- wlog.WLInfo("unable to convert the uid to int. Because: " + err.Error())
						return err
					}
					if strings.TrimSpace(copyConfig.Group) == "" {
						gid, err = strconv.Atoi(fileUser.Gid)
						if err != nil {
							wizardLog <- wlog.WLInfo("unable to convert the gid to int. Because: " + err.Error())
							return err
						}
					} else {
						fileGroup, err := user.LookupGroup(copyConfig.Group)
						if err != nil {
							return err
						}
						gid, err = strconv.Atoi(fileGroup.Gid)
						if err != nil {
							wizardLog <- wlog.WLInfo("unable to convert the gid to int. Because: " + err.Error())
							return err
						}
					}
					err = os.Chown(copyConfig.Destination, uid, gid)
					if err != nil {
						return fmt.Errorf("unable to change owner to file - %s, with uid - %d, gid - %d, - %s", copyConfig.Destination, uid, gid, err)
					}
					cRegister.Changed = true
				}
			}
		}

		if !isSrcDir && !isDestDir {
			// FILE TO FILE and glob also
			wizardLog <- wlog.WLInfo("Identified, copy file to file: " + src + " to" + copyConfig.Destination)
			if _, err := os.Stat(copyConfig.Destination); os.IsNotExist(err) {
				// Directory exists but file not exist
				wizardLog <- wlog.WLInfo("file not found at destination, copying file to file: " + src + " to" + copyConfig.Destination)
				actions.BackupSrc = src
				if err := CopyFile(src, copyConfig.Destination, copyConfig.Permission, copyConfig.SourceType); err != nil {
					return err
				}
				// Change the group and owner of the file
				fileUser, err := user.Lookup(copyConfig.Owner)
				if err != nil {
					return err
				}
				uid, err = strconv.Atoi(fileUser.Uid)
				if err != nil {
					wizardLog <- wlog.WLInfo("unable to convert the uid to int. Because: " + err.Error())
					return err
				}
				if strings.TrimSpace(copyConfig.Group) == "" {
					gid, err = strconv.Atoi(fileUser.Gid)
					if err != nil {
						wizardLog <- wlog.WLInfo("unable to convert the gid to int. Because: " + err.Error())
						return err
					}
				} else {
					fileGroup, err := user.LookupGroup(copyConfig.Group)
					if err != nil {
						return err
					}
					gid, err = strconv.Atoi(fileGroup.Gid)
					if err != nil {
						wizardLog <- wlog.WLInfo("unable to convert the gid to int. Because: " + err.Error())
						return err
					}
				}
				err = os.Chown(copyConfig.Destination, uid, gid)
				if err != nil {
					return fmt.Errorf("unable to change owner to file - %s, with uid - %d, gid - %d, - %s", copyConfig.Destination, uid, gid, err)
				}
				cRegister.Changed = true
			} else if err == nil {
				// Check overwrite func and the hash and update in condition the changed status
				wizardLog <- wlog.WLInfo("file found at destination, checking hash")
				backup, err := c.BackupConfigFile(copyConfig.Destination, "local")
				if err != nil {
					return err
				}
				actions.BackupSrc = backup
				if GetHashOfFile(src) == GetHashOfFile(copyConfig.Destination) {
					// No need to change the file
					if copyConfig.Force {
						wizardLog <- wlog.WLInfo("hash matched, but force is true, copying file to file: " + src + " to" + copyConfig.Destination)
						if err = CopyFile(src, copyConfig.Destination, copyConfig.Permission, copyConfig.SourceType); err != nil {
							return err
						}
						// Change the group and owner of the file
						fileUser, err := user.Lookup(copyConfig.Owner)
						if err != nil {
							return err
						}
						uid, err = strconv.Atoi(fileUser.Uid)
						if err != nil {
							wizardLog <- wlog.WLInfo("unable to convert the uid to int. Because: " + err.Error())
							return err
						}
						if strings.TrimSpace(copyConfig.Group) == "" {
							gid, err = strconv.Atoi(fileUser.Gid)
							if err != nil {
								wizardLog <- wlog.WLInfo("unable to convert the gid to int. Because: " + err.Error())
								return err
							}
						} else {
							fileGroup, err := user.LookupGroup(copyConfig.Group)
							if err != nil {
								return err
							}
							gid, err = strconv.Atoi(fileGroup.Gid)
							if err != nil {
								wizardLog <- wlog.WLInfo("unable to convert the gid to int. Because: " + err.Error())
								return err
							}
						}
						err = os.Chown(copyConfig.Destination, uid, gid)
						if err != nil {
							return fmt.Errorf("unable to change owner to file - %s, with uid - %d, gid - %d, - %s", copyConfig.Destination, uid, gid, err)
						}
						cRegister.Changed = true
					}
				} else {
					wizardLog <- wlog.WLInfo("hash not matched, copying file to file: " + src + " to" + copyConfig.Destination)
					backup, err := c.BackupConfigFile(copyConfig.Destination, "local")
					if err != nil {
						return err
					}
					actions.BackupSrc = backup
					if err = CopyFile(src, copyConfig.Destination, copyConfig.Permission, copyConfig.SourceType); err != nil {
						return err
					}
					// Change the group and owner of the file
					fileUser, err := user.Lookup(copyConfig.Owner)
					if err != nil {
						return err
					}
					uid, err = strconv.Atoi(fileUser.Uid)
					if err != nil {
						wizardLog <- wlog.WLInfo("unable to convert the uid to int. Because: " + err.Error())
						return err
					}
					if strings.TrimSpace(copyConfig.Group) == "" {
						gid, err = strconv.Atoi(fileUser.Gid)
						if err != nil {
							wizardLog <- wlog.WLInfo("unable to convert the gid to int. Because: " + err.Error())
							return err
						}
					} else {
						fileGroup, err := user.LookupGroup(copyConfig.Group)
						if err != nil {
							return err
						}
						gid, err = strconv.Atoi(fileGroup.Gid)
						if err != nil {
							wizardLog <- wlog.WLInfo("unable to convert the gid to int. Because: " + err.Error())
							return err
						}
					}
					err = os.Chown(copyConfig.Destination, uid, gid)
					if err != nil {
						return fmt.Errorf("unable to change owner to file - %s, with uid - %d, gid - %d, - %s", copyConfig.Destination, uid, gid, err)
					}
					cRegister.Changed = true
				} // If action changed to true then restore
			}
		}
	}

	return nil
}

func Restore(src, dest, perm, srcType string) error {
	err := CopyFile(src, dest, perm, srcType)
	if err != nil {
		return err
	}
	return nil
}

func CopyFile(src, dest, perm, srcType string) error {
	var input []byte
	var err error

	if srcType == "embed" {
		input, err = fs.ReadFile(PackageFiles, src)
	} else if srcType == "local" {
		input, err = os.ReadFile(src)
	} else {
		return fmt.Errorf("CopyFile: wrong source type")
	}
	if err != nil {
		return fmt.Errorf("CopyFile: unable to read file - %s from %s - %s", src, srcType, err)
	}
	permission, err := strconv.ParseInt(perm, 8, 32)
	if err != nil {
		return fmt.Errorf("CopyFile: unable to parse permission: %s to int. Because: %s", perm, err.Error())
	}
	err = os.WriteFile(dest, input, os.FileMode(permission))
	if err != nil {
		return fmt.Errorf("CopyFile: unable to write file to - %s from %s - %s", dest, srcType, err)
	}
	return nil
}

func GetHashOfFile(filePath string) string {
	hash := sha256.New()
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}

	// Tell the program to call the following function when the current function returns
	defer file.Close()
	// Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return "0"
	}
	return string(hash.Sum(nil))
}

func CopyDirectory(srcDir, dest string, perm string, owner string, group string, srcType string) error {
	var gid, uid int

	if srcType == "embed" {
		entries, err := fs.ReadDir(PackageFiles, srcDir)
		if err != nil {
			return fmt.Errorf("CopyDirectory: Unable to read dir - %s from %s - %s", srcDir, srcType, err)
		}

		for _, entry := range entries {
			sourcePath := filepath.Join(srcDir, entry.Name())
			destPath := filepath.Join(dest, entry.Name())

			fileInfo, err := fs.Stat(PackageFiles, sourcePath)
			if err != nil {
				return fmt.Errorf("CopyDirectory: Unable to get stat for %s from %s - %s", sourcePath, srcType, err)
			}

			switch fileInfo.Mode() & os.ModeType {
			case os.ModeDir:
				if err := CreateIfNotExists(destPath, perm); err != nil {
					return fmt.Errorf("CopyDirectory: %s", err)
				}
				if err := CopyDirectory(sourcePath, destPath, perm, owner, group, srcType); err != nil {
					return fmt.Errorf("CopyDirectory: %s", err)
				}
			case os.ModeSymlink:
				if err := CopySymLink(sourcePath, destPath); err != nil {
					return fmt.Errorf("CopyDirectory: %s", err)
				}
			default:
				if err := CopyFile(sourcePath, destPath, perm, srcType); err != nil {
					return fmt.Errorf("CopyDirectory: %s", err)
				}
			}

			isSymlink := fileInfo.Mode()&os.ModeSymlink != 0
			if !isSymlink {
				if err := os.Chmod(destPath, fileInfo.Mode()); err != nil {
					return fmt.Errorf("CopyDirectory: Unable to give permission to file - %s - %s", destPath, err)
				}
			}
			fileUser, err := user.Lookup(owner)
			if err != nil {
				return fmt.Errorf("CopyDirectory: %s", err)
			}
			uid, err = strconv.Atoi(fileUser.Uid)
			if err != nil {
				return err
			}
			if strings.TrimSpace(group) == "" {
				gid, err = strconv.Atoi(fileUser.Gid)
				if err != nil {
					return err
				}
			} else {
				fileGroup, err := user.LookupGroup(group)
				if err != nil {
					return fmt.Errorf("CopyDirectory: %s", err)
				}
				gid, err = strconv.Atoi(fileGroup.Gid)
				if err != nil {
					return err
				}
			}
			err = os.Chown(destPath, uid, gid)
			if err != nil {
				return fmt.Errorf("CopyDirectory: Unable to change owner to file - %s, with uid - %d, gid - %d, - %s", destPath, uid, gid, err)
			}
		}
	} else if srcType == "local" {
		entries, err := ioutil.ReadDir(srcDir)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			sourcePath := filepath.Join(srcDir, entry.Name())
			destPath := filepath.Join(dest, entry.Name())

			fileInfo, err := os.Stat(sourcePath)
			if err != nil {
				return err
			}

			stat, ok := fileInfo.Sys().(*syscall.Stat_t)
			if !ok {
				return fmt.Errorf("CopyDirectory: failed to get raw syscall.Stat_t data for '%s'", sourcePath)
			}

			switch fileInfo.Mode() & os.ModeType {
			case os.ModeDir:
				if err := CreateIfNotExists(destPath, perm); err != nil {
					return fmt.Errorf("CopyDirectory: %s", err)
				}
				if err := CopyDirectory(sourcePath, destPath, perm, owner, group, srcType); err != nil {
					return fmt.Errorf("CopyDirectory: %s", err)
				}
			case os.ModeSymlink:
				if err := CopySymLink(sourcePath, destPath); err != nil {
					return fmt.Errorf("CopyDirectory: %s", err)
				}
			default:
				if err := CopyFile(sourcePath, destPath, perm, srcType); err != nil {
					return fmt.Errorf("CopyDirectory: %s", err)
				}
			}

			if err := os.Lchown(destPath, int(stat.Uid), int(stat.Gid)); err != nil {
				return err
			}

			isSymlink := entry.Mode()&os.ModeSymlink != 0
			if !isSymlink {
				if err := os.Chmod(destPath, entry.Mode()); err != nil {
					return fmt.Errorf("CopyDirectory: Unable to give permission to file - %s - %s", destPath, err)
				}
			}
			fileUser, err := user.Lookup(owner)
			if err != nil {
				return fmt.Errorf("CopyDirectory: %s", err)
			}
			uid, _ = strconv.Atoi(fileUser.Uid)
			if strings.TrimSpace(group) == "" {
				gid, _ = strconv.Atoi(fileUser.Gid)
			} else {
				fileGroup, err := user.LookupGroup(group)
				if err != nil {
					return fmt.Errorf("CopyDirectory: %s", err)
				}
				gid, _ = strconv.Atoi(fileGroup.Gid)
			}
			err = os.Chown(destPath, uid, gid)
			if err != nil {
				return fmt.Errorf("CopyDirectory: Unable to change owner to file - %s, with uid - %d, gid - %d, - %s", destPath, uid, gid, err)
			}
		}
	} else {
		return fmt.Errorf("CopyDirectory: wrong source type")
	}

	return nil
}

func Exists(filePath string) bool {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return false
	}

	return true
}

func CreateIfNotExists(dir string, perm string) error {
	permission, _ := strconv.ParseInt(perm, 8, 32)
	if Exists(dir) {
		return nil
	}

	if err := os.MkdirAll(dir, os.FileMode(permission)); err != nil {
		return fmt.Errorf("CreateIfNotExists: failed to create directory: '%s', error: '%s'", dir, err.Error())
	}

	return nil
}

func CopySymLink(source, dest string) error {
	link, err := os.Readlink(source)
	if err != nil {
		return fmt.Errorf("CopySymLink: %s", err)
	}
	return os.Symlink(link, dest)
}

func hash1(files []string, open func(string) (io.ReadCloser, error)) (string, error) {
	h := sha256.New()
	files = append([]string(nil), files...)
	sort.Strings(files)
	for _, file := range files {
		if strings.Contains(file, "\n") {
			return "", errors.New("dirhash: filenames with newlines are not supported")
		}
		r, err := open(file)
		if err != nil {
			return "", err
		}
		hf := sha256.New()
		_, err = io.Copy(hf, r)
		r.Close()
		if err != nil {
			return "", err
		}
		fmt.Fprintf(h, "%x  %s\n", hf.Sum(nil), file)
	}
	return "h1:" + base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}

// hashDir returns the hash of the local file system directory dir,
// replacing the directory name itself with prefix in the file names
// used in the hash function.
func hashDir(dir, prefix string, hash Hash, srcType string) (string, error) {
	if srcType == "local" {
		files, err := dirFiles(dir, prefix)
		if err != nil {
			return "", err
		}
		osOpen := func(name string) (io.ReadCloser, error) {
			return os.Open(filepath.Join(dir, strings.TrimPrefix(name, prefix)))
		}
		return hash(files, osOpen)
	}
	files, err := dirFilesEmbed(&PackageFiles, dir)
	if err != nil {
		return "", err
	}
	osOpen := func(name string) (io.ReadCloser, error) {
		return PackageFiles.Open(name)
	}
	return hash(files, osOpen)
}

func dirFilesEmbed(efs *embed.FS, path string) (out []string, err error) {
	if len(path) == 0 {
		path = "."
	}
	entries, err := efs.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		fp := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			res, err := dirFilesEmbed(efs, fp)
			if err != nil {
				return nil, err
			}
			out = append(out, res...)
			continue
		}
		out = append(out, fp)
	}
	return
}

// dirFiles returns the list of files in the tree rooted at dir,
// replacing the directory name dir with prefix in each name.
// The resulting names always use forward slashes.
func dirFiles(dir, prefix string) ([]string, error) {
	var files []string
	dir = filepath.Clean(dir)
	err := filepath.Walk(dir, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel := file
		if dir != "." {
			rel = file[len(dir)+1:]
		}
		f := filepath.Join(prefix, rel)
		files = append(files, filepath.ToSlash(f))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
