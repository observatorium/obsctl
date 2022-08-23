package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/efficientgo/tools/core/pkg/testutil"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func TestSave(t *testing.T) {
	tmpDir := t.TempDir()
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))
	testutil.Ok(t, os.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))
	testutil.Ok(t, os.Setenv("OBSCTL_CONFIG_PATH", filepath.Join(tmpDir, "obsctl", "test", "config.json")))

	tlogger := level.NewFilter(log.NewJSONLogger(log.NewSyncWriter(os.Stderr)), level.AllowDebug())

	t.Run("empty config check", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
		}

		testutil.Ok(t, cfg.Save(tlogger))

		b, err := os.ReadFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"))
		testutil.Ok(t, err)

		var cfgExp Config
		testutil.Ok(t, json.Unmarshal(b, &cfgExp))

		testutil.Equals(t, cfg.APIs, cfgExp.APIs)
		testutil.Equals(t, cfg.Current, cfgExp.Current)
	})

	t.Run("config with one API no tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: nil},
			},
		}

		testutil.Ok(t, cfg.Save(tlogger))

		b, err := os.ReadFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"))
		testutil.Ok(t, err)

		var cfgExp Config
		testutil.Ok(t, json.Unmarshal(b, &cfgExp))

		testutil.Equals(t, cfg.APIs, cfgExp.APIs)
		testutil.Equals(t, cfg.Current, cfgExp.Current)
	})

	t.Run("config with one API and tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.Save(tlogger))

		b, err := os.ReadFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"))
		testutil.Ok(t, err)

		var cfgExp Config
		testutil.Ok(t, json.Unmarshal(b, &cfgExp))

		testutil.Equals(t, cfg.APIs, cfgExp.APIs)
		testutil.Equals(t, cfg.Current, cfgExp.Current)
	})

	t.Run("config with multiple API and tenants", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.Save(tlogger))

		b, err := os.ReadFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"))
		testutil.Ok(t, err)

		var cfgExp Config
		testutil.Ok(t, json.Unmarshal(b, &cfgExp))

		testutil.Equals(t, cfg.APIs, cfgExp.APIs)
		testutil.Equals(t, cfg.Current, cfgExp.Current)
	})
}

func TestRead(t *testing.T) {
	tmpDir := t.TempDir()
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))
	testutil.Ok(t, os.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))
	testutil.Ok(t, os.Setenv("OBSCTL_CONFIG_PATH", filepath.Join(tmpDir, "obsctl", "test", "config.json")))

	tlogger := level.NewFilter(log.NewJSONLogger(log.NewSyncWriter(os.Stderr)), level.AllowDebug())

	t.Run("empty config check", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
		}

		testutil.Ok(t, cfg.Save(tlogger))

		cfgExp, err := Read(tlogger)
		testutil.Ok(t, err)

		testutil.Equals(t, cfg, *cfgExp)
	})

	t.Run("config with one API no tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: nil},
			},
		}

		testutil.Ok(t, cfg.Save(tlogger))

		cfgExp, err := Read(tlogger)
		testutil.Ok(t, err)

		testutil.Equals(t, cfg, *cfgExp)
	})

	t.Run("config with one API and tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.Save(tlogger))

		cfgExp, err := Read(tlogger)
		testutil.Ok(t, err)

		testutil.Equals(t, cfg, *cfgExp)
	})

	t.Run("config with multiple API and tenants", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.Save(tlogger))

		cfgExp, err := Read(tlogger)
		testutil.Ok(t, err)

		testutil.Equals(t, cfg, *cfgExp)
	})
}

