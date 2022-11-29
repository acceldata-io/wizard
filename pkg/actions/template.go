package actions

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	config_gen "github.com/acceldata-io/wizard/internal/configen"
	"github.com/acceldata-io/wizard/internal/parser"
	"github.com/acceldata-io/wizard/pkg/register"
	"github.com/acceldata-io/wizard/pkg/wlog"

	"github.com/go-playground/validator/v10"
)

type template struct {
	agentName   string
	config      interface{}
	wizardFacts map[string]interface{}
	timeout     int
	register    string
}

type templateVars struct {
	SourceType  string `json:"src_type" validate:"required"`
	Source      string `json:"src" validate:"required"`
	Destination string `json:"dest" validate:"required"`
	Permission  string `json:"permission" validate:"required"`
	Owner       string `json:"owner" validate:"required"`
	Group       string `json:"group" validate:"required"`
	Force       bool   `json:"force"`
	Backup      bool   `json:"backup"`
	Parents     bool   `json:"parents"`
}

func NewTemplateAction(agentName string, config interface{}, wizardFacts map[string]interface{}, timeout int, localRegister string) Action {
	return &template{
		agentName:   agentName,
		config:      config,
		timeout:     timeout,
		wizardFacts: wizardFacts,
		register:    localRegister,
	}
}

func newTemplateVars(data map[string]interface{}) (*templateVars, error) {
	templateConfig := templateVars{}

	if dataB, err := json.Marshal(data); err == nil {
		if err := json.Unmarshal(dataB, &templateConfig); err != nil {
			return &templateConfig, err
		}
	} else {
		return &templateConfig, err
	}

	validate := validator.New()
	err := validate.Struct(templateConfig)
	if err != nil {
		return &templateConfig, err
	}
	return &templateConfig, nil
}

func (t *template) Do(actions *parser.Action, wizardLog chan interface{}) error {
	/*
		1. Get action vars
		2. generate template file with config in a tmp location
		3. compare hash -
			1. if same ignore
			2. if different copy tmpl file
			3. Should check if dest is dir or not? // doubt
	*/
	tRegister := register.RMap[t.register]

	wizardLog <- wlog.WLInfo("executing when condition")
	if actions.When != nil {
		when := NewWhen(actions.When.Command, actions.When.RVar, actions.When.ExitCode, t.timeout)
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

	tmplVars, err := newTemplateVars(actions.ActionVariables)
	if err != nil {
		wizardLog <- wlog.WLError("validation error: " + err.Error())
		return err
	}

	wizardLog <- wlog.WLInfo("creating template from src: " + tmplVars.Source + " to: /tmp")
	err = config_gen.Execute(t.config, t.wizardFacts, tmplVars.Source, "/tmp", tmplVars.SourceType, PackageFiles)
	if err != nil {
		return err
	}

	// is dest already exists compare hash
	tmpSrc := config_gen.GetDestPath(tmplVars.Source, "/tmp")

	if tmplVars.Parents {
		wizardLog <- wlog.WLInfo("creating parents")
		dir, _ := filepath.Split(tmplVars.Destination)
		err := CreateIfNotExists(dir, "0655")
		if err != nil {
			return err
		}
	}

	if _, err := os.Stat(tmplVars.Destination); err == nil {
		if tmplVars.Force || GetHashOfFile(tmplVars.Destination) != GetHashOfFile(tmpSrc) {
			// take back up and then copy?
			wizardLog <- wlog.WLInfo("destination file found, hash not matched ot force is true, copying template from src: " + tmpSrc + " to:" + tmplVars.Destination)
			err = CopyFile(tmpSrc, tmplVars.Destination, tmplVars.Permission, "local")
			if err != nil {
				return err
			}
			tRegister.Changed = true
		} else {
			wizardLog <- wlog.WLInfo("hash matched and force is false")
		}
	} else {
		wizardLog <- wlog.WLInfo("destination file not found, copying template from src: " + tmpSrc + " to:" + tmplVars.Destination)
		err = CopyFile(tmpSrc, tmplVars.Destination, tmplVars.Permission, "local")
		if err != nil {
			return err
		}
		tRegister.Changed = true
	}

	// change owner of file
	var uid, gid int

	wizardLog <- wlog.WLInfo("changing owner for the file with: " + tmplVars.Owner)
	fileUser, err := user.Lookup(tmplVars.Owner)
	if err != nil {
		return err
	}

	uid, err = strconv.Atoi(fileUser.Uid)
	if err != nil {
		wizardLog <- wlog.WLInfo("unable to convert the uid to int. Because: " + err.Error())
		return err
	}
	if strings.TrimSpace(tmplVars.Group) == "" {
		gid, err = strconv.Atoi(fileUser.Gid)
		if err != nil {
			wizardLog <- wlog.WLInfo("unable to convert the gid to int. Because: " + err.Error())
			return err
		}
	} else {
		fileGroup, err := user.LookupGroup(tmplVars.Group)
		if err != nil {
			return err
		}
		gid, err = strconv.Atoi(fileGroup.Gid)
		if err != nil {
			wizardLog <- wlog.WLInfo("unable to convert the gid to int. Because: " + err.Error())
			return err
		}
	}

	err = os.Chown(tmplVars.Destination, uid, gid)
	if err != nil {
		wizardLog <- wlog.WLError(fmt.Sprintf("unable to change owner to file - %s, with uid - %d, gid - %d, - %s", tmplVars.Destination, uid, gid, err))
		return fmt.Errorf("unable to change owner to file - %s, with uid - %d, gid - %d, - %s", tmplVars.Destination, uid, gid, err)
	}

	// Remove tmp file
	err = os.Remove(tmpSrc)
	if err != nil {
		return err
	}

	return nil
}
