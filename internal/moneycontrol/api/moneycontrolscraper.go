package moneycontrolapi

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/johnsonabraham/moneycontrolscraper/config"
	"github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/models"
	"github.com/johnsonabraham/moneycontrolscraper/internal/stores"
	"github.com/kataras/golog"
)

var (
	stocksURL = make(models.StocksInfo)
	baseURL   = "https://www.moneycontrol.com/technical-analysis"
)

// GetCompanyList returns the list of all the companies tracked via stockrate
func GetCompanyList() (list []string) {
	for key := range stocksURL {
		list = append(list, key)
	}
	return
}
func NewMoneyControllDataCollection(mlog *golog.Logger, cfg *config.AppEnvVars, service stores.MoneycontrolDataServie) (*mcDataCollection, error) {
	return &mcDataCollection{
		mlog:    mlog,
		cfg:     cfg,
		service: service,
	}, nil
}

type mcDataCollection struct {
	mlog    *golog.Logger
	cfg     *config.AppEnvVars
	service stores.MoneycontrolDataServie
}

// GetPrice returns current price, previous close, open, variation, percentage and volume for a company
func GetPrice(company string) (models.StockPrice, error) {
	var stockPrice models.StockPrice
	url, err := getURL(company)
	if err != nil {
		return stockPrice, err
	}
	doc, err := getStockQuote(url)
	if err != nil {
		return stockPrice, fmt.Errorf("Error in reading stock Price")
	}
	doc.Find(".bsedata_bx").Each(func(i int, s *goquery.Selection) {
		stockPrice.BSE.Price, _ = strconv.ParseFloat(s.Find(".span_price_wrap").Text(), 64)
		stockPrice.BSE.PreviousClose, _ = strconv.ParseFloat(s.Find(".priceprevclose").Text(), 64)
		stockPrice.BSE.Open, _ = strconv.ParseFloat(s.Find(".priceopen").Text(), 64)
		stockPrice.BSE.Variation, _ = strconv.ParseFloat(strings.Split(s.Find(".span_price_change_prcnt").Text(), " ")[0], 64)
		stockPrice.BSE.Percentage, _ = strconv.ParseFloat(strings.Split(strings.Split(s.Find(".span_price_change_prcnt").Text(), "%")[0], "(")[1], 64)
		stockPrice.BSE.Volume, _ = strconv.ParseInt(strings.ReplaceAll(s.Find(".volume_data").Text(), ",", ""), 10, 64)
	})
	doc.Find(".nsedata_bx").Each(func(i int, s *goquery.Selection) {
		stockPrice.NSE.Price, _ = strconv.ParseFloat(s.Find(".span_price_wrap").Text(), 64)
		stockPrice.NSE.PreviousClose, _ = strconv.ParseFloat(s.Find(".priceprevclose").Text(), 64)
		stockPrice.NSE.Open, _ = strconv.ParseFloat(s.Find(".priceopen").Text(), 64)
		stockPrice.NSE.Variation, _ = strconv.ParseFloat(strings.Split(s.Find(".span_price_change_prcnt").Text(), " ")[0], 64)
		stockPrice.NSE.Percentage, _ = strconv.ParseFloat(strings.Split(strings.Split(s.Find(".span_price_change_prcnt").Text(), "%")[0], "(")[1], 64)
		stockPrice.NSE.Volume, _ = strconv.ParseInt(strings.ReplaceAll(s.Find(".volume_data").Text(), ",", ""), 10, 64)
	})
	return stockPrice, nil
}

// GetTechnicals returns the technical valuations of a company with indications
func GetTechnicals(company string) (models.StockTechnicals, error) {
	stockTechnicals := make(models.StockTechnicals)
	url, err := getURL(company)
	if err != nil {
		return nil, err
	}
	doc, err := getStockQuote(url)
	if err != nil {
		return nil, fmt.Errorf("Error in reading stock Technicals %v", err.Error())
	}
	doc.Find("#techindd").Find("tbody").Find("tr").Each(func(i int, s *goquery.Selection) {
		symbol := strings.Split(strings.Split(s.Find("td").First().Text(), "(")[0], "%")[0]
		level, _ := strconv.ParseFloat(strings.ReplaceAll(s.Find("td").Find("strong").First().Text(), ",", ""), 64)
		indication := s.Find("td").Find("strong").Last().Text()
		if symbol != "" && symbol != "Bollinger Band(20,2)" {
			stockTechnicals[symbol] = models.TechnicalValue{Level: level, Indication: indication}
		}
	})
	return stockTechnicals, nil
}

// GetMovingAverage returns the 5, 10, 20, 50, 100, 200 days moving average respectively
func GetMovingAverage(company string) (models.StockMovingAverage, error) {
	stockMovingAverage := make(models.StockMovingAverage)
	url, err := getURL(company)
	if err != nil {
		return nil, err
	}
	doc, err := getStockQuote(url)
	if err != nil {
		return nil, fmt.Errorf("Error in reading stock Moving Averages %v", err.Error())
	}
	doc.Find("#movingavgd").Find("tbody").Find("tr").Each(func(i int, s *goquery.Selection) {
		period, _ := strconv.Atoi(s.Find("td").First().Text())
		sma, _ := strconv.ParseFloat(strings.ReplaceAll(s.Find("td").Find("strong").First().Text(), ",", ""), 64)
		indication := s.Find("td").Find("strong").Last().Text()
		if period != 0 {
			stockMovingAverage[period] = models.MovingAverageValue{SMA: sma, Indication: indication}
		}
	})
	return stockMovingAverage, nil
}

