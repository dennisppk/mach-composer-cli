package utils

import (
	"context"
	"fmt"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
	"github.com/zclconf/go-cty/cty/json"
	"os"
	"os/exec"
)

type SiteComponentOutput struct {
	Sensitive bool `cty:"sensitive"`
	Value     struct {
		Hash      string    `cty:"hash"`
		Variables cty.Value `cty:"variables"`
	} `cty:"value"`
	Type cty.Value `cty:"type"`
}

// RunTerraform will execute a terraform command with the given arguments in the given directory.
func RunTerraform(ctx context.Context, cwd string, args ...string) (string, error) {
	if _, err := os.Stat(cwd); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("the generated files are not found: %w", err)
		}
	}

	execPath, err := exec.LookPath("terraform")
	if err != nil {
		return "", err
	}

	return RunInteractive(ctx, execPath, cwd, args...)
}

// GetTerraformOutputByKey returns the output of a terraform command for the given key at the given path.
// If no output is found nil is returned.
func GetTerraformOutputByKey(ctx context.Context, path string, key string) (*SiteComponentOutput, error) {
	var data json.SimpleJSONValue

	output, err := RunTerraform(ctx, path, "output", "-json")
	if err != nil {
		return nil, err
	}

	if err = data.UnmarshalJSON([]byte(output)); err != nil {
		return nil, err
	}

	if !data.Type().HasAttribute(key) {
		return nil, nil
	}

	val := data.GetAttr(key)

	var scOut SiteComponentOutput
	err = gocty.FromCtyValue(val, &scOut)
	if err != nil {
		return nil, err
	}

	return &scOut, nil
}
