package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/fouched/rapidus/mailer"
	"github.com/fouched/rapidus/urlsigner"
	"myapp/data"
	"myapp/views"
	"net/http"
	"time"
)

func (h *Handlers) UserLoginGet(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, views.Login())
}

func (h *Handlers) UserLoginPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		h.App.ErrorLog.Println(err.Error())
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	user, err := h.Models.Users.GetByEmail(email)
	if err != nil {
		h.App.ErrorLog.Println(err.Error())
		return
	}

	matches, err := user.PasswordMatches(password)
	if err != nil {
		h.App.ErrorLog.Println(err.Error())
		return
	}

	// did user check remember me?
	if r.Form.Get("remember") == "remember" {
		randomString := h.randomString(12)
		hasher := sha256.New()
		_, err := hasher.Write([]byte(randomString))
		if err != nil {
			h.App.ErrorStatus(w, http.StatusBadRequest)
			return
		}

		sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
		rm := data.RememberToken{}
		err = rm.InsertToken(user.ID, sha)
		if err != nil {
			h.App.ErrorStatus(w, http.StatusBadRequest)
			return
		}

		// set the cookie
		expire := time.Now().Add(365 * 24 * 60 * 60 * time.Second) // a year!
		cookie := http.Cookie{
			Name:     fmt.Sprintf("_%s_remember", h.App.AppName),
			Value:    fmt.Sprintf("%d|%s", user.ID, sha),
			Path:     "/",
			Expires:  expire,
			HttpOnly: true,
			Domain:   h.App.Session.Cookie.Domain,
			MaxAge:   315350000,
			Secure:   h.App.Session.Cookie.Secure,
			SameSite: http.SameSiteStrictMode,
		}
		http.SetCookie(w, &cookie)
		// save hash in session
		h.App.Session.Put(r.Context(), "remember_token", sha)
	}

	if !matches {
		h.App.ErrorLog.Println("Invalid password")
		return
	}

	h.sessionPut(r.Context(), "userID", user.ID)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handlers) LogOut(w http.ResponseWriter, r *http.Request) {
	// delete remember token if it exits
	if h.App.Session.Exists(r.Context(), "remember_token") {
		rt := data.RememberToken{}
		_ = rt.Delete(h.App.Session.GetString(r.Context(), "remember_token"))
	}
	// delete cookie
	newCookie := http.Cookie{
		Name:     fmt.Sprintf("_%s_remember", h.App.AppName),
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-100 * time.Hour),
		HttpOnly: true,
		Domain:   h.App.Session.Cookie.Domain,
		MaxAge:   -1,
		Secure:   h.App.Session.Cookie.Secure,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, &newCookie)

	_ = h.sessionRenew(r.Context())
	h.sessionRemove(r.Context(), "userID")
	h.sessionRemove(r.Context(), "remember_token")
	_ = h.sessionDestroy(r.Context())
	_ = h.sessionRenew(r.Context())

	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}

func (h *Handlers) ForgotGet(w http.ResponseWriter, r *http.Request) {
	h.render(w, r, views.Forgot())
}

func (h *Handlers) ForgotPost(w http.ResponseWriter, r *http.Request) {
	// parse form
	err := r.ParseForm()
	if err != nil {
		h.App.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	// verify supplied email
	var u *data.User
	email := r.Form.Get("email")
	u, err = u.GetByEmail(email)
	if err != nil {
		h.App.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	// create link to password reset form
	link := fmt.Sprintf("%s/users/reset-password?email=%s", h.App.Server.URL, email)

	// sign the link
	sign := urlsigner.Signer{
		Secret: []byte(h.App.EncryptionKey),
	}
	signedLink := sign.GenerateTokenFromString(link)
	//h.App.InfoLog.Println("Signed link is: ", signedLink)

	// email msg
	var emailData struct {
		Link string
	}
	emailData.Link = signedLink

	msg := mailer.Message{
		To:       u.Email,
		From:     "admin@example.com",
		Subject:  "Password reset",
		Template: "password_reset",
		Data:     emailData,
	}

	h.App.Mail.Jobs <- msg
	res := <-h.App.Mail.Results
	if res.Error != nil {
		h.App.ErrorStatus(w, http.StatusBadRequest)
		return
	}

	// redir user
	h.App.Session.Put(r.Context(), "success", "An email has been sent to your address.")
	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}

func (h *Handlers) ResetPasswordGet(w http.ResponseWriter, r *http.Request) {
	// get url values
	email := r.URL.Query().Get("email")

	// validate url
	theURL := r.RequestURI
	testURL := fmt.Sprintf("%s%s", h.App.Server.URL, theURL)
	signer := urlsigner.Signer{
		Secret: []byte(h.App.EncryptionKey),
	}
	valid := signer.VerifyToken(testURL)
	if !valid {
		h.App.ErrorLog.Println("Invalid url")
		h.App.ErrorUnauthorized(w)
		return
	}

	// check expiry
	expired := signer.Expired(testURL, 60)
	if expired {
		h.App.ErrorLog.Println("Link expired")
		h.App.ErrorUnauthorized(w)
		return
	}

	// display form
	encryptedEmail, _ := h.encrypt(email)

	h.render(w, r, views.ResetPassword(encryptedEmail))
}

func (h *Handlers) ResetPasswordPost(w http.ResponseWriter, r *http.Request) {
	// parse form
	err := r.ParseForm()
	if err != nil {
		h.App.Error500(w)
		return
	}

	// get and decrypt email
	email, err := h.decrypt(r.Form.Get("email"))
	if err != nil {
		h.App.Error500(w)
		return
	}

	// get user
	var u data.User
	user, err := u.GetByEmail(email)
	if err != nil {
		h.App.Error500(w)
		return
	}

	// reset password
	err = user.ResetPassword(user.ID, r.Form.Get("password"))
	if err != nil {
		h.App.Error500(w)
		return
	}

	// redirect
	h.App.Session.Put(r.Context(), "success", "Your password has been reset. You can now log in.")
	http.Redirect(w, r, "/users/login", http.StatusSeeOther)
}
