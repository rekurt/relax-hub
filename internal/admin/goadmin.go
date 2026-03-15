package admin

import (
	"html/template"
	"strings"
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
	env := gaconfig.EnvLocal
	if strings.EqualFold(cfg.Environment, "production") {
		env = gaconfig.EnvProd
	}

	gaCfg := &gaconfig.Config{
		Env: env,
		Databases: gaconfig.DatabaseList{
			"default": {
				Driver:          gaconfig.DriverPostgresql,
				Dsn:             cfg.Database.DSN,
				MaxIdleConns:    5,
				MaxOpenConns:    10,
				ConnMaxLifetime: time.Hour,
			},
		},
		UrlPrefix: cfg.Admin.Prefix,
		Theme:     cfg.Admin.Theme,
		Language:  cfg.Admin.Language,
		Store: gaconfig.Store{
			Path:   "./uploads",
			Prefix: "uploads",
		},
		Title:           "Бани - Админ-панель",
		Logo:            template.HTML(`<b>Бани</b>`),
		MiniLogo:        template.HTML(`<b>Б</b>`),
		IndexUrl:        "/pages/",
		Debug:           false,
		SessionLifeTime: 7200,

		HideConfigCenterEntrance:      true,
		HideToolEntrance:              true,
		HideAppInfoEntrance:           true,
		HidePluginEntrance:            true,
		HideVisitorUserCenterEntrance: true,
	}

	return gaCfg
}
