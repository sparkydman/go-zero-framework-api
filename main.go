package main

import (
	"encoding/json"
	"net/http"
	"regexp"
	"sync"
)

type userHandler struct{
	store *datastore
}

type user struct {
 ID string `json:"id"`
 Name string `json:"name"`
}

type datastore struct {
	m map[string]user
	*sync.RWMutex
}

var (
	listUsers = regexp.MustCompile(`^\/users[\/]*$`)
	getUsers = regexp.MustCompile(`^\/users\/(\d+)$`)
	createUser = regexp.MustCompile(`^\/users[\/]*$`)
)

func (h *userHandler) ListUsers(w http.ResponseWriter, r *http.Request){
	h.store.RLock()
	users := make([]user, 0, len(h.store.m))
	for _, user := range h.store.m{
		users = append(users, user)
	}
	h.store.RUnlock()

	jsonBytes, err := json.Marshal(users)
	if err != nil {
		internalServerError(w,r)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *userHandler) CreateUser(w http.ResponseWriter, r *http.Request){
	var u user
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil{
		internalServerError(w,r)
	}

	h.store.Lock()
	h.store.m[u.ID] = u
	h.store.Unlock()

	jsonBytes, err := json.Marshal(u)
	if err != nil {
		internalServerError(w,r)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *userHandler) GetUser(w http.ResponseWriter, r *http.Request){
	mathUser := getUsers.FindStringSubmatch(r.URL.Path)
	if len(mathUser) < 2{
		notFound(w,r)
		return
	}

	h.store.RLock()
	u, ok := h.store.m[mathUser[1]]
	h.store.RUnlock()

	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("User not found"))
		return
	}

		jsonBytes, err := json.Marshal(u)
	if err != nil {
		internalServerError(w,r)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)

}

func internalServerError(w http.ResponseWriter, r *http.Request)  {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("Internal Server Error"))
}

func notFound(w http.ResponseWriter, r *http.Request)  {
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Resource Not Found"))
}


func (u *userHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")

	switch{
	case r.Method == http.MethodGet && listUsers.MatchString(r.URL.Path):
		u.ListUsers(w,r)
		return 
	case r.Method == http.MethodGet && getUsers.MatchString(r.URL.Path):
		u.GetUser(w,r)
		return
	case r.Method == http.MethodPost && createUser.MatchString(r.URL.Path):
		u.CreateUser(w,r)
		return
	default:
		notFound(w,r)
		return
	}

}

func main() {
	mux := http.NewServeMux()
	    userH := &userHandler{
        store: &datastore{
            m: map[string]user{
                "1": {ID: "1", Name: "bob"},
            },
            RWMutex: &sync.RWMutex{},
        },
    }
    mux.Handle("/users", userH)
    mux.Handle("/users/", userH)

    http.ListenAndServe(":9001", mux)
}