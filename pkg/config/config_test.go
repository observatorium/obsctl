package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/efficientgo/tools/core/pkg/testutil"
)

func TestSave(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-save")
	testutil.Ok(t, err)
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))

	testutil.Ok(t, ioutil.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))

	t.Run("empty config check", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
		}

		testutil.Ok(t, cfg.Save())

		b, err := os.ReadFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"))
		testutil.Ok(t, err)

		var cfgExp Config
		testutil.Ok(t, json.Unmarshal(b, &cfgExp))

		testutil.Equals(t, cfg.APIs, cfgExp.APIs)
		testutil.Equals(t, cfg.Current, cfgExp.Current)
	})

	t.Run("config with one API no tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: nil},
			},
		}

		testutil.Ok(t, cfg.Save())

		b, err := os.ReadFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"))
		testutil.Ok(t, err)

		var cfgExp Config
		testutil.Ok(t, json.Unmarshal(b, &cfgExp))

		testutil.Equals(t, cfg.APIs, cfgExp.APIs)
		testutil.Equals(t, cfg.Current, cfgExp.Current)
	})

	t.Run("config with one API and tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.Save())

		b, err := os.ReadFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"))
		testutil.Ok(t, err)

		var cfgExp Config
		testutil.Ok(t, json.Unmarshal(b, &cfgExp))

		testutil.Equals(t, cfg.APIs, cfgExp.APIs)
		testutil.Equals(t, cfg.Current, cfgExp.Current)
	})

	t.Run("config with multiple API and tenants", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.Save())

		b, err := os.ReadFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"))
		testutil.Ok(t, err)

		var cfgExp Config
		testutil.Ok(t, json.Unmarshal(b, &cfgExp))

		testutil.Equals(t, cfg.APIs, cfgExp.APIs)
		testutil.Equals(t, cfg.Current, cfgExp.Current)
	})
}

func TestRead(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-save")
	testutil.Ok(t, err)
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))

	testutil.Ok(t, ioutil.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))

	t.Run("empty config check", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
		}

		testutil.Ok(t, cfg.Save())

		cfgExp, err := Read(filepath.Join(tmpDir, "obsctl", "test", "config.json"))
		testutil.Ok(t, err)

		testutil.Equals(t, cfg, *cfgExp)
	})

	t.Run("config with one API no tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: nil},
			},
		}

		testutil.Ok(t, cfg.Save())

		cfgExp, err := Read(filepath.Join(tmpDir, "obsctl", "test", "config.json"))
		testutil.Ok(t, err)

		testutil.Equals(t, cfg, *cfgExp)
	})

	t.Run("config with one API and tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.Save())

		cfgExp, err := Read(filepath.Join(tmpDir, "obsctl", "test", "config.json"))
		testutil.Ok(t, err)

		testutil.Equals(t, cfg, *cfgExp)
	})

	t.Run("config with multiple API and tenants", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.Save())

		cfgExp, err := Read(filepath.Join(tmpDir, "obsctl", "test", "config.json"))
		testutil.Ok(t, err)

		testutil.Equals(t, cfg, *cfgExp)
	})
}

func TestAddAPI(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-save")
	testutil.Ok(t, err)
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))

	testutil.Ok(t, ioutil.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))

	t.Run("first or empty config", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
		}

		testutil.Ok(t, cfg.AddAPI("stage", "http://stage.obs.api"))

		exp := map[APIName]APIConfig{"stage": {URL: "http://stage.obs.api", Contexts: nil}}

		testutil.Equals(t, cfg.APIs, exp)
	})

	t.Run("config with one API no tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: nil},
			},
		}

		testutil.Ok(t, cfg.AddAPI("prod", "https://prod.api:8080"))

		exp := map[APIName]APIConfig{
			"stage": {URL: "https://stage.api:9090", Contexts: nil},
			"prod":  {URL: "https://prod.api:8080", Contexts: nil},
		}

		testutil.Equals(t, cfg.APIs, exp)
	})

	t.Run("config with one API and tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.AddAPI("prod", "https://prod.api:8080"))

		exp := map[APIName]APIConfig{
			"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
				"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
			"prod": {URL: "https://prod.api:8080", Contexts: nil},
		}

		testutil.Equals(t, cfg.APIs, exp)
	})

	t.Run("config with multiple API and tenants", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.AddAPI("test", "https://test.api:8080"))

		exp := map[APIName]APIConfig{
			"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
				"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
			"prod": {URL: "https://prod.api:9090", Contexts: map[TenantName]TenantConfig{
				"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
			"test": {URL: "https://test.api:8080", Contexts: nil},
		}

		testutil.Equals(t, cfg.APIs, exp)
	})

	t.Run("api with no name", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.AddAPI("", "https://prod.api:8080"))

		exp := map[APIName]APIConfig{
			"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
				"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
			"prod.api:8080": {URL: "https://prod.api:8080", Contexts: nil},
		}

		testutil.Equals(t, cfg.APIs, exp)
	})

	t.Run("api with no name and invalid url", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		err := cfg.AddAPI("", "abcdefghijk")
		testutil.NotOk(t, err)

		testutil.Equals(t, fmt.Errorf("abcdefghijk is not a valid URL"), err)
	})
}

