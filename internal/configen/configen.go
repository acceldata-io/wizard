package config_gen

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/acceldata-io/wizard/internal/parser"
)

// Execute function is used to generate files using tmpl file and varsData to a certain destination
func Execute(templateData interface{}, facts map[string]interface{}, TmplPath, DestPath, scrType string, Files embed.FS) error {
	conf, err := GetFileAsString(TmplPath, scrType, Files)
	if err != nil {
		return err
	}

	funcMap := parser.MergeFuncMap(sprig.GenericFuncMap(), facts)

	t, err := template.New("AgentConfig").Funcs(funcMap).Parse(conf)
	if err != nil {
		return err
	}

	if _, err := os.Stat(DestPath); os.IsNotExist(err) {
		err := os.MkdirAll(DestPath, 0o644)
		if err != nil {
			return err
		}
	}

	var f *os.File
	f, err = os.Create(GetDestPath(TmplPath, DestPath))
	if err != nil {
		return err
	}

	err = t.Execute(f, templateData)
	if err != nil {
		return err
	}

	return nil
}

func GetDestPath(TmplPath, DestPath string) string {
	_, fileName := filepath.Split(TmplPath)
	destPath := DestPath + "/" + strings.TrimSuffix(fileName, ".tmpl")
	return destPath
}

func GetFileAsString(filePath string, srcType string, Files embed.FS) (string, error) {
	file := ""
	var fileData []byte

	if srcType == "local" {
		_, err := os.Stat(filePath)
		if err != nil {
			return file, err
		}
		fileData, err = os.ReadFile(filePath)
		if err != nil {
			return file, err
		}
	} else if srcType == "embed" {
		_, err := fs.Stat(Files, filePath)
		if err != nil {
			return file, err
		}
		fileData, err = fs.ReadFile(Files, filePath)
		if err != nil {
			return file, err
		}
	}

	return string(fileData), nil
}
