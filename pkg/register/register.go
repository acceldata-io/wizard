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

package register

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
)

type Register struct {
	Changed  bool
	StdOut   string
	StdErr   string
	ExitCode int
}

var RMap = make(map[string]*Register)

func Reset() {
	for k := range RMap {
		delete(RMap, k)
	}
}

func ParseRegisterExp(exp string) (bool, error) {
	tokens := strings.Split(exp, " ")
	rStack := newRegisterStack()

	for _, token := range tokens {
		subTokens := strings.Split(token, ".")
		if len(subTokens) != 2 && len(subTokens) == 1 {
			// Check for logical operators - and, or, not, eq
			if !pushRegisterOperator(rStack, subTokens[0]) {
				pushAny(rStack, subTokens[0])
			}
		} else if len(subTokens) == 2 {
			if !pushRegisterVal(rStack, subTokens[0], subTokens[1]) {
				return false, fmt.Errorf("invalid register field %q", subTokens[1])
			}
		} else {
			return false, fmt.Errorf("invalid register field %q", token)
		}
	}

	nStack := newRegisterStack()
	for {
		if rStack.len() == 1 {
			nStack.push(rStack.pop())
		} else {
			var rVal, rFieldOp, rFieldVal string
			if rStack.len() != 0 {
				_, rFieldVal = rStack.pop()
			}
			if rStack.len() != 0 {
				_, rFieldOp = rStack.pop()
			}
			if rStack.len() != 0 {
				_, rVal = rStack.pop()
			}

			res := doOperations(rFieldOp, rFieldVal, rVal)
			nStack.push("val", strconv.FormatBool(res))
		}

		if rStack.len() != 0 {
			nStack.push(rStack.pop())
		}

		if nStack.len() == 0 || nStack.len() == 1 {
			break
		}

		if rStack.len() == 0 {
			rStack = nStack
			nStack = newRegisterStack()
		}
	}

	_, val := nStack.pop()
	uBool, err := strconv.ParseBool(val)
	if err != nil {
		return uBool, err
	}

	return uBool, nil
}

func doOperations(op, uVal, rVal string) bool {
	switch op {
	case "eq":
		return uVal == rVal
	case "neq":
		return uVal != rVal
	case "and":
		uBool, _ := strconv.ParseBool(uVal)
		rBool, _ := strconv.ParseBool(rVal)
		return uBool == true && rBool == true
	case "or":
		uBool, _ := strconv.ParseBool(uVal)
		rBool, _ := strconv.ParseBool(rVal)
		return uBool == true || rBool == true
	default:
		return false
	}
}

func pushRegisterOperator(rStack *stack, rField string) bool {
	switch rField {
	case "eq":
		rStack.push("op", rField)
		return true
	case "and":
		rStack.push("op", rField)
		return true
	case "or":
		rStack.push("op", rField)
		return true
	case "neq":
		rStack.push("op", rField)
		return true
	}
	return false
}

func pushRegisterVal(rStack *stack, rKey string, rField string) bool {
	switch rField {
	case "changed":
		rStack.push("bool", strconv.FormatBool(RMap[rKey].Changed))
		return true
	case "stdout":
		rStack.push("string", RMap[rKey].StdOut)
		return true
	case "stderr":
		rStack.push("string", RMap[rKey].StdErr)
		return true
	case "exit_code":
		rStack.push("int", strconv.FormatInt(int64(RMap[rKey].ExitCode), 10))
		return true
	}
	return false
}

func pushAny(rStack *stack, rField string) {
	rStack.push("val", rField)
}

func GetHash(s string) string {
	h := sha1.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
