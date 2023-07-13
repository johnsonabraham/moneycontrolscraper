package main

import (
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/jwt"

	"github.com/iris-contrib/swagger/swaggerFiles"
	"github.com/iris-contrib/swagger/v12"

	"github.com/johnsonabraham/moneycontrolscraper/config"
	"github.com/johnsonabraham/moneycontrolscraper/internal/auth"
	"github.com/johnsonabraham/moneycontrolscraper/pkg/log"

	api "github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/api"
	repository "github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/repository"
	service "github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/service"
	P "github.com/johnsonabraham/moneycontrolscraper/internal/persist"
)

var (
	errStartingMoneybsMS = errors.New("failed to start the app")
	tokenMaxTimeOut      = 30000 * time.Minute
)

func main() {
	applog := log.New("info", "json")
	applog.Info("starting Moneycontrol app.")

	app := iris.New()
	app.Validator = validator.New()

	applog.Info("swagger config started")

	swaggerConfig := &swagger.Config{
		// The url pointing to API definition.
		URL:         "http://localhost:8091/swagger/doc.json",
		DeepLinking: true,
	}

	swaggerUI := swagger.CustomWrapHandler(swaggerConfig, swaggerFiles.Handler)

	applog.Info("swagger config completed")
	applog.Info("loading env vars")

	cfg := config.LoadEnvVars(applog)

	mlog := log.NewEnvLog(cfg.MsEnv)

	app.Get("/swagger/{any:path}", swaggerUI)

	signer := jwt.NewSigner(jwt.HS512, []byte(cfg.APIKey), tokenMaxTimeOut)

	verifier := jwt.NewVerifier(jwt.HS512, []byte(cfg.APIKey))
	verifier.WithDefaultBlocklist()
	verifyMiddleware := verifier.Verify(func() interface{} {
		return new(auth.UserClaims)
	})

	app.OnErrorCode(iris.StatusNotFound, func(ctx iris.Context) {
		if err := ctx.JSON(iris.Map{
			"error": "Given handler not found. Try /* or /api/v1/*",
		}); err != nil {
			mlog.Error("failed to respond: ", err)
		}
	})

	app.Get("/hc", api.AppHealthCheck)

	app.Get("/status", api.AppStatus)

	db := P.ConnectDB(cfg)

	moneyControlRepository := repository.NewMoneycontrolRepository(db, mlog, cfg)
	moneyControlService := service.NewMoneyControlService(mlog, cfg, moneyControlRepository)
	moneyControlHandler := api.NewMoneyControlHandler(moneyControlService, mlog, cfg)

	apiv1 := app.Party("/api/v1")
	apiv1.Get("/auth", auth.GenerateToken(signer, cfg))
	apiv1.Use(verifyMiddleware)

	apiv1.Get("/collectCompanySymbols", moneyControlHandler.CollectMoneycontrolSymbols)
	apiv1.Get("/scrapeDividendHistory", moneyControlHandler.ScrapeDividendData)
	port := ":" + cfg.AppPort
	if err := app.Listen(port, iris.WithOptimizations); err != nil {
		app.Logger().Fatalf("%s: due to :%s", errStartingMoneybsMS, err)
	}

	mlog.Info("MoneyBS app started.")
}
