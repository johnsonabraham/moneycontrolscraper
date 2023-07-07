package stores

import (
	"github.com/johnsonabraham/moneycontrolscraper/config"
	"github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/models"
	"github.com/kataras/golog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MoneycontrolDataServie interface {
	InsertMoneyControlSymbols([]models.CompanyInfo) error
	FetchCompanyByNameConstant(companyName string) (models.CompanyInfo, error)
}

type moneycontrolDataService struct {
	db      *gorm.DB
	envVars *config.AppEnvVars
	vlog    *golog.Logger
}

func NewMoneycontrolDataService(db *gorm.DB, vlog *golog.Logger, envVars *config.AppEnvVars) *moneycontrolDataService {
	return &moneycontrolDataService{
		db:      db,
		vlog:    vlog,
		envVars: envVars,
	}
}

func (s *moneycontrolDataService) InsertMoneyControlSymbols(result []models.CompanyInfo) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		er := s.db.Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&result).Error
		return er
	})
	if err != nil {
		s.vlog.Error(err)
		return err
	}
	return err
}

func (s *moneycontrolDataService) FetchCompanyByNameConstant(companyName string) (models.CompanyInfo, error) {
	var company models.CompanyInfo
	err := s.db.Transaction(func(tx *gorm.DB) error {
		er := s.db.Where("company_name=?", companyName).First(&company).Error
		return er
	})
	if err != nil {
		s.vlog.Error(err)
		return company, err
	}
	return company, nil
}
