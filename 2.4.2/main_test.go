package main

import (
	"app/internal/infrastructure/responder"
	"app/internal/models"
	"app/internal/modules/user/controller"
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

type mockUserService struct {
	create  func(ctx context.Context, user models.User) error
	getByID func(ctx context.Context, id string) (models.User, error)
	update  func(ctx context.Context, user models.User) error
	delete  func(ctx context.Context, id string) error
	list    func(ctx context.Context, c models.Conditions) ([]models.User, error)
}

func (m mockUserService) Create(ctx context.Context, user models.User) error {
	return m.create(ctx, user)
}

func (m mockUserService) GetByID(ctx context.Context, id string) (models.User, error) {
	return m.getByID(ctx, id)
}

func (m mockUserService) Update(ctx context.Context, user models.User) error {
	return m.update(ctx, user)
}

func (m mockUserService) Delete(ctx context.Context, id string) error {
	return m.delete(ctx, id)
}

func (m mockUserService) List(ctx context.Context, c models.Conditions) ([]models.User, error) {
	return m.list(ctx, c)
}

func TestUser(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		userService func(ctx context.Context, user models.User) error
		wantStatus  int
		wantBody    string
	}{
		{
			name: "valid body",
			body: `{"name": "Rodney William Whitaker","email": "dasdasd"}`,
			userService: func(ctx context.Context, user models.User) error {
				return nil
			},
			wantStatus: 200,
			wantBody:   "{\"success\":true,\"message\":\"user created\"}\n",
		},
		{
			name: "error create user",
			body: `{"name": "Rodney William Whitaker","email": "dsfsddf"}`,
			userService: func(ctx context.Context, user models.User) error {
				return fmt.Errorf("error")
			},
			wantStatus: 400,
			wantBody:   "{\"success\":false,\"message\":\"error\"}\n",
		},
		{
			name: "invalid body",
			body: `{"name" "Rodney William Whitaker"}`,
			userService: func(ctx context.Context, user models.User) error {
				return nil
			},
			wantStatus: 400,
			wantBody:   "{\"success\":false,\"message\":\"invalid character '\\\"' after object key\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest("POST", "/api/users/", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")
			userService := mockUserService{
				create: tt.userService,
			}
			responder := responder.NewResponder()
			ac := controller.NewUserController(userService, responder)
			ac.Create(w, r)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}

func TestServer(t *testing.T) {
	pathDB = "./test.db"
	s := NewServer(":8088")

	go s.Serve()

	s.Stop()
	time.Sleep(2 * time.Second)

	os.Remove("./test.db")
}

func TestUserUpdate(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		userService func(ctx context.Context, user models.User) error
		wantStatus  int
		wantBody    string
	}{
		{
			name: "valid body",
			body: `{"name": "Rodney William Whitaker","email": "dasdasd"}`,
			userService: func(ctx context.Context, user models.User) error {
				return nil
			},
			wantStatus: 200,
			wantBody:   "{\"success\":true,\"message\":\"user updated\"}\n",
		},
		{
			name: "error create updated",
			body: `{"name": "Rodney William Whitaker","email": "dsfsddf"}`,
			userService: func(ctx context.Context, user models.User) error {
				return fmt.Errorf("error")
			},
			wantStatus: 400,
			wantBody:   "{\"success\":false,\"message\":\"error\"}\n",
		},
		{
			name: "invalid body",
			body: `{"name" "Rodney William Whitaker"}`,
			userService: func(ctx context.Context, user models.User) error {
				return nil
			},
			wantStatus: 400,
			wantBody:   "{\"success\":false,\"message\":\"invalid character '\\\"' after object key\"}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			r := httptest.NewRequest("POST", "/api/users/{userID}", strings.NewReader(tt.body))
			r.Header.Set("Content-Type", "application/json")

			ctx := chi.NewRouteContext()
			ctx.URLParams.Add("userID", "1")
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, ctx))

			userService := mockUserService{
				update: tt.userService,
			}
			responder := responder.NewResponder()
			ac := controller.NewUserController(userService, responder)
			ac.Update(w, r)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Equal(t, tt.wantBody, w.Body.String())
		})
	}
}
