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

	"github.com/acceldata-io/wizard/internal/parser"
)

//go:generate mockgen -source ./actions.go -destination ./mocks/mock_actions.go

type Action interface {
	Do(actions *parser.Action, wizardLog chan interface{}) error
}

var PackageFiles embed.FS