func TestAddAPI(t *testing.T) {
	tmpDir := t.TempDir()
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))
	testutil.Ok(t, os.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))
	testutil.Ok(t, os.Setenv("OBSCTL_CONFIG_PATH", filepath.Join(tmpDir, "obsctl", "test", "config.json")))

	tlogger := level.NewFilter(log.NewJSONLogger(log.NewSyncWriter(os.Stderr)), level.AllowDebug())

	t.Run("first or empty config", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
		}

		testutil.Ok(t, cfg.AddAPI(tlogger, "stage", "http://stage.obs.api/"))

		exp := map[string]APIConfig{"stage": {URL: "http://stage.obs.api/", Contexts: nil}}

		testutil.Equals(t, cfg.APIs, exp)
	})

	t.Run("config with one API no tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090/", Contexts: nil},
			},
		}

		testutil.Ok(t, cfg.AddAPI(tlogger, "prod", "https://prod.api:8080/"))

		exp := map[string]APIConfig{
			"stage": {URL: "https://stage.api:9090/", Contexts: nil},
			"prod":  {URL: "https://prod.api:8080/", Contexts: nil},
		}

		testutil.Equals(t, cfg.APIs, exp)
	})

	t.Run("config with one API and tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090/", Contexts: map[string]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.AddAPI(tlogger, "prod", "https://prod.api:8080/"))

		exp := map[string]APIConfig{
			"stage": {URL: "https://stage.api:9090/", Contexts: map[string]TenantConfig{
				"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
			"prod": {URL: "https://prod.api:8080/", Contexts: nil},
		}

		testutil.Equals(t, cfg.APIs, exp)
	})

	t.Run("config with multiple API and tenants", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090/", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090/", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.AddAPI(tlogger, "test", "https://test.api:8080"))

		exp := map[string]APIConfig{
			"stage": {URL: "https://stage.api:9090/", Contexts: map[string]TenantConfig{
				"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
			"prod": {URL: "https://prod.api:9090/", Contexts: map[string]TenantConfig{
				"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
			"test": {URL: "https://test.api:8080/", Contexts: nil},
		}

		testutil.Equals(t, cfg.APIs, exp)
	})

	t.Run("api with no name", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090/", Contexts: map[string]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.AddAPI(tlogger, "", "https://prod.api:8080/"))

		exp := map[string]APIConfig{
			"stage": {URL: "https://stage.api:9090/", Contexts: map[string]TenantConfig{
				"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
			"prod.api:8080": {URL: "https://prod.api:8080/", Contexts: nil},
		}

		testutil.Equals(t, cfg.APIs, exp)
	})

	t.Run("api with no name and invalid url", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		err := cfg.AddAPI(tlogger, "", "abcdefghijk")
		testutil.NotOk(t, err)

		testutil.Equals(t, fmt.Errorf("abcdefghijk is not a valid URL (scheme: ,host: )"), err)
	})

	t.Run("api with no trailing slash", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090/", Contexts: map[string]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.AddAPI(tlogger, "", "https://prod.api:8080"))

		exp := map[string]APIConfig{
			"stage": {URL: "https://stage.api:9090/", Contexts: map[string]TenantConfig{
				"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
			"prod.api:8080": {URL: "https://prod.api:8080/", Contexts: nil},
		}

		testutil.Equals(t, cfg.APIs, exp)
	})

	t.Run("api with slash in name", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090/", Contexts: map[string]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		err := cfg.AddAPI(tlogger, "prod/123", "https://prod.api:8080")
		testutil.NotOk(t, err)

		testutil.Equals(t, fmt.Errorf("api name prod/123 cannot contain slashes"), err)
	})
}

func TestRemoveAPI(t *testing.T) {
	tmpDir := t.TempDir()
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))
	testutil.Ok(t, os.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))
	testutil.Ok(t, os.Setenv("OBSCTL_CONFIG_PATH", filepath.Join(tmpDir, "obsctl", "test", "config.json")))

	tlogger := level.NewFilter(log.NewJSONLogger(log.NewSyncWriter(os.Stderr)), level.AllowDebug())

	t.Run("empty config", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
		}

		err := cfg.RemoveAPI(tlogger, "stage")
		testutil.NotOk(t, err)
		testutil.Equals(t, fmt.Errorf("api with name stage doesn't exist"), err)
	})

	t.Run("config with one API no tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: nil},
			},
		}

		testutil.Ok(t, cfg.RemoveAPI(tlogger, "stage"))
		testutil.Equals(t, cfg.APIs, map[string]APIConfig{})
	})

	t.Run("config with one API and no name given", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: nil},
			},
		}

		testutil.Ok(t, cfg.RemoveAPI(tlogger, ""))
		testutil.Equals(t, cfg.APIs, map[string]APIConfig{})
	})

	t.Run("config with one API and tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.RemoveAPI(tlogger, "stage"))
		testutil.Equals(t, cfg.APIs, map[string]APIConfig{})
	})

	t.Run("config with one API and tenant but no name given", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.RemoveAPI(tlogger, ""))
		testutil.Equals(t, cfg.APIs, map[string]APIConfig{})
	})

	t.Run("config with one current API and tenant but no name given", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
			Current: struct {
				API    string `json:"api"`
				Tenant string `json:"tenant"`
			}{
				API:    "stage",
				Tenant: "first",
			},
		}

		testutil.Ok(t, cfg.RemoveAPI(tlogger, ""))
		testutil.Equals(t, cfg.APIs, map[string]APIConfig{})
	})

	t.Run("config with multiple API and tenants", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.RemoveAPI(tlogger, "stage"))

		exp := map[string]APIConfig{
			"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
				"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
		}

		testutil.Equals(t, cfg.APIs, exp)
	})

	t.Run("config with multiple API and tenants and no name given", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		err := cfg.RemoveAPI(tlogger, "")
		testutil.NotOk(t, err)

		testutil.Equals(t, fmt.Errorf("api with name  doesn't exist"), err)
	})

	t.Run("config with multiple API and tenants", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.RemoveAPI(tlogger, "stage"))

		exp := map[string]APIConfig{
			"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
				"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
		}

		testutil.Equals(t, cfg.APIs, exp)
	})

	t.Run("config with multiple API and current", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
			Current: struct {
				API    string `json:"api"`
				Tenant string `json:"tenant"`
			}{
				API:    "stage",
				Tenant: "first",
			},
		}

		testutil.Ok(t, cfg.RemoveAPI(tlogger, "stage"))

		exp := map[string]APIConfig{
			"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
				"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
		}

		testutil.Equals(t, cfg.APIs, exp)
		testutil.Equals(t, cfg.Current, struct {
			API    string `json:"api"`
			Tenant string `json:"tenant"`
		}{
			API:    "",
			Tenant: "",
		})
	})
}

