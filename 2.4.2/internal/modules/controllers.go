package modules

import (
	"app/internal/infrastructure/responder"
	"app/internal/modules/user/controller"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type Controller struct {
	User controller.UserControllerer
}

func NewController(services *Service, respond responder.Responder) *Controller {
	return &Controller{
		User: controller.NewUserController(services.User, respond),
	}
}

func (c *Controller) InitRoutesUser() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Route("/users", func(r chi.Router) {
		r.Post("/", c.User.Create)
		r.Get("/", c.User.List)

		r.Route("/{userID}", func(r chi.Router) {
			r.Get("/", c.User.GetByID)
			r.Post("/", c.User.Update)
			r.Delete("/", c.User.Delete)
		})
	})

	return r
}