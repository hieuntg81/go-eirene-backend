package request

// GetNotificationsRequest represents notification list request
type GetNotificationsRequest struct {
	Page  int `form:"page" validate:"omitempty,min=1"`
	Limit int `form:"limit" validate:"omitempty,min=1,max=100"`
}

// GetDefaultPage returns the page number or default
func (r *GetNotificationsRequest) GetDefaultPage() int {
	if r.Page <= 0 {
		return 1
	}
	return r.Page
}

// GetDefaultLimit returns the limit or default
func (r *GetNotificationsRequest) GetDefaultLimit() int {
	if r.Limit <= 0 {
		return 20
	}
	return r.Limit
}

// GetOffset calculates the offset for pagination
func (r *GetNotificationsRequest) GetOffset() int {
	return (r.GetDefaultPage() - 1) * r.GetDefaultLimit()
}

// GeocodeReverseRequest represents reverse geocoding request
type GeocodeReverseRequest struct {
	Latitude  float64 `form:"lat" validate:"required,min=-90,max=90"`
	Longitude float64 `form:"lng" validate:"required,min=-180,max=180"`
}

// GeocodeSearchRequest represents geocoding search request
type GeocodeSearchRequest struct {
	Query string `form:"q" validate:"required,min=2"`
	Limit int    `form:"limit" validate:"omitempty,min=1,max=10"`
}

// GetDefaultLimit returns the limit or default
func (r *GeocodeSearchRequest) GetDefaultLimit() int {
	if r.Limit <= 0 {
		return 5
	}
	return r.Limit
}
