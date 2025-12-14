package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"bamboo-rescue/internal/handler/dto/response"
	"bamboo-rescue/internal/middleware"
	"bamboo-rescue/internal/service"
	pkgresponse "bamboo-rescue/pkg/response"
)

// GeocodeHandler handles geocoding requests
type GeocodeHandler struct {
	geocodeService service.GeocodeService
}

// NewGeocodeHandler creates a new GeocodeHandler
func NewGeocodeHandler(geocodeService service.GeocodeService) *GeocodeHandler {
	return &GeocodeHandler{
		geocodeService: geocodeService,
	}
}

// ReverseGeocode handles reverse geocoding
// @Summary Reverse geocode
// @Description Get address from coordinates
// @Tags Geocode
// @Produce json
// @Param latitude query number true "Latitude"
// @Param longitude query number true "Longitude"
// @Success 200 {object} pkgresponse.Response{data=response.GeocodeResponse}
// @Failure 400 {object} pkgresponse.Response
// @Router /geocode/reverse [get]
func (h *GeocodeHandler) ReverseGeocode(c *gin.Context) {
	latStr := c.Query("latitude")
	lngStr := c.Query("longitude")

	if latStr == "" || lngStr == "" {
		pkgresponse.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Latitude and longitude are required", 400))
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		pkgresponse.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid latitude", 400))
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		pkgresponse.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Invalid longitude", 400))
		return
	}

	result, err := h.geocodeService.ReverseGeocode(c.Request.Context(), lat, lng)
	if err != nil {
		pkgresponse.Error(c, err)
		return
	}

	pkgresponse.Success(c, http.StatusOK, response.GeocodeResponse{
		Address: result.Address,
		Location: response.GeoPointResponse{
			Latitude:  result.Location.Latitude,
			Longitude: result.Location.Longitude,
		},
		PlaceID:   result.PlaceID,
		PlaceType: result.PlaceType,
	})
}

// SearchAddress handles address search
// @Summary Search address
// @Description Search for addresses
// @Tags Geocode
// @Produce json
// @Param query query string true "Search query"
// @Success 200 {object} pkgresponse.Response{data=[]response.GeocodeResponse}
// @Failure 400 {object} pkgresponse.Response
// @Router /geocode/search [get]
func (h *GeocodeHandler) SearchAddress(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		pkgresponse.Error(c, middleware.NewAppError("VALIDATION_ERROR", "Query is required", 400))
		return
	}

	results, err := h.geocodeService.SearchAddress(c.Request.Context(), query)
	if err != nil {
		pkgresponse.Error(c, err)
		return
	}

	responses := make([]response.GeocodeResponse, len(results))
	for i, r := range results {
		responses[i] = response.GeocodeResponse{
			Address: r.Address,
			Location: response.GeoPointResponse{
				Latitude:  r.Location.Latitude,
				Longitude: r.Location.Longitude,
			},
			PlaceID:   r.PlaceID,
			PlaceType: r.PlaceType,
		}
	}

	pkgresponse.Success(c, http.StatusOK, responses)
}
