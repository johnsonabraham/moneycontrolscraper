package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/johnsonabraham/moneycontrolscraper/config"
	"github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/models"
	repository "github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/repository"
	"github.com/kataras/golog"
	"github.com/kataras/iris/v12"
)

var (
	stocksURL = make(models.StocksInfo)
	baseURL   = "https://www.moneycontrol.com/technical-analysis"
)

type CompanyAdditionalDetailsJson struct {
	Data data `json:"data"`
}
type data struct {
	BSEID        string                 `json:"BSEID"`
	MKTCAP       float64                `json:"MKTCAP"`
	NSEID        string                 `json:"NSEID"`
	MainSector   string                 `json:"main_sector"`
	NewSubSector string                 `json:"newSubsector"`
	Best5Set     map[string]interface{} `json:"best_5_set"`
}

// GetCompanyList returns the list of all the companies tracked via stockrate
func GetCompanyList() (list []string) {
	for key := range stocksURL {
		list = append(list, key)
	}
	return
}

type MoneycontrolService interface {
	CaptureSymbols() error
	ScrapeDividendHistory(companyName string) error
	CaptureHistoricalData(ticker string) error
}

func NewMoneyControlService(mlog *golog.Logger, cfg *config.AppEnvVars, moneycontrolRepository repository.MoneycontrolRepository) *moneyControlService {
	return &moneyControlService{
		mlog:                   mlog,
		cfg:                    cfg,
		moneycontrolRepository: moneycontrolRepository,
	}
}

type moneyControlService struct {
	mlog                   *golog.Logger
	cfg                    *config.AppEnvVars
	moneycontrolRepository repository.MoneycontrolRepository
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
		return stockPrice, fmt.Errorf("error in reading stock Price")
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
		return nil, fmt.Errorf("error in reading stock Technicals %v", err.Error())
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
		return nil, fmt.Errorf("error in reading stock Moving Averages %v", err.Error())
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
		return nil, fmt.Errorf("error in reading stock Pivot Levels %v", err.Error())
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

// getURL checks whether we can read data for company and returns its data source URL
func getURL(company string) (URL string, err error) {
	if val, found := stocksURL[strings.ToLower(company)]; found {
		URL = baseURL + "/" + val.Company + "/" + val.Symbol + "/daily"
		return
	}
	return "", fmt.Errorf("company not found")
}

// Here stocks information necessary is saved and stored, which is calculated everytime package is imported
func (i *moneyControlService) CaptureSymbols() error {
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
	if err := i.moneycontrolRepository.InsertMoneyControlSymbols(companyInfos); err != nil {
		i.mlog.Error("Error while saving Symbols")
		return err
	}
	i.mlog.Info(fmt.Sprintf("Captured %s Symbols", strconv.Itoa(len(companyInfos))))
	go i.CaptureAdditionalCompanyInfo(companyInfos)
	return nil
}

func (i *moneyControlService) CaptureAdditionalCompanyInfo(companyInfos []models.CompanyInfo) {
	for _, companyInfo := range companyInfos {
		time.Sleep(5 * time.Second)
		response, err := http.Get(fmt.Sprintf(i.cfg.MoneyControlCompDetailsUrl, companyInfo.Symbol))
		if err != nil {
			i.mlog.Error(fmt.Sprintf("Error while saving gathering additional data for %s", companyInfo.Symbol), err)
		}
		defer response.Body.Close()

		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			i.mlog.Error(fmt.Sprintf("Failed to read the response body while fetching addition data for %s:",
				companyInfo.Symbol), err)
			continue
		}
		var additionalDetails CompanyAdditionalDetailsJson
		err = json.Unmarshal(body, &additionalDetails)
		if err != nil {
			i.mlog.Error(fmt.Sprintf("Failed to unmarshall the response body while fetching addition data for %s:",
				companyInfo.Symbol), err)
			continue
		}
		companyInfo.BSEID = additionalDetails.Data.BSEID
		companyInfo.MarketCap = additionalDetails.Data.MKTCAP
		companyInfo.NSEID = additionalDetails.Data.NSEID
		companyInfo.MainSectorDetails = additionalDetails.Data.MainSector
		companyInfo.SubSectorDetails = additionalDetails.Data.NewSubSector
		if err := i.moneycontrolRepository.UpdateSymbol(companyInfo); err != nil {
			i.mlog.Error(fmt.Sprintf("Failed to update additional company info for %s:",
				companyInfo.Symbol), err)
			continue
		}
		i.mlog.Info(fmt.Sprintf("Done collecting additional info for %s", companyInfo.Company))
	}
	i.mlog.Info("Done collecting additional info for companies")
}

