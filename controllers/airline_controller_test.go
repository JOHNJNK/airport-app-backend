package controllers

import (
	"airport-app-backend/mocks"
	"airport-app-backend/models"
	"airport-app-backend/models/factory"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

var AIRLINES = "/airlines"
var AIRLINE = "/airline"

var mockAirlineRepository *mocks.MockIAirlineRepository
var airlineController *AirlineController
var airlineContext *gin.Context
var airlineResponseRecorder *httptest.ResponseRecorder

func beforeEachAirlineTest(t *testing.T) {
	gomockController := gomock.NewController(t)
	defer gomockController.Finish()

	mockAirlineRepository = mocks.NewMockIAirlineRepository(gomockController)
	airlineController = NewAirlineController(mockAirlineRepository)
	airlineResponseRecorder = httptest.NewRecorder()
	airlineContext, _ = gin.CreateTestContext(airlineResponseRecorder)
}

func TestHandleGetAllAirlines(t *testing.T) {
	beforeEachAirlineTest(t)
	var airlines []models.Airline
	airline1 := factory.ConstructAirline()
	airlines = append(airlines, airline1)
	airline2 := factory.ConstructAirline()
	airlines = append(airlines, airline2)
	airline3 := factory.ConstructAirline()
	airlines = append(airlines, airline3)

	mockAirlineRepository.EXPECT().GetAllAirlines(gomock.Any()).Return(airlines, nil)
	airlineContext.Request, _ = http.NewRequest(http.MethodGet, AIRLINES, nil)

	airlineController.HandleGetAllAirlines(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	var airlinesFromResponse []models.Airline
	assert.NoError(t, json.Unmarshal(responseBody, &airlinesFromResponse))

	assert.Equal(t, 3, len(airlinesFromResponse))
	assert.Contains(t, airlinesFromResponse, airline1)
	assert.Contains(t, airlinesFromResponse, airline2)
	assert.Contains(t, airlinesFromResponse, airline3)
}

func TestHandleGetAllAirlinesWhenPageIsNonNumeric(t *testing.T) {
	beforeEachAirlineTest(t)
	airlineContext.Request, _ = http.NewRequest(http.MethodGet, AIRLINES+"?page=abc", nil)

	airlineController.HandleGetAllAirlines(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "{\"error\":\"invalid page parameter\"}", string(responseBody))
}

// TestHandleGetAllAirlinesWhenPageIsNegative verifies that ?page=-1 returns 400
// because page numbers must be >= 0.
func TestHandleGetAllAirlinesWhenPageIsNegative(t *testing.T) {
	beforeEachAirlineTest(t)
	airlineContext.Request, _ = http.NewRequest(http.MethodGet, AIRLINES+"?page=-1", nil)

	airlineController.HandleGetAllAirlines(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "{\"msg\":\"Page number must be greater than 0\"}", string(responseBody))
}

func TestHandleGetAllAirlinesWhenServiceReturnsError(t *testing.T) {
	beforeEachAirlineTest(t)
	mockAirlineRepository.EXPECT().GetAllAirlines(gomock.Any()).Return(nil, errors.New("Invalid"))
	airlineContext.Request, _ = http.NewRequest(http.MethodGet, AIRLINES, nil)

	airlineController.HandleGetAllAirlines(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "{\"Error\":\"Internal server error\"}", string(responseBody))
}

func TestHandleGetAirline(t *testing.T) {
	beforeEachAirlineTest(t)
	airline := factory.ConstructAirline()
	airlineId := "123"
	mockAirlineRepository.EXPECT().GetAirline(airlineId).Return(&airline, nil)
	airlineContext.Request, _ = http.NewRequest(http.MethodGet, AIRLINE, nil)
	airlineContext.AddParam("id", airlineId)

	airlineController.HandleGetAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	var airlineFromResponse models.Airline
	assert.NoError(t, json.Unmarshal(responseBody, &airlineFromResponse))

	assert.Equal(t, airline, airlineFromResponse)
}

func TestHandleGetAirlineWhenRecordDoesntExist(t *testing.T) {
	beforeEachAirlineTest(t)
	nonExistentAirlineId := "-23243"
	mockAirlineRepository.EXPECT().GetAirline(nonExistentAirlineId).Return(nil, errors.New("foo bar"))
	airlineContext.Request, _ = http.NewRequest("GET", AIRLINE, nil)
	airlineContext.AddParam("id", nonExistentAirlineId)

	airlineController.HandleGetAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, fmt.Sprintf("{\"Error\":\"Incorrect airline id: %s\"}", nonExistentAirlineId), string(responseBody))
}

func TestHandleCreateNewAirline(t *testing.T) {
	beforeEachAirlineTest(t)
	airline := factory.ConstructAirline()
	reqBody, _ := json.Marshal(&airline)
	airlineContext.Request, _ = http.NewRequest(http.MethodPost, AIRLINE, strings.NewReader(string(reqBody)))
	mockAirlineRepository.EXPECT().CreateNewAirline(&airline).Return(nil)

	airlineController.HandleCreateNewAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusCreated, response.StatusCode)
}

