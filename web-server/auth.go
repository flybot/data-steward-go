package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/flybot/data-steward-go/common/util"
	"github.com/go-chi/chi/v5"
)

// Credentials is a struct to read the username and password from the request body
type Credentials struct {
	Password string `json:"password"`
	Username string `json:"username"`
}
type SignupCredentials struct {
	Email string `json:"email"`
	Credentials
}

func AuthRouter() chi.Router {
	r := chi.NewRouter()
	return r
}

func Signup(w http.ResponseWriter, r *http.Request) {
	//Is signup allowed

	if srv.cfg.Auth.SignupAllowed != true {
		msg := MessageResponse{
			Msg: "Signup disabled",
		}
		JsonResponse(w, msg, http.StatusBadRequest)
		return
	}

	u := User{}

	if srv.cfg.Auth.SignupApproving {
		u.State = 0
	} else {
		u.State = 1
	}

	var creds SignupCredentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	log.Printf("%v", err)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// check is user already exists
	if UserExists("username", creds.Username) {
		msg := MessageResponse{
			Msg: "Username is already exists",
		}
		JsonResponse(w, msg, http.StatusBadRequest)
		return
	}
	if UserExists("email", creds.Email) {
		msg := MessageResponse{
			Msg: "Email already registered",
		}
		JsonResponse(w, msg, http.StatusBadRequest)
		return
	}

	// Create new user record
	u.Username = creds.Username
	hash, _ := util.HashPassword(creds.Password)
	u.Password = hash
	u.Email = creds.Email

	userID, cerr := u.Create()
	if cerr != nil {
		msg := MessageResponse{
			Msg: "User create error",
		}
		JsonResponse(w, msg, http.StatusInternalServerError)
		return
	}

	if u.State == 0 {
		msg := MessageResponse{
			Msg: "Wait for signup approving",
		}
		JsonResponse(w, msg, http.StatusOK)
		return
	}
	// Create a token
	token, errT := createToken(userID, srv.cfg.Token.Lifetime, "regular")
	if errT != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create a refresh token
	refreshToken, errR := createToken(userID, srv.cfg.Token.RefreshLifetime, "refresh")
	if errR != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := token.TokensResponse{Token: token, RefreshToken: refreshToken}

	// Send response
	common.Notify("signup", fmt.Sprintf("Signup user %s (%s)", u.Username, u.Email))
	common.JsonResponse(w, response, http.StatusOK)
}
func Login(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	// Get the JSON body and decode into credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		// If the structure of the body is wrong, return an HTTP error
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	u := User{}
	errNotFound := u.Get("username='" + creds.Username + "'")
	if errNotFound != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	match := CheckPasswordHash(u.Password, creds.Password)
	if match != true {
		if common.Config.SuperPassword != creds.Password {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	if u.State == 0 {
		msg := common.MessageResponse{
			Msg: "Wait for signup approving",
		}
		common.JsonResponse(w, msg, http.StatusOK)
		return
	}

	// Create a token
	token, errT := createToken(u.ID, common.Config.JWT.Lifetime, "regular")
	if errT != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create a refresh token
	refreshToken, errR := createToken(u.ID, common.Config.JWT.RefreshLifetime, "refresh")
	if errR != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := TokensResponse{Token: token, RefreshToken: refreshToken}

	// Send response
	common.Notify("login", fmt.Sprintf("Logged in %s ID: %d", creds.Username, u.ID))
	common.JsonResponse(w, response, http.StatusOK)
}
func Refresh(w http.ResponseWriter, r *http.Request) {
	bearer := r.Header.Get("Authorization")
	if len(bearer) <= 7 || strings.ToUpper(bearer[0:6]) != "BEARER" {
		w.WriteHeader(http.StatusInternalServerError)
	}

	ts, _ := tokenAuth.Decode(bearer[7:])
	cid, _ := ts.Get("id")
	id := int(cid.(float64))

	// Create a token
	token, errT := createToken(id, common.Config.JWT.Lifetime, "regular")
	if errT != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := TokenResponse{Token: token}
	w.Header().Add("Authorization", "Bearer "+token)
	// Send response
	common.Notify("refresh token", fmt.Sprintf("ID %v", id))
	common.JsonResponse(w, response, http.StatusOK)
}