func TestRemoveAPI(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-save")
	testutil.Ok(t, err)
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))

	testutil.Ok(t, ioutil.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))

	t.Run("empty config", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
		}

		err := cfg.RemoveAPI("stage")
		testutil.NotOk(t, err)
		testutil.Equals(t, fmt.Errorf("api with name stage doesn't exist"), err)
	})

	t.Run("config with one API no tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: nil},
			},
		}

		testutil.Ok(t, cfg.RemoveAPI("stage"))
		testutil.Equals(t, cfg.APIs, map[APIName]APIConfig{})
	})

	t.Run("config with one API and tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.RemoveAPI("stage"))
		testutil.Equals(t, cfg.APIs, map[APIName]APIConfig{})
	})

	t.Run("config with multiple API and tenants", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.RemoveAPI("stage"))

		exp := map[APIName]APIConfig{
			"prod": {URL: "https://prod.api:9090", Contexts: map[TenantName]TenantConfig{
				"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
		}

		testutil.Equals(t, cfg.APIs, exp)
	})

	t.Run("config with multiple API and current", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
			Current: struct {
				API    APIName    `json:"api"`
				Tenant TenantName `json:"tenant"`
			}{
				API:    "stage",
				Tenant: "first",
			},
		}

		testutil.Ok(t, cfg.RemoveAPI("stage"))

		exp := map[APIName]APIConfig{
			"prod": {URL: "https://prod.api:9090", Contexts: map[TenantName]TenantConfig{
				"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
		}

		testutil.Equals(t, cfg.APIs, exp)
		testutil.Equals(t, cfg.Current, struct {
			API    APIName    `json:"api"`
			Tenant TenantName `json:"tenant"`
		}{
			API:    "",
			Tenant: "",
		})
	})
}

func TestAddTenant(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-save")
	testutil.Ok(t, err)
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))

	testutil.Ok(t, ioutil.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))

	testoidc := &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}

	t.Run("config with one API no tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: nil},
			},
		}

		testutil.Ok(t, cfg.AddTenant("first", "stage", "first", testoidc))

		exp := map[APIName]APIConfig{
			"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
				"first": {Tenant: "first", OIDC: testoidc},
			}},
		}

		testutil.Equals(t, cfg.APIs, exp)
		testutil.Equals(t, cfg.Current, struct {
			API    APIName    `json:"api"`
			Tenant TenantName `json:"tenant"`
		}{
			API:    "stage",
			Tenant: "first",
		})

	})

	t.Run("config with one API and tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first": {Tenant: "first", OIDC: testoidc},
				}},
			},
		}

		testutil.Ok(t, cfg.AddTenant("second", "stage", "second", testoidc))

		exp := map[APIName]APIConfig{
			"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
				"first":  {Tenant: "first", OIDC: testoidc},
				"second": {Tenant: "second", OIDC: testoidc},
			}},
		}

		testutil.Equals(t, cfg.APIs, exp)
		testutil.Equals(t, cfg.Current, struct {
			API    APIName    `json:"api"`
			Tenant TenantName `json:"tenant"`
		}{
			API:    "stage",
			Tenant: "second",
		})
	})

	t.Run("tenant already exists", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		err := cfg.AddTenant("second", "stage", "second", testoidc)
		testutil.NotOk(t, err)

		testutil.Equals(t, fmt.Errorf("tenant with name second already exists in api stage"), err)
	})

	t.Run("no such api", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		err := cfg.AddTenant("second", "prod", "second", testoidc)
		testutil.NotOk(t, err)

		testutil.Equals(t, fmt.Errorf("api with name prod doesn't exist"), err)
	})
}

