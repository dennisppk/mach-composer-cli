package generator

import (
	"context"
	"fmt"
	"github.com/mach-composer/mach-composer-cli/internal/config"
	"github.com/mach-composer/mach-composer-cli/internal/utils"
	"slices"
	"strings"
)

type componentContext struct {
	ComponentName       string
	ComponentVersion    string
	ComponentHash       string
	ComponentVariables  string
	ComponentSecrets    string
	SiteName            string
	Environment         string
	Source              string
	PluginResources     []string
	PluginProviders     []string
	PluginDependsOn     []string
	PluginVariables     []string
	HasCloudIntegration bool
}

func renderSiteComponent(ctx context.Context, cfg *config.MachConfig, site *config.SiteConfig, siteComponent *config.SiteComponentConfig) (string, error) {
	result := []string{
		"# This file is auto-generated by MACH composer",
		fmt.Sprintf("# DeploymentSiteComponent: %s", siteComponent.Name),
	}

	// Render the terraform config
	val, err := renderSiteComponentTerraformConfig(cfg, site, siteComponent)
	if err != nil {
		return "", fmt.Errorf("renderSiteTerraformConfig: %w", err)
	}
	result = append(result, val)

	// Render all the file sources
	val, err = renderFileSources(cfg, site)
	if err != nil {
		return "", fmt.Errorf("failed to render file sources: %w", err)
	}
	result = append(result, val)

	// Render all the resources required by the site siteComponent
	val, err = renderSiteComponentResources(cfg, site, siteComponent)
	if err != nil {
		return "", fmt.Errorf("failed to render resources: %w", err)
	}
	result = append(result, val)

	// Render data links to other deployments
	val, err = renderRemoteSources(cfg, siteComponent)
	if err != nil {
		return "", fmt.Errorf("failed to render remote sources: %w", err)
	}
	result = append(result, val)

	// Render the siteComponent module
	val, err = renderComponentModule(ctx, cfg, site, siteComponent)
	if err != nil {
		return "", fmt.Errorf("failed to render component: %w", err)
	}
	result = append(result, val)

	content := strings.Join(result, "\n")

	return content, nil
}

func renderSiteComponentTerraformConfig(cfg *config.MachConfig, site *config.SiteConfig,
	siteComponent *config.SiteComponentConfig) (string, error) {
	var providers []string
	for _, plugin := range cfg.Plugins.Names(siteComponent.Definition.Integrations...) {
		content, err := plugin.RenderTerraformProviders(site.Identifier)
		if err != nil {
			return "", fmt.Errorf("plugin %s failed to render providers: %w", plugin.Name, err)
		}
		if content != "" {
			providers = append(providers, content)
		}
	}

	if !cfg.StateRepository.Has(site.Identifier) {
		return "", fmt.Errorf("state repository does not have a backend for site %s", site.Identifier)
	}
	backendConfig, err := cfg.StateRepository.Get(siteComponent.Name).Backend()
	if err != nil {
		return "", err
	}

	tpl, err := templates.ReadFile("templates/terraform.tmpl")
	if err != nil {
		return "", err
	}

	templateContext := struct {
		Providers     []string
		BackendConfig string
		IncludeSOPS   bool
	}{
		Providers:     providers,
		BackendConfig: backendConfig,
		IncludeSOPS:   cfg.Variables.HasEncrypted(site.Identifier),
	}
	return utils.RenderGoTemplate(string(tpl), templateContext)
}

func renderSiteComponentResources(cfg *config.MachConfig, site *config.SiteConfig, siteComponent *config.SiteComponentConfig) (string, error) {
	var resources []string
	for _, plugin := range cfg.Plugins.Names(siteComponent.Definition.Integrations...) {
		content, err := plugin.RenderTerraformResources(site.Identifier)
		if err != nil {
			return "", fmt.Errorf("plugin %s failed to render resources: %w", plugin.Name, err)
		}

		if content != "" {
			resources = append(resources, content)
		}
	}

	tpl, err := templates.ReadFile("templates/resources.tmpl")
	if err != nil {
		return "", err
	}

	return utils.RenderGoTemplate(string(tpl), resources)
}