// Captures and stores dividend data of the provided company
func (i *moneyControlService) ScrapeDividendHistory(ticker string) error {
	var dividendHistory []models.Dividend
	companyInfo, err := i.moneycontrolRepository.FetchCompanyByNameConstant(ticker)
	if err != nil {
		i.mlog.Error("Error fetching provided company", err)
		return err
	}
	response, err := http.Get(fmt.Sprintf(i.cfg.MoneyControlDividendURL, companyInfo.CompanyName, companyInfo.Symbol))
	if err != nil {
		i.mlog.Error(fmt.Sprintf("Error while scraping dividend data for %s", ticker), err)
	}
	defer response.Body.Close()

	// Parse HTML document
	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		i.mlog.Error(fmt.Sprintf("Error parsing response of dividend page for %s", ticker), err)
	}
	// Scrape historical dividend data
	doc.Find("table.mctable1>tbody>tr").Each(func(count int, s *goquery.Selection) {
		var dividend models.Dividend
		var dateFormat = "02-01-2006"
		t, err := time.Parse(dateFormat, s.Find("td:nth-child(1)").Text())
		if err != nil {
			i.mlog.Error(fmt.Sprintf("Error converting dividend announcement date for %s", ticker), err)
		}
		dividend.AnnouncementDate = t.Unix()
		t, err = time.Parse(dateFormat, s.Find("td:nth-child(2)").Text())
		if err != nil {
			i.mlog.Error(fmt.Sprintf("Error converting dividend ex date for %s", ticker), err)
		}
		dividend.ExDate = t.Unix()
		dividend.DividendType = s.Find("td:nth-child(3)").Text()
		dividend.DividendPercentage, err = strconv.ParseFloat(s.Find("td:nth-child(4)").Text(), 64)
		if err != nil {
			i.mlog.Error(fmt.Sprintf("Error converting dividend percentage to float for %s", ticker), err)
		}
		dividend.Dividend, err = strconv.ParseFloat(s.Find("td:nth-child(5)").Text(), 64)
		if err != nil {
			i.mlog.Error(fmt.Sprintf("Error converting dividend amount to float for %s", ticker), err)
		}
		dividend.Remark = s.Find("td:nth-child(6)").Text()
		dividendHistory = append(dividendHistory, dividend)

	})
	token, err := i.GetMoneyBSToken(i.cfg)
	if err != nil {
		i.mlog.Error("Error while generating token for MoneyBS", err)
		return err
	}
	jsonBody, err := json.Marshal(dividendHistory)
	if err != nil {
		i.mlog.Error("Error while marshalling dividend data", err)
		return nil
	}

	req, err := http.NewRequest("POST", fmt.Sprintf(i.cfg.MoneyBSBaseURL+i.cfg.MoneyBSHistoricalDividendDataEndpoint, companyInfo.NSEID), bytes.NewBuffer(jsonBody))
	if err != nil {
		i.mlog.Error(err)
		return err
	}
	var bearer = "Bearer " + *token

	req.Header.Set("Authorization", bearer)
	client := &http.Client{}
	response, err = client.Do(req)
	if err != nil {
		i.mlog.Error(err)
		return err
	}
	defer response.Body.Close()
	return nil
}

func (i *moneyControlService) CaptureHistoricalData(ticker string) error {
	companyInfo, err := i.moneycontrolRepository.FetchCompanyByNameConstant(ticker)
	if err != nil {
		i.mlog.Error("Error fetching provided company", err)
		return err
	}
	url := fmt.Sprintf(i.cfg.MoneyControlHistoricalDataUrl, companyInfo.NSEID, fmt.Sprint(time.Now().Unix()))
	response, err := http.Get(url)
	if err != nil {
		i.mlog.Error(err)
		return err
	}
	if response.StatusCode != iris.StatusOK {
		i.mlog.Error(err)
		return err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		i.mlog.Error(fmt.Sprintf("Failed to read the response body while fetching addition data for %s:",
			ticker), err)
		return err
	}
	token, err := i.GetMoneyBSToken(i.cfg)
	if err != nil {
		i.mlog.Error("Error while generating token for MoneyBS", err)
		return err
	}
	req, err := http.NewRequest("POST", fmt.Sprintf(i.cfg.MoneyBSBaseURL+i.cfg.MoneyBSHistoricalDataEndpoint, ticker), bytes.NewBuffer(body))
	if err != nil {
		i.mlog.Error("Error creating post request to MoneyBS", err)
		return err
	}
	var bearer = "Bearer " + *token

	req.Header.Set("Authorization", bearer)
	client := &http.Client{}
	go client.Do(req)
	return nil
}

func (i *moneyControlService) GetMoneyBSToken(cfg *config.AppEnvVars) (*string, error) {
	req, err := http.NewRequest("GET", cfg.MoneyBSBaseURL+cfg.MoneyBSAuthEndpoint, nil)
	if err != nil {
		i.mlog.Error(err)
		return nil, err
	}
	req.Header.Add("x-api-key", cfg.MoneyBSAPIKey)
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		i.mlog.Error(err)
		return nil, err
	}

	defer response.Body.Close()
	// Read response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		i.mlog.Error(err)
		return nil, err
	}
	b := string(body)
	return &b, nil
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
