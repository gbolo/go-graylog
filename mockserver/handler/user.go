package handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
	"github.com/suzuki-shunsuke/go-graylog"
	"github.com/suzuki-shunsuke/go-graylog/mockserver/logic"
)

// POST /users Create a new user account.
func HandleCreateUser(
	ms *logic.Server,
	w http.ResponseWriter, r *http.Request, _ httprouter.Params,
) (int, interface{}, error) {
	requiredFields := []string{
		"username", "email", "permissions", "full_name", "password"}
	allowedFields := []string{
		"startpage", "timezone", "session_timeout_ms", "roles"}
	body, sc, err := validateRequestBody(r.Body, requiredFields, allowedFields, nil)
	if err != nil {
		return sc, nil, err
	}

	user := &graylog.User{}
	if err := msDecode(body, user); err != nil {
		ms.Logger().WithFields(log.Fields{
			"body": body, "error": err,
		}).Info("Failed to parse request body as User")
		return 400, nil, err
	}

	if sc, err := ms.AddUser(user); err != nil {
		return sc, nil, err
	}
	ms.SafeSave()
	return 201, nil, nil
}

// GET /users List all users
func HandleGetUsers(
	ms *logic.Server,
	w http.ResponseWriter, r *http.Request, _ httprouter.Params,
) (int, interface{}, error) {
	users, sc, err := ms.GetUsers()
	if err != nil {
		return sc, users, err
	}
	return sc, &graylog.UsersBody{Users: users}, nil
}

// GET /users/{username} Get user details
func HandleGetUser(
	ms *logic.Server,
	w http.ResponseWriter, r *http.Request, ps httprouter.Params,
) (int, interface{}, error) {
	name := ps.ByName("username")
	user, sc, err := ms.GetUser(name)
	return sc, user, err
}

// PUT /users/{username} Modify user details.
func HandleUpdateUser(
	ms *logic.Server,
	w http.ResponseWriter, r *http.Request, ps httprouter.Params,
) (int, interface{}, error) {
	// required fields is nil
	acceptedFields := []string{
		"email", "permissions", "full_name", "password"}
	body, sc, err := validateRequestBody(r.Body, nil, nil, acceptedFields)
	if err != nil {
		return sc, nil, err
	}

	user := &graylog.User{Username: ps.ByName("username")}
	if err := msDecode(body, &user); err != nil {
		ms.Logger().WithFields(log.Fields{
			"body": body, "error": err,
		}).Info("Failed to parse request body as User")
		return 400, nil, err
	}

	if sc, err := ms.UpdateUser(user); err != nil {
		return sc, nil, err
	}
	ms.SafeSave()
	return 200, nil, nil
}

// DELETE /users/{username} Removes a user account
func HandleDeleteUser(
	ms *logic.Server,
	w http.ResponseWriter, r *http.Request, ps httprouter.Params,
) (int, interface{}, error) {
	name := ps.ByName("username")
	if sc, err := ms.DeleteUser(name); err != nil {
		return sc, nil, err
	}
	ms.SafeSave()
	return 204, nil, nil
}
