package repositories

import (
	"airport-app-backend/models"
)

type IAirlineRepository interface {
	GetAllAirlines(int) ([]models.Airline, error)
	SearchAirlinesByName(name string) ([]models.Airline, error)
	GetAirline(string) (*models.Airline, error)
	CreateNewAirline(*models.Airline) error
	UpdateAirline(airline *models.Airline, airlineId string) error
	DeleteAirline(string) error
}

var DEFAULT_PAGE_LIMIT int = 10

func (sr *ServiceRepository) GetAllAirlines(pageNum int) ([]models.Airline, error) {
	var airline []models.Airline
	result := sr.db.Limit(DEFAULT_PAGE_LIMIT).Offset(pageNum * DEFAULT_PAGE_LIMIT).Find(&airline)
	if result.Error != nil {
		return nil, result.Error
	}
	return airline, nil
}

func (sr *ServiceRepository) SearchAirlinesByName(name string) ([]models.Airline, error) {
	var airlines []models.Airline
	err := sr.db.Where("LOWER(name) LIKE LOWER(?)", "%"+name+"%").Find(&airlines).Error
	return airlines, err
}

func (sr *ServiceRepository) GetAirline(id string) (*models.Airline, error) {
	var airline *models.Airline
	result := sr.db.First(&airline, "id=?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return airline, nil
}

func (sr *ServiceRepository) CreateNewAirline(airline *models.Airline) error {
	result := sr.db.Save(airline)
	return result.Error
}

func (sr *ServiceRepository) UpdateAirline(airline *models.Airline, airlineId string) error {
	result := sr.db.Where("id = ?", airlineId).Updates(airline)
	return result.Error
}

func (sr *ServiceRepository) DeleteAirline(id string) error {
	var airline *models.Airline
	result := sr.db.Delete(&airline, "id=?", id)
	return result.Error
}
