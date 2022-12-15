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
	"fmt"
	"os"
	"strconv"
	"strings"
)

type unixUser struct {
	username string
	password string
	uid      int
	gid      int
	info     string
	homedir  string
	shell    string
}

type unixGroup struct {
	groupName string
	password  string
	gid       int
	groupList []string
}

func userLookup(username string) (unixUser, error) {
	userFound := unixUser{}
	userList, err := userInfoList("/etc/passwd")
	if err != nil {
		return userFound, fmt.Errorf("userLookup: %s", err)
	}
	for _, unixUser := range userList {
		if unixUser.username == username {
			userFound = unixUser
		}
	}
	return userFound, nil
}

func groupInfoLookUp(path string) ([]unixGroup, error) {
	var groupList []unixGroup
	content, err := os.ReadFile(path)
	if err != nil {
		return groupList, fmt.Errorf("GroupInfoLookUp: unable to read file - %s - %s", path, err)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	groupList = make([]unixGroup, len(lines))
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(strings.TrimSpace(line), ":")
		if len(parts) != 4 {
			return groupList, fmt.Errorf("GroupInfoLookUp: Group line had wrong number of parts %d != 7", len(parts))
		}
		groupList[i].groupName = strings.TrimSpace(parts[0])
		groupList[i].password = strings.TrimSpace(parts[1])

		gid, err := strconv.Atoi(parts[2])
		if err != nil {
			return groupList, fmt.Errorf("GroupInfoLookUp: Group line had badly formatted gid %s", parts[2])
		}
		groupList[i].gid = gid
		groupList[i].groupList = strings.Split(parts[3], ",")
	}

	return groupList, nil
}

func userInfoList(path string) ([]unixUser, error) {
	var userList []unixUser
	content, err := os.ReadFile(path)
	if err != nil {
		return userList, fmt.Errorf("userInfoList: unable to read file - %s - %s", path, err)
	}
	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	userList = make([]unixUser, len(lines))
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(strings.TrimSpace(line), ":")
		if len(parts) != 7 {
			return userList, fmt.Errorf("userInfoList: Passwd line had wrong number of parts %d != 7", len(parts))
		}
		userList[i].username = strings.TrimSpace(parts[0])
		userList[i].password = strings.TrimSpace(parts[1])

		uid, err := strconv.Atoi(parts[2])
		if err != nil {
			return userList, fmt.Errorf("userInfoList: Passwd line had badly formatted uid %s", parts[2])
		}
		userList[i].uid = uid

		gid, err := strconv.Atoi(parts[3])
		if err != nil {
			return userList, fmt.Errorf("userInfoList: Passwd line had badly formatted gid %s", parts[2])
		}
		userList[i].gid = gid

		userList[i].info = strings.TrimSpace(parts[4])
		userList[i].homedir = strings.TrimSpace(parts[5])
		userList[i].shell = strings.TrimSpace(parts[6])
	}
	return userList, nil
}

func isUserPresent(user string) (bool, error) {
	found := false
	userList, err := userInfoList("/etc/passwd")
	if err != nil {
		return false, fmt.Errorf("IsUserPresent: %s", err)
	}
	for _, unixUser := range userList {
		if unixUser.username == user {
			found = true
		}
	}
	return found, nil
}

func isUserIDPresent(uid int) (bool, error) {
	found := false
	userList, err := userInfoList("/etc/passwd")
	if err != nil {
		return false, fmt.Errorf("IsUserIDPresent: %s", err)
	}

	for _, unixUser := range userList {
		if unixUser.uid == uid {
			found = true
		}
	}
	return found, nil
}

func isGroupNamePresent(group string) (bool, error) {
	found := false
	groupList, err := groupInfoLookUp("/etc/group")
	if err != nil {
		return false, fmt.Errorf("IsGroupNamePresent: %s", err)
	}
	for _, unixGroup := range groupList {
		if unixGroup.groupName == group {
			found = true
		}
	}
	return found, nil
}

func isGroupIDPresent(gid int) (bool, error) {
	found := false
	groupList, err := groupInfoLookUp("/etc/group")
	if err != nil {
		return false, fmt.Errorf("IsGroupIDPresent: %s", err)
	}
	for _, unixGroup := range groupList {
		if unixGroup.gid == gid {
			found = true
		}
	}
	return found, nil
}

func isHomeDirPresent(dirPath string) (bool, string, error) {
	found := false
	userList, err := userInfoList("/etc/passwd")
	if err != nil {
		return false, "", fmt.Errorf("IsHomeDirPresent: %s", err)
	}
	userName := ""
	for _, unixUser := range userList {
		if unixUser.homedir == dirPath {
			found = true
			userName = unixUser.username
		}
	}
	return found, userName, nil
}