func TestRemoveTenant(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-save")
	testutil.Ok(t, err)
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))

	testutil.Ok(t, ioutil.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))

	t.Run("empty config", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
		}

		err := cfg.RemoveTenant("first", "stage")
		testutil.NotOk(t, err)
		testutil.Equals(t, fmt.Errorf("api with name stage doesn't exist"), err)
	})

	t.Run("config with one API no tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: nil},
			},
		}

		err := cfg.RemoveTenant("first", "stage")

		testutil.NotOk(t, err)
		testutil.Equals(t, fmt.Errorf("tenant with name first doesn't exist in api stage"), err)
	})

	t.Run("config with one API and tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.RemoveTenant("first", "stage"))

		testutil.Equals(t, cfg.APIs, map[APIName]APIConfig{"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{}}})
	})

	t.Run("config with multiple API and tenants", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.RemoveTenant("second", "stage"))
		testutil.Ok(t, cfg.RemoveTenant("first", "prod"))

		exp := map[APIName]APIConfig{
			"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
				"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
			"prod": {URL: "https://prod.api:9090", Contexts: map[TenantName]TenantConfig{
				"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
		}

		testutil.Equals(t, cfg.APIs, exp)
	})
}

func TestGetCurrent(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-save")
	testutil.Ok(t, err)
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))

	testutil.Ok(t, ioutil.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))

	t.Run("empty config", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
		}

		_, _, err := cfg.GetCurrent()
		testutil.NotOk(t, err)
		testutil.Equals(t, fmt.Errorf("current context is empty"), err)
	})

	t.Run("config with multiple API and current", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
			Current: struct {
				API    APIName    `json:"api"`
				Tenant TenantName `json:"tenant"`
			}{
				API:    "stage",
				Tenant: "second",
			},
		}

		tenantConfig, apiConfig, err := cfg.GetCurrent()
		testutil.Ok(t, err)

		tenantExp := TenantConfig{Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}}

		apiExp := APIConfig{URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
			"first":  {Tenant: "first", CAFile: nil, OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			"second": {Tenant: "second", CAFile: nil, OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
		}}

		testutil.Equals(t, tenantConfig, tenantExp)
		testutil.Equals(t, apiConfig, apiExp)
	})
}

func TestSetCurrent(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "test-save")
	testutil.Ok(t, err)
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))

	testutil.Ok(t, ioutil.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))

	t.Run("empty config", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
		}

		err := cfg.SetCurrent("stage", "first")
		testutil.NotOk(t, err)
		testutil.Equals(t, fmt.Errorf("api with name stage doesn't exist"), err)
	})

	t.Run("config with one API no tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: nil},
			},
		}

		err := cfg.SetCurrent("stage", "first")

		testutil.NotOk(t, err)
		testutil.Equals(t, fmt.Errorf("tenant with name first doesn't exist in api stage"), err)
	})

	t.Run("config with multiple API and no current", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.SetCurrent("prod", "first"))

		testutil.Equals(t, cfg.Current, struct {
			API    APIName    `json:"api"`
			Tenant TenantName `json:"tenant"`
		}{
			API:    "prod",
			Tenant: "first",
		})
	})

	t.Run("config with multiple API and current", func(t *testing.T) {
		cfg := Config{
			pathOverride: []string{filepath.Join(tmpDir, "obsctl", "test", "config.json")},
			APIs: map[APIName]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[TenantName]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
			Current: struct {
				API    APIName    `json:"api"`
				Tenant TenantName `json:"tenant"`
			}{
				API:    "stage",
				Tenant: "second",
			},
		}

		testutil.Ok(t, cfg.SetCurrent("prod", "first"))

		testutil.Equals(t, cfg.Current, struct {
			API    APIName    `json:"api"`
			Tenant TenantName `json:"tenant"`
		}{
			API:    "prod",
			Tenant: "first",
		})
	})
}
