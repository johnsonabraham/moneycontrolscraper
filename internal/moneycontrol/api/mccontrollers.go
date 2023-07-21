package moneycontrolapi

import (
	"errors"
	"fmt"

	"github.com/johnsonabraham/moneycontrolscraper/config"
	"github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/models"
	service "github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/service"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
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

func (h *MoneyControlHandler) CollectDividendData(ctx iris.Context) {
	company := ctx.URLParam("company")

	h.mlog.Info(fmt.Sprintf("Moneycontrol dividend collection started for %s", company))
	err := h.moneyControlService.ScrapeDividendHistory(company)
	if err != nil {
		var errMsg string
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errMsg = "Company not found"
		} else {
			errMsg = "Something went wrong, please try again after some time"
			h.mlog.Error(fmt.Sprintf("Error scraping dividend data for %s", company))
		}

		failedRes := models.FailedResponse{
			Status:   iris.StatusInternalServerError,
			ErrorMsg: errMsg,
		}
		ctx.StopWithJSON(
			iris.StatusNotFound,
			failedRes,
		)
		return
	}
	h.mlog.Info(fmt.Sprintf("Dividend history scraped for company %s", company))

	response := models.Response{
		Status: "success",
		Msg:    "Dividend data saved successfully",
	}
	ctx.StatusCode(iris.StatusOK)
	ctx.JSON(response)
}

func (h *MoneyControlHandler) CollectHistoricalDailyDate(ctx iris.Context) {
	company := ctx.URLParam("company")

	h.mlog.Info(fmt.Sprintf("Moneycontrol historical data collection started for %s", company))
	err := h.moneyControlService.CaptureHistoricalData(company)
	if err != nil {
		var errMsg string
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errMsg = "Company not found"
		} else {
			errMsg = "Something went wrong, please try again after some time"
			h.mlog.Error(fmt.Sprintf("Error saving historical daily data for %s", company))
		}

		failedRes := models.FailedResponse{
			Status:   iris.StatusInternalServerError,
			ErrorMsg: errMsg,
		}
		ctx.StopWithJSON(
			iris.StatusNotFound,
			failedRes,
		)
		return
	}
	h.mlog.Info(fmt.Sprintf("Historical daily data collected for company %s", company))

	response := models.Response{
		Status: "success",
		Msg:    "Historical daily data saved successfully",
	}
	ctx.StatusCode(iris.StatusOK)
	ctx.JSON(response)
}
