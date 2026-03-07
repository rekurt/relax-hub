package admin

import (
	"html/template"
	"time"

	_ "github.com/GoAdminGroup/go-admin/adapter/chi"
	gaconfig "github.com/GoAdminGroup/go-admin/modules/config"
	_ "github.com/GoAdminGroup/go-admin/modules/db/drivers/postgres"
	_ "github.com/GoAdminGroup/themes/adminlte"
	_ "github.com/GoAdminGroup/themes/sword"

	appconfig "github.com/nikitaaldaev/bani/config"
)

// BuildGoAdminConfig creates a GoAdmin config from the app config.
// This is separated from engine initialization because GoAdmin's AddConfig
// immediately connects to the database, so it must be deferred to fx lifecycle.
func BuildGoAdminConfig(cfg *appconfig.Config) *gaconfig.Config {
	theme := cfg.Admin.Theme
	if theme == "" {
		theme = "adminlte"
	}

	prefix := cfg.Admin.Prefix
	if prefix == "" {
		prefix = "/admin-panel"
	}

	language := cfg.Admin.Language
	if language == "" {
		language = "ru"
	}

	gaCfg := &gaconfig.Config{
		Env: gaconfig.EnvLocal,
		Databases: gaconfig.DatabaseList{
			"default": {
				Driver:          gaconfig.DriverPostgresql,
				Dsn:             cfg.Database.DSN,
				MaxIdleConns:    5,
				MaxOpenConns:    10,
				ConnMaxLifetime: time.Hour,
			},
		},
		UrlPrefix: prefix,
		Theme:     theme,
		Language:  language,
		Store: gaconfig.Store{
			Path:   "./uploads",
			Prefix: "uploads",
		},
		Title:           "Бани - Админ-панель",
		Logo:            template.HTML(`<b>Бани</b>`),
		MiniLogo:        template.HTML(`<b>Б</b>`),
		IndexUrl:        "/pages/",
		Debug:           true,
		SessionLifeTime: 7200,

		HideConfigCenterEntrance:      true,
		HideToolEntrance:              true,
		HideAppInfoEntrance:           true,
		HidePluginEntrance:            true,
		HideVisitorUserCenterEntrance: true,
	}

	if cfg.Storage.Endpoint != "" {
		gaCfg.FileUploadEngine = gaconfig.FileUploadEngine{
			Name: "local",
		}
	}

	return gaCfg
}
