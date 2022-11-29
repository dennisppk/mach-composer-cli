package commercetools

import (
	"fmt"

	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"

	"github.com/labd/mach-composer/internal/plugins/shared"
)

type CommercetoolsPlugin struct {
	environment string
	provider    string
	siteConfigs map[string]*SiteConfig
}

func NewCommercetoolsPlugin() *CommercetoolsPlugin {
	return &CommercetoolsPlugin{
		provider:    "0.30.0",
		siteConfigs: map[string]*SiteConfig{},
	}
}

func (p *CommercetoolsPlugin) Configure(environment string, provider string) error {
	p.environment = environment
	if provider != "" {
		p.provider = provider
	}
	return nil
}

func (p *CommercetoolsPlugin) IsEnabled() bool {
	return len(p.siteConfigs) > 0
}

func (p *CommercetoolsPlugin) Identifier() string {
	return "commercetools"
}

func (p *CommercetoolsPlugin) SetRemoteStateBackend(data map[string]any) error {
	return fmt.Errorf("not supported by this plugin")
}

func (p *CommercetoolsPlugin) SetGlobalConfig(data map[string]any) error {
	return nil
}

func (p *CommercetoolsPlugin) SetSiteConfig(site string, data map[string]any) error {
	cfg := SiteConfig{
		Components: map[string]ComponentConfig{},
	}
	if err := mapstructure.Decode(data, &cfg); err != nil {
		return err
	}

	if cfg.Frontend != nil {
		logrus.Warnf("%s: commercetools frontend block is deprecated and will be removed soon\n", site)
	}

	if err := defaults.Set(&cfg); err != nil {
		return err
	}
	p.siteConfigs[site] = &cfg

	return nil
}

func (p *CommercetoolsPlugin) SetSiteComponentConfig(site string, component string, data map[string]any) error {
	siteConfig := p.getSiteConfig(site)
	if siteConfig == nil {
		return fmt.Errorf("no site config found")
	}

	cfg := ComponentConfig{}
	if err := mapstructure.Decode(data, &cfg); err != nil {
		return err
	}
	siteConfig.Components[component] = cfg

	return nil
}

func (p *CommercetoolsPlugin) SetSiteEndpointsConfig(site string, data map[string]any) error {
	return nil
}

func (p *CommercetoolsPlugin) SetComponentConfig(component string, data map[string]any) error {
	return nil
}

func (p *CommercetoolsPlugin) SetComponentEndpointsConfig(component string, endpoints map[string]string) error {
	return nil
}

func (p *CommercetoolsPlugin) getSiteConfig(site string) *SiteConfig {
	cfg, ok := p.siteConfigs[site]
	if !ok {
		return nil
	}
	return cfg
}

func (p *CommercetoolsPlugin) TerraformRenderStateBackend(site string) (string, error) {
	return "", nil
}

func (p *CommercetoolsPlugin) TerraformRenderProviders(site string) (string, error) {
	cfg := p.getSiteConfig(site)
	if cfg == nil {
		return "", nil
	}

	result := fmt.Sprintf(`
		commercetools = {
		source = "labd/commercetools"
		version = "%s"
		}
	`, shared.VersionConstraint(p.provider))
	return result, nil
}

func (p *CommercetoolsPlugin) TerraformRenderResources(site string) (string, error) {
	cfg := p.getSiteConfig(site)
	if cfg == nil {
		return "", nil
	}

	content, err := renderResources(cfg)
	if err != nil {
		return "", err
	}

	return content, nil
}

func (p *CommercetoolsPlugin) TerraformRenderComponentResources(site string, component string) (string, error) {
	return "", nil
}

func (p *CommercetoolsPlugin) TerraformRenderComponentVars(site string, component string) (string, error) {
	cfg := p.getSiteConfig(site)
	if cfg == nil {
		return "", nil
	}

	templateContext := struct {
		Site      *SiteConfig
		Component *ComponentConfig
	}{
		Site:      cfg,
		Component: cfg.getComponentSiteConfig(component),
	}

	template := `
		ct_project_key    = {{ .Site.ProjectKey|printf "%q" }}
		ct_api_url        = {{ .Site.APIURL|printf "%q" }}
		ct_auth_url       = {{ .Site.TokenURL|printf "%q" }}

		ct_stores = {
			{{ $component := .Component }}
			{{ range $store := .Site.Stores }}
				{{ $store.Key }} =  {
					key = {{ $store.Key|printf "%q" }}
					variables = {
						{{ range $key, $value := index $component.StoreVariables $store.Key }}
						{{ $key }} = {{ $value|printf "%#v" }}
						{{ end }}
					}
					secrets = {
						{{ range $key, $value := index $component.StoreSecrets $store.Key }}
						{{ $key }} = {{ $value|printf "%#v" }}
						{{ end }}
					}
				}
			{{ end }}
		}
	`
	return shared.RenderGoTemplate(template, templateContext)
}

func (p *CommercetoolsPlugin) TerraformRenderComponentDependsOn(site string, component string) ([]string, error) {
	return []string{"null_resource.commercetools"}, nil
}

func (p *CommercetoolsPlugin) TerraformRenderComponentProviders(site string, component string) ([]string, error) {
	return []string{}, nil
}
