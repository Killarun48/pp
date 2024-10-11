package controller

import (
	"app/internal/infrastructure/responder"
	"app/internal/models"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type UserControllerer interface {
	Create(w http.ResponseWriter, r *http.Request)
	GetByID(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
}

type UserServicer interface {
	Create(ctx context.Context, user models.User) error
	GetByID(ctx context.Context, id string) (models.User, error)
	Update(ctx context.Context, user models.User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, c models.Conditions) ([]models.User, error)
}

type UserController struct {
	userService UserServicer
	responder responder.Responder
}

func NewUserController(userService UserServicer, respond responder.Responder) UserControllerer {
	return &UserController{
		userService: userService,
		responder: respond,
	}
}

// @Summary Создание пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param name,email body RequestUser true "Данные пользователя"
// @Success 200 {object} responder.Response
// @Failure 400 {object} responder.Response
// @Router /api/users [post]
func (uc UserController) Create(w http.ResponseWriter, r *http.Request) {
	var user RequestUser

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		uc.responder.ErrorBadRequest(w, err)
		/* w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, err.Error()) */
		return
	}

	modelUser := models.User{
		Name:  user.Name,
		Email: user.Email,
	}

	err = uc.userService.Create(context.Background(), modelUser)
	if err != nil {
		uc.responder.ErrorBadRequest(w, err)
		return
	}

	uc.responder.Success(w, "user created")
	//fmt.Fprintln(w, "user created")
}

// @Summary Поиск пользователя по ID
// @Description Данный метод не возращает удаленного пользователя, используйте "Получение всех пользователей"
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} models.User
// @Failure 400,404 {object} responder.Response
// @Router /api/users/{id} [get]
func (uc UserController) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	user, err := uc.userService.GetByID(context.Background(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			uc.responder.ErrorNotFound(w, errors.New("user not found"))
			/* w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(Response{Success: false, Message: "user not found"}) */
			return
		}

		uc.responder.ErrorBadRequest(w, err)
		return
	}

	res, err := json.Marshal(user)
	if err != nil {
		/* w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintln(w, err.Error()) */
		uc.responder.ErrorBadRequest(w, err)
		return
	}

	fmt.Fprintln(w, string(res))
}

// @Summary Обновление данных пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param name,email body RequestUser true "Данные пользователя"
// @Param id path int true "ID пользователя"
// @Success 200 {object} responder.Response
// @Failure 400,404 {object} responder.Response
// @Router /api/users/{id} [post]
func (uc UserController) Update(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	userIDint, err := strconv.Atoi(userID)
	if err != nil {
		uc.responder.ErrorBadRequest(w, err)
		return
	}

	var user RequestUser

	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		uc.responder.ErrorBadRequest(w, err)
		return
	}

	modelUser := models.User{
		ID:    userIDint,
		Name:  user.Name,
		Email: user.Email,
	}

	err = uc.userService.Update(context.Background(), modelUser)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			uc.responder.ErrorNotFound(w, errors.New("user not found"))
			return
		}
		/* w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintln(w, err.Error()) */
		uc.responder.ErrorBadRequest(w, err)
		return
	}

	uc.responder.Success(w, "user updated")
	//json.NewEncoder(w).Encode(Response{Success: true, Message: "user updated"})
}

// @Summary Удаление пользователя
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "ID пользователя"
// @Success 200 {object} responder.Response
// @Failure 400 {object} responder.Response
// @Router /api/users/{id} [delete]
func (uc UserController) Delete(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	err := uc.userService.Delete(context.Background(), userID)
	if err != nil {
		/* w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintln(w, err.Error()) */
		uc.responder.ErrorBadRequest(w, err)
		return
	}

	uc.responder.Success(w, "user deleted")
	//json.NewEncoder(w).Encode(Response{Success: true, Message: "user deleted"})
	//fmt.Fprintln(w, "user deleted")
}

// @Summary Получение всех пользователей
// @Description Возвращает всех пользователей, включая "удаленных". Offset может использоваться только вместе с Limit! Для Limit ограничений нет.
// @Tags users
// @Accept json
// @Produce json
// @Param limit query int false "Параметры пагинации"
// @Param offset query int false "Параметры пагинации"
// @Success 200 {object} []models.User
// @Failure 400 {object} responder.Response
// @Router /api/users [get]
func (uc UserController) List(w http.ResponseWriter, r *http.Request) {
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	conditions := models.Conditions{
		Limit:  limit,
		Offset: offset,
	}

	users, err := uc.userService.List(context.Background(), conditions)
	if err != nil {
		/* w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintln(w, err.Error()) */
		uc.responder.ErrorBadRequest(w, err)
		return
	}

	res, err := json.Marshal(users)
	if err != nil {
		/* w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintln(w, err.Error()) */
		uc.responder.ErrorBadRequest(w, err)
		return
	}

	fmt.Fprintln(w, string(res))
}