func TestHandleCreateNewAirlineWhenTheRequestPayloadIsEmpty(t *testing.T) {
	beforeEachAirlineTest(t)
	reqBody := `{}`
	airlineContext.Request, _ = http.NewRequest(http.MethodPost, AIRLINE, strings.NewReader(reqBody))

	airlineController.HandleCreateNewAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "{\"Error\":\"Key: 'Airline.Name' Error:Field validation for 'Name' failed on the 'required' tag\"}", string(responseBody))
}

func TestHandleCreateNewAirlineWhenTheMandatoryValueIsAbsent(t *testing.T) {
	beforeEachAirlineTest(t)
	reqBody := `{"Name":""}`
	airlineContext.Request, _ = http.NewRequest(http.MethodPost, AIRLINE, strings.NewReader(reqBody))

	airlineController.HandleCreateNewAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "{\"Error\":\"Key: 'Airline.Name' Error:Field validation for 'Name' failed on the 'required' tag\"}", string(responseBody))
}

func TestHandleCreateNewAirlineWhenTheMandatoryKeyIsAbsent(t *testing.T) {
	beforeEachAirlineTest(t)
	reqBody := `{"Count":2}`
	airlineContext.Request, _ = http.NewRequest(http.MethodPost, AIRLINE, strings.NewReader(reqBody))

	airlineController.HandleCreateNewAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "{\"Error\":\"Key: 'Airline.Name' Error:Field validation for 'Name' failed on the 'required' tag\"}", string(responseBody))
}

func TestHandleCreateNewAirlineWhenDataOfDifferentDatatypeIsGiven(t *testing.T) {
	beforeEachAirlineTest(t)
	reqBody := `{"name":123}`
	airlineContext.Request, _ = http.NewRequest(http.MethodPost, AIRLINE, strings.NewReader(reqBody))

	airlineController.HandleCreateNewAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "{\"Error\":\"json: cannot unmarshal number into Go struct field Airline.name of type string\"}", string(responseBody))
}

func TestHandleCreateNewAirlineWhereErrorIsThrownInRepositoryLayer(t *testing.T) {
	beforeEachAirlineTest(t)
	reqBody := `{"name":"Test"}`
	mockAirlineRepository.EXPECT().CreateNewAirline(gomock.Any()).Return(errors.New("invalid request"))
	airlineContext.Request, _ = http.NewRequest(http.MethodPost, AIRLINE, strings.NewReader(reqBody))

	airlineController.HandleCreateNewAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "{\"Error\":\"invalid request\"}", string(responseBody))
}

func TestHandleDeleteAirline(t *testing.T) {
	beforeEachAirlineTest(t)
	airlineId := "123"
	mockAirlineRepository.EXPECT().DeleteAirline(airlineId).Return(nil)
	airlineContext.Request, _ = http.NewRequest(http.MethodDelete, AIRLINE, nil)
	airlineContext.AddParam("id", airlineId)

	airlineController.HandleDeleteAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "\"Deleted the airline successfully\"", string(responseBody))
}

func TestHandleDeleteNewAirlineWhereErrorIsThrownInRepositoryLayer(t *testing.T) {
	beforeEachAirlineTest(t)
	nonExistentAirlineId := "-23243"
	mockAirlineRepository.EXPECT().DeleteAirline(nonExistentAirlineId).Return(errors.New("invalid request"))
	airlineContext.Request, _ = http.NewRequest(http.MethodDelete, AIRLINE, nil)
	airlineContext.AddParam("id", nonExistentAirlineId)

	airlineController.HandleDeleteAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, fmt.Sprintf("{\"Error\":\"Incorrect airline id: %s\"}", nonExistentAirlineId), string(responseBody))
}

