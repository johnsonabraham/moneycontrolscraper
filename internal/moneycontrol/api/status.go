package moneycontrolapi

import (
	"github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/models"
	errhandler "github.com/johnsonabraham/moneycontrolscraper/pkg/errorhandler"
	"github.com/kataras/iris/v12"
)

func AppStatus(ctx iris.Context) {
	response := models.Status{
		AppVersion:    "moneybs-1.0",
		IsDBConnected: false,
	}

	errhandler.Res(ctx.JSON(response))
}

func AppHealthCheck(ctx iris.Context) {
	response := models.HC{
		App: "OK",
	}

	errhandler.Res(ctx.JSON(response))
}
