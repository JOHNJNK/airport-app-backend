package repositories

import (
	"airport-app-backend/models"
	"strconv"

	"github.com/rs/zerolog/log"
)

var DEFAULT_PAGE_SIZE = 10

type IGateRepository interface {
	GetAllGates(page int, floor string) ([]models.Gate, error)
	GetGate(string) (*models.Gate, error)
	CreateNewGate(*models.Gate) error
	UpdateGate(string, models.Gate) error
	DeleteGate(id string) error
}

func (sr *ServiceRepository) GetAllGates(page int, floorStr string) ([]models.Gate, error) {
	log.Debug().Msg("Fetching list of gates")

	var gates []models.Gate
	offset := page * DEFAULT_PAGE_SIZE
	query := sr.db.Offset(offset).Limit(DEFAULT_PAGE_SIZE)
	if floorStr != "*" {
		floor, _ := strconv.Atoi(floorStr)
		query = query.Where("floor_number = ?", floor)
	}
  query = query.Order("gate_number ASC")
	if err := query.Find(&gates).Error; err != nil {
		return nil, err
	}
	return gates, nil
}

func (sr *ServiceRepository) GetGate(id string) (*models.Gate, error) {
	log.Debug().Msg("repository layer for retrieving gate details by id")

	var gate models.Gate
	if err := sr.db.Where("id = ?", id).First(&gate).Error; err != nil {
		return nil, err
	}
	return &gate, nil
}

func (sr *ServiceRepository) CreateNewGate(gate *models.Gate) error {
	err := sr.db.Save(gate)
	return err.Error
}

func (sr *ServiceRepository) UpdateGate(id string, gate models.Gate) error {
	err := sr.db.Where("id=?", id).Updates(gate)
	return err.Error
}

func (sr *ServiceRepository) DeleteGate(id string) error {
	var gate models.Gate
	result := sr.db.Delete(&gate, "id=?", id)
	return result.Error
}