func TestHandleUpdateAirline(t *testing.T) {
	beforeEachAirlineTest(t)
	airlineId := "1"
	airline := factory.ConstructAirline()
	reqBody, _ := json.Marshal(airline)
	airlineContext.AddParam("id", airlineId)
	mockAirlineRepository.EXPECT().UpdateAirline(&airline, airlineId).Return(nil)
	airlineContext.Request, _ = http.NewRequest(http.MethodPut, AIRLINE, strings.NewReader(string(reqBody)))

	airlineController.HandleUpdateAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusCreated, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "{\"message\":\"update success\"}", string(responseBody))
}

func TestHandleUpdateAirlineWhenTheRequestPayloadIsEmpty(t *testing.T) {
	beforeEachAirlineTest(t)
	reqBody := `{}`
	airlineContext.Request, _ = http.NewRequest(http.MethodPut, AIRLINE, strings.NewReader(reqBody))

	airlineController.HandleUpdateAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "{\"Error\":\"Key: 'Airline.Name' Error:Field validation for 'Name' failed on the 'required' tag\"}", string(responseBody))
}

func TestHandleUpdateAirlineWhenTheMandatoryValueIsAbsent(t *testing.T) {
	beforeEachAirlineTest(t)
	reqBody := `{"Name":""}`
	airlineContext.Request, _ = http.NewRequest(http.MethodPut, AIRLINE, strings.NewReader(reqBody))

	airlineController.HandleUpdateAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "{\"Error\":\"Key: 'Airline.Name' Error:Field validation for 'Name' failed on the 'required' tag\"}", string(responseBody))
}

func TestHandleUpdateAirlineWhereErrorIsThrownInRepositoryLayer(t *testing.T) {
	beforeEachAirlineTest(t)
	invalidId := "-1"
	airline := factory.ConstructAirline()
	airlineContext.AddParam("id", invalidId)
	reqBody, _ := json.Marshal(&airline)
	mockAirlineRepository.EXPECT().UpdateAirline(&airline, invalidId).Return(errors.New("invalid Request"))
	airlineContext.Request, _ = http.NewRequest(http.MethodPut, AIRLINE, strings.NewReader(string(reqBody)))

	airlineController.HandleUpdateAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "{\"Error\":\"invalid Request\"}", string(responseBody))
}

func TestHandleUpdateAirlineWhereErrorIsThrownWhenIdIsUpdated(t *testing.T) {
	beforeEachAirlineTest(t)
	reqBody := `{"name":"Test", "id":"56yfh"}`
	mockAirlineRepository.EXPECT().UpdateAirline(gomock.Any(), gomock.Any()).Return(nil)
	airlineContext.Request, _ = http.NewRequest(http.MethodPut, AIRLINE, strings.NewReader(reqBody))

	airlineController.HandleUpdateAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "{\"Error\":\"ID cannot be updated\"}", string(responseBody))
}

func TestHandleUpdateAirlineWhenTheMandatoryKeyIsAbsent(t *testing.T) {
	beforeEachAirlineTest(t)
	reqBody := `{"Count":2}`
	airlineContext.Request, _ = http.NewRequest(http.MethodPut, AIRLINE, strings.NewReader(reqBody))

	airlineController.HandleUpdateAirline(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "{\"Error\":\"Key: 'Airline.Name' Error:Field validation for 'Name' failed on the 'required' tag\"}", string(responseBody))
}

// --- SearchAirlinesByName tests ---

