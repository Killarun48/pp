package controller

import (
	"test/geoservice"
	"test/responder"
)

type Controller struct {
	responder responder.Responder
	geoService geoservice.GeoProvider
}

type ControllerOption func(*Controller)

func WithResponder(responder responder.Responder) ControllerOption {
	return func(c *Controller) {
		c.responder = responder
	}
}

func WithGeoService(geoService geoservice.GeoProvider) ControllerOption {
	return func(c *Controller) {
		c.geoService = geoService
	}
}

func NewController(options ...ControllerOption) *Controller {
	controller := &Controller{}

	for _, option := range options {
		option(controller)
	}

	return controller
}
