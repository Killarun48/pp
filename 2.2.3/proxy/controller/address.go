package controller

import (
	"encoding/json"
	"net/http"

	_ "test/responder"
)

// @Summary Геокодирование (координаты по адресу)
// @Tags api
// @Accept json
// @Produce json
// @Param query body RequestAddressSearch true "Адрес"
// @Success 200 {object} ResponseAddress
// @Failure 400,403 {object} responder.Response
// @Router /api/address/search [post]
func (c *Controller) Search(w http.ResponseWriter, r *http.Request) {
	var body RequestAddressSearch

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		c.responder.ErrorBadRequest(w, err)
		return
	}

	resp, err := c.geoService.AddressSearch(body.Query)
	if err != nil {
		c.responder.ErrorBadRequest(w, err)
		return
	}

	c.responder.OutputJSON(w, ResponseAddress{
		Addresses: resp,
	})
}

// @Summary Обратное геокодирование (адрес по координатам)
// @Tags api
// @Accept json
// @Produce json
// @Param lat,lng body RequestAddressGeocode true "Координаты. lat - Географическая широта. lng - Географическая долгота."
// @Success 200 {object} ResponseAddress
// @Failure 400,403 {object} responder.Response
// @Router /api/address/geocode [post]
func (c *Controller) Geocode(w http.ResponseWriter, r *http.Request) {
	var body RequestAddressGeocode

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		c.responder.ErrorBadRequest(w, err)
		return
	}

	resp, err := c.geoService.GeoCode(body.Lat, body.Lng)
	if err != nil {
		c.responder.ErrorBadRequest(w, err)
		return
	}
	
	c.responder.OutputJSON(w, ResponseAddress{
		Addresses: resp,
	})
}
