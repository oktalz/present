package handlers

import (
	"fmt"
	"net/http"

	"github.com/oklog/ulid/v2"
	"github.com/oktalz/present/hash"
)

// cookieIDValue retrieves or sets the value of the "present-id" cookie from the request.
//
// It returns the value of the cookie as a string.
func cookieIDValue(w http.ResponseWriter, r *http.Request) string {
	cookie, err := r.Cookie("present-id")
	if err == nil {
		// Cookie exists, you can access its value using cookie.Value
		fmt.Println("present-id", cookie.Value)
		return cookie.Value
	}
	cookieID := http.Cookie{
		Name:  "present-id",
		Value: ulid.Make().String(),
		Path:  "/",
	}
	http.SetCookie(w, &cookieID)
	return cookieID.Value
}

// cookieAuth checks if the provided user password and admin password match the cookie value in the request.
//
// Parameters:
//   - userPwd: the user password to compare with the cookie value.
//   - adminPwd: the admin password to compare with the cookie value.
//   - r: the HTTP request containing the cookie.
//
// Returns:
//   - user: a boolean indicating if the user has user rights.
//   - admin: a boolean indicating if the user has admin rights.
func cookieAuth(userPwd, adminPwd string, r *http.Request) (user bool, admin bool) { //nolint:nonamedreturns
	cookie, err := r.Cookie("present")
	if err != nil {
		return false, false
	}

	passwordOKUser := hash.Equal(cookie.Value, userPwd)
	passwordOKAdmin := hash.Equal(cookie.Value, adminPwd)
	if passwordOKAdmin {
		passwordOKUser = true
	}
	return passwordOKUser, passwordOKAdmin
}

// cookieAdminAuth checks if the provided user password and admin password match the cookie value in the request.
//
// Parameters:
//   - adminPwd: the admin password to compare with the cookie value.
//   - r: the HTTP request containing the cookie.
//
// Returns:
//   - admin: a boolean indicating if the user has admin rights.
func cookieAdminAuth(adminPwd string, r *http.Request) (admin bool) { //nolint:nonamedreturns
	cookie, err := r.Cookie("present")
	if err != nil {
		return false
	}

	return hash.Equal(cookie.Value, adminPwd)
}
