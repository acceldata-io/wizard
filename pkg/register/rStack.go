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

type rNode struct {
	rType  string
	rValue string
}

type stack struct {
	node []rNode
}

func newRegisterStack() *stack {
	return &stack{node: []rNode{}}
}

func (r *stack) push(rType, rValue string) {
	r.node = append(r.node, rNode{
		rType:  rType,
		rValue: rValue,
	})
}

func (r *stack) pop() (string, string) {
	if len(r.node) <= 0 {
		return "", ""
	}
	rNode := r.node[len(r.node)-1]
	r.node = r.node[:len(r.node)-1]
	return rNode.rType, rNode.rValue
}

func (r *stack) peek() (string, string) {
	if len(r.node) <= 0 {
		return "", ""
	}

	rNode := r.node[len(r.node)-1]
	return rNode.rType, rNode.rValue
}

func (r *stack) len() int {
	return len(r.node)
}