func TestAddTenant(t *testing.T) {
	tmpDir := t.TempDir()
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))
	testutil.Ok(t, os.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))
	testutil.Ok(t, os.Setenv("OBSCTL_CONFIG_PATH", filepath.Join(tmpDir, "obsctl", "test", "config.json")))

	tlogger := level.NewFilter(log.NewJSONLogger(log.NewSyncWriter(os.Stderr)), level.AllowDebug())

	testoidc := &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}

	t.Run("config with one API no tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: nil},
			},
		}

		testutil.Ok(t, cfg.AddTenant(tlogger, "first", "stage", "first", testoidc))

		exp := map[string]APIConfig{
			"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
				"first": {Tenant: "first", OIDC: testoidc},
			}},
		}

		testutil.Equals(t, cfg.APIs, exp)
		testutil.Equals(t, cfg.Current, struct {
			API    string `json:"api"`
			Tenant string `json:"tenant"`
		}{
			API:    "stage",
			Tenant: "first",
		})

	})

	t.Run("config with one API and tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first": {Tenant: "first", OIDC: testoidc},
				}},
			},
		}

		testutil.Ok(t, cfg.AddTenant(tlogger, "second", "stage", "second", testoidc))

		exp := map[string]APIConfig{
			"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
				"first":  {Tenant: "first", OIDC: testoidc},
				"second": {Tenant: "second", OIDC: testoidc},
			}},
		}

		testutil.Equals(t, cfg.APIs, exp)
		testutil.Equals(t, cfg.Current, struct {
			API    string `json:"api"`
			Tenant string `json:"tenant"`
		}{
			API:    "stage",
			Tenant: "second",
		})
	})

	t.Run("tenant already exists", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		err := cfg.AddTenant(tlogger, "second", "stage", "second", testoidc)
		testutil.NotOk(t, err)

		testutil.Equals(t, fmt.Errorf("tenant with name second already exists in api stage"), err)
	})

	t.Run("no such api", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		err := cfg.AddTenant(tlogger, "second", "prod", "second", testoidc)
		testutil.NotOk(t, err)

		testutil.Equals(t, fmt.Errorf("api with name prod doesn't exist"), err)
	})

	t.Run("tenant name has slash", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		err := cfg.AddTenant(tlogger, "test/123", "stage", "test/123", testoidc)
		testutil.NotOk(t, err)

		testutil.Equals(t, fmt.Errorf("tenant name test/123 cannot contain slashes"), err)
	})
}