// GetPivotLevels returns the important pivot levels of a stock given in order R1, R2, R3, Pivot, S1, S2, S3
func GetPivotLevels(company string) (models.StockPivotLevels, error) {
	stockPivotLevels := make(models.StockPivotLevels)
	url, err := getURL(company)
	if err != nil {
		return nil, err
	}
	doc, err := getStockQuote(url)
	if err != nil {
		return nil, fmt.Errorf("Error in reading stock Pivot Levels %v", err.Error())
	}
	doc.Find("#pevotld").Find("table").First().Find("tbody").Find("tr").Each(func(i int, s *goquery.Selection) {
		pivotType := s.Find("td").First().Text()
		if pivotType != "" {
			var levels []float64
			s.Find("td").Next().Each(func(i int, s *goquery.Selection) {
				level, _ := strconv.ParseFloat(strings.ReplaceAll(s.Text(), ",", ""), 64)
				levels = append(levels, level)
			})
			stockPivotLevels[pivotType] = models.PivotPointsValue{
				R1: levels[0], R2: levels[1], R3: levels[2], Pivot: levels[3], S1: levels[4], S2: levels[5], S3: levels[6],
			}
		}
	})
	return stockPivotLevels, nil
}

// getStockQuote creates and returns the web document from a web URL
func getStockQuote(URL string) (*goquery.Document, error) {
	res, err := http.Get(URL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// getURL checks whether we can read data for company and returns its data source URL
func getURL(company string) (URL string, err error) {
	if val, found := stocksURL[strings.ToLower(company)]; found {
		URL = baseURL + "/" + val.Company + "/" + val.Symbol + "/daily"
		return
	}
	return "", fmt.Errorf("Company Not Found")
}

// Here stocks information necessary is saved and stored, which is calculated everytime package is imported
func (i *mcDataCollection) CaptureSymbols() []models.CompanyInfo {
	var companyInfos []models.CompanyInfo
	capAlphabets := []string{"A", "B", "C", "D", "E", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}
	for _, char := range capAlphabets {
		doc, err := getStockQuote(i.cfg.MoneyControlSymbolURL + char)
		if err != nil {
			i.mlog.Error("Error in fetching stock URLs ", err.Error())
		}
		doc.Find(".bl_12").Each(func(i int, s *goquery.Selection) {
			var companyInfo models.CompanyInfo
			link, _ := s.Attr("href")
			stockName := s.Text()
			if match, _ := regexp.MatchString(`^(http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/)?[a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5}(:[0-9]{1,5})?(\/.*)?$`, link); match {
				stockURLSplit := strings.Split(link, "/")
				companyInfo.Company = strings.ToLower(stockName)
				companyInfo.Sector = stockURLSplit[5]
				companyInfo.CompanyName = stockURLSplit[6]
				companyInfo.Symbol = stockURLSplit[7]
				companyInfos = append(companyInfos, companyInfo)
			}
		})
	}
	return companyInfos
}

// Captures and stores dividend data of the provided company
func (i *mcDataCollection) ScrapeDividendHistory(companyName string) ([]models.Dividend, error) {
	var dividendHistory []models.Dividend
	companyInfo, err := i.service.FetchCompanyByNameConstant(companyName)
	if err != nil {
		i.mlog.Error("Error fetching provided company", err)
		return dividendHistory, err
	}
	i.mlog.Info(fmt.Sprintf(i.cfg.MoneyControlDividendURL, companyInfo.CompanyName, companyInfo.Sector))
	response, err := http.Get(fmt.Sprintf(i.cfg.MoneyControlDividendURL, companyInfo.CompanyName, companyInfo.Symbol))
	if err != nil {
		i.mlog.Error(fmt.Sprintf("Error while scraping dividend data for %s", companyName), err)
	}
	defer response.Body.Close()

	// Parse HTML document
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		i.mlog.Error(fmt.Sprintf("Error parsing response of dividend page for %s", companyName), err)
	}
	// Scrape historical dividend data
	doc.Find("table.mctable1>tbody>tr").Each(func(count int, s *goquery.Selection) {
		var dividend models.Dividend
		var dateFormat = "02-01-2006"
		t, err := time.Parse(dateFormat, s.Find("td:nth-child(1)").Text())
		if err != nil {
			i.mlog.Error(fmt.Sprintf("Error converting dividend announcement date for %s", companyName), err)
		}
		dividend.AnnouncementDate = t.Unix()
		t, err = time.Parse(dateFormat, s.Find("td:nth-child(2)").Text())
		if err != nil {
			i.mlog.Error(fmt.Sprintf("Error converting dividend ex date for %s", companyName), err)
		}
		dividend.ExDate = t.Unix()
		dividend.DividendType = s.Find("td:nth-child(3)").Text()
		dividend.DividendPercentage, err = strconv.ParseFloat(s.Find("td:nth-child(4)").Text(), 64)
		if err != nil {
			i.mlog.Error(fmt.Sprintf("Error converting dividend percentage to float for %s", companyName), err)
		}
		dividend.Dividend, err = strconv.ParseFloat(s.Find("td:nth-child(5)").Text(), 64)
		if err != nil {
			i.mlog.Error(fmt.Sprintf("Error converting dividend amount to float for %s", companyName), err)
		}
		dividend.Remark = s.Find("td:nth-child(6)").Text()
		dividendHistory = append(dividendHistory, dividend)

	})
	return dividendHistory, nil
}
