package repository

import (
	"github.com/johnsonabraham/moneycontrolscraper/config"
	"github.com/johnsonabraham/moneycontrolscraper/internal/moneycontrol/models"
	"github.com/kataras/golog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MoneycontrolRepository interface {
	InsertMoneyControlSymbols([]models.CompanyInfo) error
	FetchCompanyByNameConstant(companyName string) (*models.CompanyInfo, error)
	UpdateSymbol(result models.CompanyInfo) error
}

type moneycontrolRepository struct {
	db      *gorm.DB
	envVars *config.AppEnvVars
	vlog    *golog.Logger
}

func NewMoneycontrolRepository(db *gorm.DB, vlog *golog.Logger, envVars *config.AppEnvVars) *moneycontrolRepository {
	return &moneycontrolRepository{
		db:      db,
		vlog:    vlog,
		envVars: envVars,
	}
}

func (s *moneycontrolRepository) InsertMoneyControlSymbols(result []models.CompanyInfo) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		er := s.db.Exec("delete from company_infos").Error
		if er != nil {
			return er
		}
		er = s.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "symbol"}},
			DoNothing: true,
		}).Create(&result).Error
		return er
	})
	if err != nil {
		s.vlog.Error(err)
		return err
	}
	return err
}

func (s *moneycontrolRepository) UpdateSymbol(result models.CompanyInfo) error {
	err := s.db.Transaction(func(tx *gorm.DB) error {
		er := s.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "symbol"}},
			UpdateAll: true,
		}).Save(&result).Error
		return er
	})
	if err != nil {
		s.vlog.Error(err)
		return err
	}
	return err
}

func (s *moneycontrolRepository) FetchCompanyByNameConstant(companyName string) (*models.CompanyInfo, error) {
	var company models.CompanyInfo
	err := s.db.Transaction(func(tx *gorm.DB) error {
		er := s.db.Where("nse_id=?", companyName).First(&company).Error
		return er
	})
	if err != nil {
		s.vlog.Error(err)
		return &company, err
	}
	return &company, nil
}
