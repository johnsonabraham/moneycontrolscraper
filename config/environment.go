package config

import (
	"errors"

	"github.com/caarlos0/env/v7"
	"github.com/kataras/golog"
	"github.com/sirupsen/logrus"
)

var errLoadingEnvVar = errors.New("failed to load the env vars")

type AppEnvVars struct {
	PGHostIP                              string `env:"PG_HOST"`
	PGUser                                string `env:"PG_USER"`
	PGPassword                            string `env:"PG_PASSWORD"`
	PGDbName                              string `env:"PG_DBNAME"`
	MsEnv                                 string `env:"MS_ENV"`
	APIKey                                string `env:"API_KEY"`
	MoneyControlSymbolURL                 string `env:"MONEYCONTROL_SYMBOL_URL"`
	MoneyControlDividendURL               string `env:"MONEYCONTROL_DIVIDEND_URL"`
	MoneyControlCompDetailsUrl            string `env:"MONEYCONTROL_COMP_DETAILS_URL"`
	MoneyControlHistoricalDataUrl         string `env:"MONEYCONTROL_HISTORICAL_DATA_URL"`
	MoneyBSAPIKey                         string `env:"MONEYBS_API_KEY"`
	MoneyBSBaseURL                        string `env:"MONEYBS_BASE_URL"`
	MoneyBSAuthEndpoint                   string `env:"MONEYBS_AUTH_ENDPOINT"`
	MoneyBSHistoricalDataEndpoint         string `env:"MONEYBS_HISTORICAL_DATA_ENDPOINT"`
	MoneyBSHistoricalDividendDataEndpoint string `env:"MONEYBS_HISTORICAL_DIVIDEND_DATA_ENDPOINT"`
	AppPort                               string `env:"APP_PORT"`
}

func LoadEnvVars(vlog *golog.Logger) *AppEnvVars {
	var appEnvVars AppEnvVars

	opts := &env.Options{RequiredIfNoDef: true}
	if err := env.Parse(&appEnvVars, *opts); err != nil {
		logrus.Fatalf("%s : %s", errLoadingEnvVar, err)
	}

	vlog.Info("Environment Variables Loaded..")
	vlog.Debug("loaded env vars %v", appEnvVars)

	return &appEnvVars
}