func TestRemoveTenant(t *testing.T) {
	tmpDir := t.TempDir()
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))
	testutil.Ok(t, os.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))
	testutil.Ok(t, os.Setenv("OBSCTL_CONFIG_PATH", filepath.Join(tmpDir, "obsctl", "test", "config.json")))

	tlogger := level.NewFilter(log.NewJSONLogger(log.NewSyncWriter(os.Stderr)), level.AllowDebug())

	t.Run("empty config", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
		}

		err := cfg.RemoveTenant(tlogger, "first", "stage")
		testutil.NotOk(t, err)
		testutil.Equals(t, fmt.Errorf("api with name stage doesn't exist"), err)
	})

	t.Run("config with one API no tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: nil},
			},
		}

		err := cfg.RemoveTenant(tlogger, "first", "stage")

		testutil.NotOk(t, err)
		testutil.Equals(t, fmt.Errorf("tenant with name first doesn't exist in api stage"), err)
	})

	t.Run("config with one API and tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.RemoveTenant(tlogger, "first", "stage"))

		testutil.Equals(t, cfg.APIs, map[string]APIConfig{"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{}}})
	})

	t.Run("config with multiple API and tenants", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.RemoveTenant(tlogger, "second", "stage"))
		testutil.Ok(t, cfg.RemoveTenant(tlogger, "first", "prod"))

		exp := map[string]APIConfig{
			"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
				"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
			"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
				"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
		}

		testutil.Equals(t, cfg.APIs, exp)
	})
}

func TestGetCurrentContext(t *testing.T) {
	tmpDir := t.TempDir()
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))
	testutil.Ok(t, os.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))
	testutil.Ok(t, os.Setenv("OBSCTL_CONFIG_PATH", filepath.Join(tmpDir, "obsctl", "test", "config.json")))

	t.Run("empty config", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
		}

		_, _, err := cfg.GetCurrentContext()
		testutil.NotOk(t, err)
		testutil.Equals(t, fmt.Errorf("current context is empty"), err)
	})

	t.Run("config with multiple API and current", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
			Current: struct {
				API    string `json:"api"`
				Tenant string `json:"tenant"`
			}{
				API:    "stage",
				Tenant: "second",
			},
		}

		tenantConfig, apiConfig, err := cfg.GetCurrentContext()
		testutil.Ok(t, err)

		tenantExp := TenantConfig{Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}}

		apiExp := APIConfig{URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
			"first":  {Tenant: "first", CAFile: nil, OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			"second": {Tenant: "second", CAFile: nil, OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
		}}

		testutil.Equals(t, tenantConfig, tenantExp)
		testutil.Equals(t, apiConfig, apiExp)
	})
}

