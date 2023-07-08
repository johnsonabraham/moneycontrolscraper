package moneycontrolapi

import (
	"fmt"
	"strconv"

	"github.com/johnsonabraham/moneycontrolscraper/config"
	"github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/models"
	"github.com/johnsonabraham/moneycontrolscraper/internal/stores"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
)

type MoneyControlHandler struct {
	service stores.MoneycontrolDataServie
	mlog    *golog.Logger
	cfg     *config.AppEnvVars
}

func NewMoneyControlHandler(store stores.MoneycontrolDataServie, vlog *golog.Logger, cfg *config.AppEnvVars) *MoneyControlHandler {
	return &MoneyControlHandler{
		service: store,
		mlog:    vlog,
		cfg:     cfg,
	}
}

func (h *MoneyControlHandler) CollectMoneycontrolSymbols(ctx iris.Context) {
	h.mlog.Info("Moneycontrol symbol collection started")
	mcDataCollection, _ := NewMoneyControllDataCollection(h.mlog, h.cfg, h.service)
	moneyControlSymbols := mcDataCollection.CaptureSymbols()
	if err := h.service.InsertMoneyControlSymbols(moneyControlSymbols); err != nil {
		h.mlog.Error("Error while saving Symbols")
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
	h.mlog.Info(fmt.Sprintf("Captured %s Symbols", strconv.Itoa(len(moneyControlSymbols))))
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
	mcDataCollection, _ := NewMoneyControllDataCollection(h.mlog, h.cfg, h.service)
	dividendHistory, err := mcDataCollection.ScrapeDividendHistory(company)
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
