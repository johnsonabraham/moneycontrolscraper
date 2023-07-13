package models

type Dividend struct {
	AnnouncementDate   int64
	ExDate             int64
	DividendType       string
	DividendPercentage float64
	Dividend           float64
	Remark             string
	NSEID              string
	MarketCap          string
	SectorDetails      string
	BSEID              string
	MoreData           string
}

type CompanyInfo struct {
	ID          int64 `gorm:"primary_key NOT NULL AUTO_INCREMENT"`
	CompanyName string
	Company     string
	Sector      string
	Symbol      string
}

type (
	// StocksInfo holds the information related to datasource of stocks
	StocksInfo map[string]StockURLValue

	// StockPrice holds current price, previous close, open, variation, percentage, volume of stocks from BSE and NSE
	StockPrice struct {
		BSE SymbolPriceValue
		NSE SymbolPriceValue
	}

	// StockTechnicals holds stock technical valuations
	StockTechnicals map[string]TechnicalValue

	// StockMovingAverage holds Moving average for 5, 10, 15, 20, 50, 100, 200 days respectively
	StockMovingAverage map[int]MovingAverageValue

	// StockPivotLevels stores a stock pivote levels
	StockPivotLevels map[string]PivotPointsValue

	StockURLValue struct {
		Sector  string
		Company string
		Symbol  string
	}

	SymbolPriceValue struct {
		Price         float64
		PreviousClose float64
		Open          float64
		Variation     float64
		Percentage    float64
		Volume        int64
	}

	TechnicalValue struct {
		Level      float64
		Indication string
	}

	MovingAverageValue struct {
		SMA        float64
		Indication string
	}

	PivotPointsValue struct {
		R1    float64
		R2    float64
		R3    float64
		Pivot float64
		S1    float64
		S2    float64
		S3    float64
	}
)