func TestSetCurrentContext(t *testing.T) {
	tmpDir := t.TempDir()
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))
	testutil.Ok(t, os.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))
	testutil.Ok(t, os.Setenv("OBSCTL_CONFIG_PATH", filepath.Join(tmpDir, "obsctl", "test", "config.json")))

	tlogger := level.NewFilter(log.NewJSONLogger(log.NewSyncWriter(os.Stderr)), level.AllowDebug())

	t.Run("empty config", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
		}

		err := cfg.SetCurrentContext(tlogger, "stage", "first")
		testutil.NotOk(t, err)
		testutil.Equals(t, fmt.Errorf("api with name stage doesn't exist"), err)
	})

	t.Run("config with one API no tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: nil},
			},
		}

		err := cfg.SetCurrentContext(tlogger, "stage", "first")

		testutil.NotOk(t, err)
		testutil.Equals(t, fmt.Errorf("tenant with name first doesn't exist in api stage"), err)
	})

	t.Run("config with multiple API and no current", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.SetCurrentContext(tlogger, "prod", "first"))

		testutil.Equals(t, cfg.Current, struct {
			API    string `json:"api"`
			Tenant string `json:"tenant"`
		}{
			API:    "prod",
			Tenant: "first",
		})
	})

	t.Run("config with multiple API and current", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
			Current: struct {
				API    string `json:"api"`
				Tenant string `json:"tenant"`
			}{
				API:    "stage",
				Tenant: "second",
			},
		}

		testutil.Ok(t, cfg.SetCurrentContext(tlogger, "prod", "first"))

		testutil.Equals(t, cfg.Current, struct {
			API    string `json:"api"`
			Tenant string `json:"tenant"`
		}{
			API:    "prod",
			Tenant: "first",
		})
	})
}

func TestRemoveContext(t *testing.T) {
	tmpDir := t.TempDir()
	t.Cleanup(func() { testutil.Ok(t, os.RemoveAll(tmpDir)) })
	testutil.Ok(t, os.MkdirAll(filepath.Join(tmpDir, "obsctl", "test"), os.ModePerm))
	testutil.Ok(t, os.WriteFile(filepath.Join(tmpDir, "obsctl", "test", "config.json"), []byte(""), os.ModePerm))
	testutil.Ok(t, os.Setenv("OBSCTL_CONFIG_PATH", filepath.Join(tmpDir, "obsctl", "test", "config.json")))

	tlogger := level.NewFilter(log.NewJSONLogger(log.NewSyncWriter(os.Stderr)), level.AllowDebug())

	t.Run("empty config", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
		}

		err := cfg.RemoveContext(tlogger, "stage", "first")
		testutil.NotOk(t, err)
		testutil.Equals(t, fmt.Errorf("api with name stage doesn't exist"), err)
	})

	t.Run("config with one API no tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: nil},
			},
		}

		err := cfg.RemoveContext(tlogger, "stage", "first")

		testutil.NotOk(t, err)
		testutil.Equals(t, fmt.Errorf("tenant with name first doesn't exist in api stage"), err)
	})

	t.Run("config with one API and one tenant", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.RemoveContext(tlogger, "stage", "first"))

		testutil.Equals(t, cfg.APIs, map[string]APIConfig{})
	})

	t.Run("config with multiple APIs and tenants", func(t *testing.T) {
		cfg := Config{
			pathOverride: filepath.Join(tmpDir, "obsctl", "test", "config.json"),
			APIs: map[string]APIConfig{
				"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
				"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
					"first":  {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
					"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
				}},
			},
		}

		testutil.Ok(t, cfg.RemoveContext(tlogger, "stage", "second"))
		testutil.Ok(t, cfg.RemoveContext(tlogger, "prod", "first"))

		exp := map[string]APIConfig{
			"stage": {URL: "https://stage.api:9090", Contexts: map[string]TenantConfig{
				"first": {Tenant: "first", OIDC: &OIDCConfig{Audience: "obs", ClientID: "first", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
			"prod": {URL: "https://prod.api:9090", Contexts: map[string]TenantConfig{
				"second": {Tenant: "second", OIDC: &OIDCConfig{Audience: "obs", ClientID: "second", ClientSecret: "secret", IssuerURL: "sso.obs.com"}},
			}},
		}

		testutil.Equals(t, cfg.APIs, exp)
	})
}
