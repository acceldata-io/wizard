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
