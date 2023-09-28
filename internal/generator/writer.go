package generator

import (
	"context"
	"errors"
	"fmt"
	"github.com/mach-composer/mach-composer-cli/internal/state"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/rs/zerolog/log"

	"github.com/mach-composer/mach-composer-cli/internal/config"
	"github.com/mach-composer/mach-composer-cli/internal/lockfile"
)

type GenerateOptions struct {
	OutputPath string
	Site       string
}

func WriteFiles(ctx context.Context, cfg *config.MachConfig, options *GenerateOptions) (map[string]string, error) {
	locations := FileLocations(cfg, options)

	for _, site := range cfg.Sites {
		renderer, err := state.NewRenderer(
			state.Type(cfg.Global.TerraformStateProvider),
			site.Identifier,
			cfg.Global.TerraformConfig.RemoteState,
		)
		if err != nil {
			return nil, err
		}
		err = cfg.StateRepository.Add(renderer.Key(), renderer)
		if err != nil {
			return nil, err
		}

		if options.Site != "" && site.Identifier != options.Site {
			continue
		}

		path := locations[site.Identifier]
		lock, err := lockfile.GetLock(cfg, path)
		if err != nil {
			return nil, err
		}

		if !lock.HasChanges(cfg) {
			log.Info().Msgf("Files for site %s are up-to-date", site.Identifier)
			continue
		}

		filename := filepath.Join(path, "site.tf")

		log.Info().Msgf("Writing %s", filename)
		body, err := renderSite(ctx, cfg, &site)
		if err != nil {
			return nil, err
		}

		// Format and validate the file
		formatted := formatFile([]byte(body))
		if err := validateFile(formatted); err != nil {
			log.Error().Msg("The generated terraform code is invalid. " +
				"This is a bug in mach composer. Please report the issue at " +
				"https://github.com/mach-composer/mach-composer-cli")
		}

		if err := os.MkdirAll(path, 0700); err != nil {
			return nil, fmt.Errorf("error creating directory structure: %w", err)
		}

		if err := os.WriteFile(filename, formatted, 0700); err != nil {
			return nil, fmt.Errorf("error writing file: %w", err)
		}

		for _, fs := range cfg.Variables.GetEncryptedSources(site.Identifier) {
			target := filepath.Join(path, fs.Filename)
			log.Info().Msgf("Copying %s", target)
			if err := copyFile(fs.Filename, target); err != nil {
				return nil, fmt.Errorf("error writing extra file: %w", err)
			}
		}

		if err := lock.Update(cfg); err != nil {
			return nil, err
		}
		if err := lockfile.WriteLock(lock); err != nil {
			return nil, err
		}
	}
	return locations, nil
}

func FileLocations(cfg *config.MachConfig, options *GenerateOptions) map[string]string {
	path := strings.TrimSuffix(filepath.Base(cfg.Filename), filepath.Ext(cfg.Filename))
	sitesPath := filepath.Join(options.OutputPath, path)

	locations := map[string]string{}

	for i := range cfg.Sites {
		site := cfg.Sites[i]
		if options.Site != "" && site.Identifier != options.Site {
			continue
		}
		locations[site.Identifier] = filepath.Join(sitesPath, site.Identifier)
	}
	return locations
}

func formatFile(src []byte) []byte {
	// Trim whitespaces prefix
	regex := regexp.MustCompile(`(?m)^\s*`)
	src = regex.ReplaceAll(src, []byte(""))

	// Trim whitespace suffix
	regex = regexp.MustCompile(`(?m)\s*$`)
	src = regex.ReplaceAll(src, []byte(""))

	// Close empty curly blocks on same line
	regex = regexp.MustCompile(`(?m){$\s+}$`)
	src = regex.ReplaceAll(src, []byte("{}"))

	// Close empty array blocks on same line
	regex = regexp.MustCompile(`(?m)\[$\s+\]$`)
	src = regex.ReplaceAll(src, []byte("[]"))

	// Return re-formatted version
	src = hclwrite.Format(src)

	// Insert newline after closing curly brace
	regex = regexp.MustCompile("(?m)^}$")
	src = regex.ReplaceAll(src, []byte("}\n"))

	return src
}

func validateFile(src []byte) error {
	parser := hclparse.NewParser()

	_, diags := parser.ParseHCL(src, "site.tf")
	if diags.HasErrors() {
		log.Debug().Msg("Generate HCL has errors:")
		for _, err := range diags.Errs() {
			log.Debug().Err(err).Msg("error")
		}
		return errors.New("generated HCL is invalid")
	}
	return nil
}

func copyFile(srcPath, dstPath string) error {
	// Open the source file
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	// Read the contents of the source file
	srcContents, err := ioutil.ReadAll(src)
	if err != nil {
		return err
	}

	// Create the destination file
	dst, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	// WriteLock the contents of the source file to the destination file
	_, err = dst.Write(srcContents)
	if err != nil {
		return err
	}

	return nil
}