// renderComponentModule uses templates/component.tf to generate a terraform snippet for each component
func renderComponentModule(_ context.Context, cfg *config.MachConfig, site *config.SiteConfig, siteComponent *config.SiteComponentConfig) (string, error) {
	hash, err := siteComponent.Hash()
	if err != nil {
		return "", err
	}

	tc := componentContext{
		ComponentName:    siteComponent.Name,
		ComponentVersion: siteComponent.Definition.Version,
		ComponentHash:    hash,
		SiteName:         site.Identifier,
		Environment:      cfg.Global.Environment,
		Source:           siteComponent.Definition.Source,
		PluginResources:  []string{},
		PluginVariables:  []string{},
		PluginDependsOn:  []string{},
		PluginProviders:  []string{},
	}

	for _, plugin := range cfg.Plugins.Names(siteComponent.Definition.Integrations...) {
		plugin, err := cfg.Plugins.Get(plugin.Name)
		if err != nil {
			return "", err
		}

		cr, err := plugin.RenderTerraformComponent(site.Identifier, siteComponent.Name)
		if err != nil {
			return "", fmt.Errorf("plugin %s failed to render siteComponent: %w", plugin.Name, err)
		}

		if cr == nil {
			continue
		}

		tc.PluginResources = append(tc.PluginResources, cr.Resources)
		tc.PluginVariables = append(tc.PluginVariables, cr.Variables)
		tc.PluginProviders = append(tc.PluginProviders, cr.Providers...)
		tc.PluginDependsOn = append(tc.PluginDependsOn, cr.DependsOn...)
	}

	tpl, err := templates.ReadFile("templates/site_component.tmpl")
	if err != nil {
		return "", err
	}

	if siteComponent.HasCloudIntegration(&cfg.Global) {
		tc.HasCloudIntegration = true
		tc.ComponentVariables = "variables = {}"
		tc.ComponentSecrets = "secrets = {}"
	}

	if len(siteComponent.Variables) > 0 {
		val, err := serializeToHCL("variables", siteComponent.Variables, siteComponent.Deployment.Type, cfg.StateRepository)
		if err != nil {
			return "", err
		}
		tc.ComponentVariables = val
	}
	if len(siteComponent.Secrets) > 0 {
		val, err := serializeToHCL("secrets", siteComponent.Secrets, siteComponent.Deployment.Type, cfg.StateRepository)
		if err != nil {
			return "", err
		}
		tc.ComponentSecrets = val
	}

	if siteComponent.Definition.IsGitSource() {
		// When using Git, we will automatically add a reference to the string
		// so that the given version is used when fetching the module itself
		// from Git as well
		tc.Source += fmt.Sprintf("?ref=%s", siteComponent.Definition.Version)
	}

	val, err := utils.RenderGoTemplate(string(tpl), tc)
	if err != nil {
		return "", fmt.Errorf("renderSiteTerraformConfig: %w", err)
	}
	return val, nil
}

func renderRemoteSources(cfg *config.MachConfig, component *config.SiteComponentConfig) (string, error) {
	parents := append(
		component.Variables.ListReferencedComponents(),
		component.Secrets.ListReferencedComponents()...,
	)

	var links []string
	for _, parent := range parents {
		key, ok := cfg.StateRepository.Key(parent)
		if !ok {
			return "", fmt.Errorf("missing remoteState for %s", parent)
		}
		links = append(links, key)
	}

	links = slices.Compact(links)

	var result []string

	for _, link := range links {
		remoteState, err := cfg.StateRepository.Get(link).RemoteState()
		if err != nil {
			return "", err
		}

		result = append(result, remoteState)
	}

	return strings.Join(result, "\n"), nil
}
