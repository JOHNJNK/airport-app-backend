package controllers

import (
	"airport-app-backend/models"
	"airport-app-backend/repositories"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type AirlineController struct {
	repository repositories.IAirlineRepository
}

func NewAirlineController(repository repositories.IAirlineRepository) *AirlineController {
	return &AirlineController{
		repository: repository,
	}
}

// @Summary			Get all airlines
// @Router 			/airlines [get]
// @Description 	Gets all the airlines
// @ID 				get-all-airlines
// @Tags 			airline
// @Produce  		json
// @Param   		page	query	int		false	"Page number (default = 0)"
// @Success 		200		"ok"
// @Failure 		500		"Internal server error"
func (ac *AirlineController) HandleGetAllAirlines(ctx *gin.Context) {
	// TODO: Convert to using a pagination library to handle this and other edge cases
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "0"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid page parameter"})
		return
	}
	if page < 0 {
		ctx.JSON(400, gin.H{"msg": "Page number must be greater than 0"})
		return
	}

	airlines, err := ac.repository.GetAllAirlines(page)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"Error": "Internal server error"})
		return
	}
	ctx.JSON(http.StatusOK, airlines)
}

// @Summary			Get airline by Id
// @Router			/airline/{id} [get]
// @Description 	Gets airline by Id
// @ID 				get-airline-by-id
// @Tags 			airline
// @Produce  		json
// @Param   		id		path		string		true		"Airline Id"
// @Success 		200		"ok"
// @Failure 		400		"Airline not found"
func (ac *AirlineController) HandleGetAirline(ctx *gin.Context) {
	airlineId := ctx.Param("id")

	airline, err := ac.repository.GetAirline(airlineId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error": "Incorrect airline id: " + airlineId})
		return
	}
	ctx.JSON(http.StatusOK, airline)
}

// @Summary			Create new airline
// @Router			/airline [post]
// @Description 	Create new airline
// @ID 				create-airline
// @Tags 			airline
// @Produce  		json
// @Param   		airline		body		models.Airline		true		"Airline Object"
// @Success 		200		"ok"
// @Failure 		400		" Airline not found"
func (ac *AirlineController) HandleCreateNewAirline(ctx *gin.Context) {
	var airline models.Airline

	err := ctx.ShouldBindWith(&airline, binding.JSON)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	repositoryError := ac.repository.CreateNewAirline(&airline)
	if repositoryError != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error": repositoryError.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, "Created a new airline successfully")
}

// @Summary Delete airline by Id
// @Router /airline/{id} [delete]
// @Summary Delete airline by Id
// @Description Delete the airline details of the particular id
// @ID delete-airline-by-id
// @Tags airline
// @Param id path string true "Airline Id"
// @Success 200  "ok"
// @Failure 400 "Airline not found"
func (ac *AirlineController) HandleDeleteAirline(ctx *gin.Context) {
	airlineId := ctx.Param("id")
	err := ac.repository.DeleteAirline(airlineId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error": "Incorrect airline id: " + airlineId})
		return
	}
	ctx.JSON(http.StatusOK, "Deleted the airline successfully")
}

// @Summary Update airline by ID
// @Router /airline/{id} [put]
// @Summary Update airline by ID
// @Description update the airline details by its ID
// @ID update-airline
// @Tags airline
// @Produce  json
// @Param id path string true "Airline ID"
// @Param airline body models.Airline true "Updated airline object"
// @Success 200  "ok"
// @Failure 500 "Internal server error"
// @Failure 400 "Bad request"
func (ac *AirlineController) HandleUpdateAirline(ctx *gin.Context) {
	airlineId := ctx.Param(`id`)
	var airline models.Airline

	err := ctx.ShouldBindWith(&airline, binding.JSON)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
		return
	}

	if len(airline.Id) > 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error": "ID cannot be updated"})
		return
	}

	repositoryError := ac.repository.UpdateAirline(&airline, airlineId)
	if repositoryError != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"Error": repositoryError.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"message": "update success"})
}
