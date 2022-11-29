package parser

import (
	"os"
	"testing"
)

func TestParseConfigFail(t *testing.T) {
	file, _ := os.ReadFile("../../testdata/parser_config_fail.yaml")
	_, err := ParseConfig(file)
	if err == nil {
		t.Fail()
	}
}

func TestParseConfigPass(t *testing.T) {
	file, _ := os.ReadFile("../../testdata/parser_config_pass.json")
	_, err := ParseConfig(file)
	if err != nil {
		t.Fail()
	}
}
