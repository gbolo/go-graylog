package graylog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
	log "github.com/sirupsen/logrus"
)

// AddUser adds a user to the MockServer.
// If ms.DataPath != "", the data is written in a file for persistence.
func (ms *MockServer) AddUser(user *User) {
	ms.Users[user.Username] = *user
	ms.safeSave()
}

// UpdateUser updates a user of the MockServer.
// If ms.DataPath != "", the data is written in a file for persistence.
func (ms *MockServer) UpdateUser(name string, user *User) {
	delete(ms.Users, name)
	ms.AddUser(user)
}

// DeleteUser removes a user from the MockServer.
// If ms.DataPath != "", the data is written in a file for persistence.
func (ms *MockServer) DeleteUser(name string) {
	delete(ms.Users, name)
	ms.safeSave()
}

// UserList returns a list of all users.
func (ms *MockServer) UserList() []User {
	if ms.Users == nil {
		return []User{}
	}
	size := len(ms.Users)
	arr := make([]User, size)
	i := 0
	for _, user := range ms.Users {
		arr[i] = user
		i++
	}
	return arr
}

func validateUser(user *User) (int, []byte) {
	if user.Username == "" {
		return 400, []byte(`{"type": "ApiError", "message": "Can not construct instance of org.graylog2.rest.models.users.responses.UserResponse, problem: Null name\n at [Source: org.glassfish.jersey.message.internal.ReaderInterceptorExecutor$UnCloseableInputStream@472db3c8; line: 1, column: 31]"}`)
	}
	return 200, []byte("")
}

// POST /users Create a new user account.
func (ms *MockServer) handleCreateUser(
	w http.ResponseWriter, r *http.Request, _ httprouter.Params,
) {
	ms.Logger.WithFields(log.Fields{
		"path": r.URL.Path, "method": r.Method,
	}).Info("request start")
	w.Header().Set("Content-Type", "application/json")
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		write500Error(w)
		return
	}
	user := &User{}
	err = json.Unmarshal(b, user)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(`{"message":"400 Bad Request"}`))
		return
	}
	sc, msg := validateUser(user)
	if sc != 200 {
		w.WriteHeader(sc)
		w.Write(msg)
		return
	}
	if _, ok := ms.Users[user.Username]; ok {
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf(
			`{"type": "ApiError", "message": "User %s already exists."}`,
			user.Username)))
		return
	}
	ms.AddUser(user)
}

// GET /users List all users
func (ms *MockServer) handleGetUsers(
	w http.ResponseWriter, r *http.Request, _ httprouter.Params,
) {
	ms.Logger.WithFields(log.Fields{
		"path": r.URL.Path, "method": r.Method,
	}).Info("request start")
	w.Header().Set("Content-Type", "application/json")
	arr := ms.UserList()
	users := usersBody{Users: arr}
	b, err := json.Marshal(&users)
	if err != nil {
		write500Error(w)
	} else {
		w.Write(b)
	}
}

// GET /users/{username} Get user details
func (ms *MockServer) handleGetUser(
	w http.ResponseWriter, r *http.Request, ps httprouter.Params,
) {
	ms.Logger.WithFields(log.Fields{
		"path": r.URL.Path, "method": r.Method,
	}).Info("request start")
	w.Header().Set("Content-Type", "application/json")
	name := ps.ByName("username")
	user, ok := ms.Users[name]
	if !ok {
		w.WriteHeader(404)
		w.Write([]byte(fmt.Sprintf(
			`{"type": "ApiError", "message": "No user found with name %s"}`, name)))
		return
	}
	b, err := json.Marshal(&user)
	if err != nil {
		write500Error(w)
	} else {
		w.Write(b)
	}
}

// PUT /users/{username} Modify user details.
func (ms *MockServer) handleUpdateUser(
	w http.ResponseWriter, r *http.Request, ps httprouter.Params,
) {
	ms.Logger.WithFields(log.Fields{
		"path": r.URL.Path, "method": r.Method,
	}).Info("request start")
	w.Header().Set("Content-Type", "application/json")
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		write500Error(w)
		return
	}
	name := ps.ByName("username")
	if _, ok := ms.Users[name]; !ok {
		w.WriteHeader(404)
		w.Write([]byte(fmt.Sprintf(
			`{"type": "ApiError", "message": "No user found with name %s"}`, name)))
		return
	}
	user := &User{}
	err = json.Unmarshal(b, user)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(`{"message":"400 Bad Request"}`))
		return
	}
	sc, msg := validateUser(user)
	if sc != 200 {
		w.WriteHeader(sc)
		w.Write(msg)
		return
	}
	ms.UpdateUser(name, user)
}

// DELETE /users/{username} Removes a user account
func (ms *MockServer) handleDeleteUser(
	w http.ResponseWriter, r *http.Request, ps httprouter.Params,
) {
	ms.Logger.WithFields(log.Fields{
		"path": r.URL.Path, "method": r.Method,
	}).Info("request start")
	w.Header().Set("Content-Type", "application/json")
	name := ps.ByName("username")
	_, ok := ms.Users[name]
	if !ok {
		w.WriteHeader(404)
		w.Write([]byte(fmt.Sprintf(
			`{"type": "ApiError", "message": "No user found with name %s"}`, name)))
		return
	}
	ms.DeleteUser(name)
}
