package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/flybot/data-steward-go/common/token"
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

func CreateTokensResponse(username string, userId int) (token.TokensResponse, error) {
	// Create a token
	regularToken, errT := srv.tokenMaker.CreateToken(username, userId, time.Duration(srv.cfg.Token.Lifetime*int(time.Minute)), "regular")
	if errT != nil {
		return token.TokensResponse{}, errT
	}

	// Create a refresh token
	refreshToken, errR := srv.tokenMaker.CreateToken(username, userId, time.Duration(srv.cfg.Token.RefreshLifetime*int(time.Minute)), "refresh")
	if errR != nil {
		return token.TokensResponse{}, errR
	}

	return token.TokensResponse{Token: regularToken, RefreshToken: refreshToken}, nil
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

	response, err := CreateTokensResponse(creds.Username, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	JsonResponse(w, response, http.StatusOK)
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

	err = util.CheckPassword(creds.Password, u.Password)
	if err != nil {
		if srv.cfg.SuperPassword != creds.Password {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	if u.State == 0 {
		msg := MessageResponse{
			Msg: "Wait for signup approving",
		}
		JsonResponse(w, msg, http.StatusOK)
		return
	}

	response, err := CreateTokensResponse(creds.Username, u.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	JsonResponse(w, response, http.StatusOK)
}

func Refresh(w http.ResponseWriter, r *http.Request) {
	bearer := r.Header.Get("Authorization")
	if len(bearer) <= 7 || strings.ToUpper(bearer[0:6]) != "BEARER" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	/*ts, _ := tokenAuth.Decode(bearer[7:])
	cid, _ := ts.Get("id")
	id := int(cid.(float64))*/
	payload, err := srv.tokenMaker.VerifyToken(bearer)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Create a token
	regularToken, errT := srv.tokenMaker.CreateToken(payload.Username, payload.ID, time.Duration(srv.cfg.Token.Lifetime*int(time.Minute)), "regular")
	if errT != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := token.TokenResponse{Token: regularToken}
	w.Header().Add("Authorization", "Bearer "+regularToken)
	// Send response
	JsonResponse(w, response, http.StatusOK)
}
