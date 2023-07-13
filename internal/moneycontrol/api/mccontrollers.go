package moneycontrolapi

import (
	"fmt"

	"github.com/johnsonabraham/moneycontrolscraper/config"
	"github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/models"
	service "github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/service"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
)

type MoneyControlHandler struct {
	moneyControlService service.MoneycontrolService
	mlog                *golog.Logger
	cfg                 *config.AppEnvVars
}

func NewMoneyControlHandler(service service.MoneycontrolService, vlog *golog.Logger, cfg *config.AppEnvVars) *MoneyControlHandler {
	return &MoneyControlHandler{
		moneyControlService: service,
		mlog:                vlog,
		cfg:                 cfg,
	}
}

func (h *MoneyControlHandler) CollectMoneycontrolSymbols(ctx iris.Context) {
	h.mlog.Info("Moneycontrol symbol collection started")
	if err := h.moneyControlService.CaptureSymbols(); err != nil {
		failedRes := models.FailedResponse{
			Status:   iris.StatusInternalServerError,
			ErrorMsg: "Something went wrong, please try again after some time",
		}
		ctx.StopWithJSON(
			iris.StatusNotFound,
			failedRes,
		)
		return
	}
	h.mlog.Info("Moneycontrol symbol collection ended")
	response := models.Response{
		Status: "success",
		Msg:    "Symbols collected successfully",
	}
	ctx.StatusCode(iris.StatusOK)
	ctx.JSON(response)
}

func (h *MoneyControlHandler) ScrapeDividendData(ctx iris.Context) {
	company := ctx.URLParam("company")

	h.mlog.Info(fmt.Sprintf("Moneycontrol dividend collection started for %s", company))
	dividendHistory, err := h.moneyControlService.ScrapeDividendHistory(company)
	if err != nil {
		h.mlog.Error("Error scraping dividend data for %s", company)
		failedRes := models.FailedResponse{
			Status:   iris.StatusInternalServerError,
			ErrorMsg: "Something went wrong, please try again after some time",
		}
		ctx.StopWithJSON(
			iris.StatusNotFound,
			failedRes,
		)
		return
	}
	h.mlog.Info(fmt.Sprintf("Dividend history scraped for company %s", company))

	ctx.StatusCode(iris.StatusOK)
	ctx.JSON(dividendHistory)
}