// TestHandleSearchAirlinesByNameFound verifies that when ?name=<term> matches airlines,
// the handler returns 200 with the matching results.
func TestHandleSearchAirlinesByNameFound(t *testing.T) {
	beforeEachAirlineTest(t)

	airline1 := models.Airline{Name: "British Airways"}
	airline2 := models.Airline{Name: "British Midland"}
	expectedAirlines := []models.Airline{airline1, airline2}

	mockAirlineRepository.EXPECT().SearchAirlinesByName("British").Return(expectedAirlines, nil)
	airlineContext.Request, _ = http.NewRequest(http.MethodGet, AIRLINES+"?name=British", nil)

	airlineController.HandleGetAllAirlines(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	var airlinesFromResponse []models.Airline
	assert.NoError(t, json.Unmarshal(responseBody, &airlinesFromResponse))

	assert.Equal(t, 2, len(airlinesFromResponse))
	assert.Contains(t, airlinesFromResponse, airline1)
	assert.Contains(t, airlinesFromResponse, airline2)
}

// TestHandleSearchAirlinesByNameNotFound verifies that when ?name=<term> matches no
// airlines, the handler returns 200 with an empty JSON array (not null).
func TestHandleSearchAirlinesByNameNotFound(t *testing.T) {
	beforeEachAirlineTest(t)

	mockAirlineRepository.EXPECT().SearchAirlinesByName("nonexistent").Return([]models.Airline{}, nil)
	airlineContext.Request, _ = http.NewRequest(http.MethodGet, AIRLINES+"?name=nonexistent", nil)

	airlineController.HandleGetAllAirlines(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	var airlinesFromResponse []models.Airline
	assert.NoError(t, json.Unmarshal(responseBody, &airlinesFromResponse))

	assert.Equal(t, 0, len(airlinesFromResponse))
	// Confirm wire format is [] not null
	assert.Equal(t, "[]", string(responseBody))
}

// TestHandleSearchAirlinesByNameNilSliceNormalized verifies that when the repository
// returns (nil, nil) — a valid empty-result signal — the handler normalises the nil
// slice to an empty slice so the JSON response is [] rather than null.
func TestHandleSearchAirlinesByNameNilSliceNormalized(t *testing.T) {
	beforeEachAirlineTest(t)

	// Repository returns nil slice (not an empty allocated slice) with no error.
	mockAirlineRepository.EXPECT().SearchAirlinesByName("air").Return(nil, nil)
	airlineContext.Request, _ = http.NewRequest(http.MethodGet, AIRLINES+"?name=air", nil)

	airlineController.HandleGetAllAirlines(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	// Must be the JSON empty-array literal, not "null".
	assert.Equal(t, "[]", string(responseBody))
}

// TestHandleSearchAirlinesByNameEmptyFallsThroughToPagination verifies that when
// ?name= (empty string) is provided, the handler falls through to the normal
// pagination path and SearchAirlinesByName is never called.
func TestHandleSearchAirlinesByNameEmptyFallsThroughToPagination(t *testing.T) {
	beforeEachAirlineTest(t)

	airline1 := factory.ConstructAirline()
	airlines := []models.Airline{airline1}

	// SearchAirlinesByName must NOT be called — only GetAllAirlines.
	mockAirlineRepository.EXPECT().GetAllAirlines(gomock.Any()).Return(airlines, nil)
	airlineContext.Request, _ = http.NewRequest(http.MethodGet, AIRLINES+"?name=", nil)

	airlineController.HandleGetAllAirlines(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	var airlinesFromResponse []models.Airline
	assert.NoError(t, json.Unmarshal(responseBody, &airlinesFromResponse))

	assert.Equal(t, 1, len(airlinesFromResponse))
}

// TestHandleSearchAirlinesByNameRepositoryError verifies that when
// SearchAirlinesByName returns an error the handler responds with 500 and the
// standard uppercase-key error envelope.
func TestHandleSearchAirlinesByNameRepositoryError(t *testing.T) {
	beforeEachAirlineTest(t)

	mockAirlineRepository.EXPECT().SearchAirlinesByName("air").Return(nil, errors.New("db error"))
	airlineContext.Request, _ = http.NewRequest(http.MethodGet, AIRLINES+"?name=air", nil)

	airlineController.HandleGetAllAirlines(airlineContext)

	response := airlineResponseRecorder.Result()
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	assert.Equal(t, "{\"Error\":\"Internal server error\"}", string(responseBody))
}

// TestHandleSearchAirlinesByNameIgnoresPageParam verifies that when both ?name and
// ?page are supplied, the name-search path takes priority and ?page is silently
// ignored — even when the page value would otherwise be invalid (e.g. -1).
func TestHandleSearchAirlinesByNameIgnoresPageParam(t *testing.T) {
	beforeEachAirlineTest(t)

	airline1 := models.Airline{Name: "Air France"}
	expectedAirlines := []models.Airline{airline1}

	// The name path must fire; GetAllAirlines must NOT be called.
	mockAirlineRepository.EXPECT().SearchAirlinesByName("Air").Return(expectedAirlines, nil)
	airlineContext.Request, _ = http.NewRequest(http.MethodGet, AIRLINES+"?name=Air&page=-1", nil)

	airlineController.HandleGetAllAirlines(airlineContext)

	response := airlineResponseRecorder.Result()
	// 200 — name search ran, negative page was never evaluated.
	assert.Equal(t, http.StatusOK, response.StatusCode)

	responseBody, _ := io.ReadAll(response.Body)
	var airlinesFromResponse []models.Airline
	assert.NoError(t, json.Unmarshal(responseBody, &airlinesFromResponse))

	assert.Equal(t, 1, len(airlinesFromResponse))
	assert.Contains(t, airlinesFromResponse, airline1)
}
