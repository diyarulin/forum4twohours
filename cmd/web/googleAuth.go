package main

import (
	"encoding/json"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
	"net/http"
)

var oauth2Config = &oauth2.Config{
	ClientID:     "274551296930-b4mtjr1b9260o3pnsr1jc6mp2c9mvscf.apps.googleusercontent.com", // Замените на ваш Google Client ID
	ClientSecret: "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",                                      // Замените на ваш Google Client Secret
	RedirectURL:  "https://localhost:4000/user/googlecallback",
	Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
	Endpoint:     google.Endpoint,
}

func (app *application) googleLogin(w http.ResponseWriter, r *http.Request) {
	// Генерация URL для авторизации через Google
	authURL := oauth2Config.AuthCodeURL("", oauth2.AccessTypeOffline)
	// Редирект на страницу авторизации Google
	http.Redirect(w, r, authURL, http.StatusFound)
}

func (app *application) googleCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	token, err := oauth2Config.Exchange(r.Context(), code)
	if err != nil {
		app.serverError(w, err)
		return
	}

	client := oauth2Config.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer resp.Body.Close()

	userData := struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		app.serverError(w, err)
		return
	}

	userID, err := app.users.GetOrCreateOAuthUser(userData.Email, userData.Name, "google", userData.ID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.setSession(w, userID)
	app.flash(w, r, "Logged in with Google account!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

var githubOauth2Config = &oauth2.Config{
	ClientID:     "Ov23li5UZErJbCnv7tYX",                       // Замените на ваш GitHub Client ID
	ClientSecret: "e5aa8791ae6e3cb1989f3ccc4ab12223730acd48",   // Замените на ваш GitHub Client Secret
	RedirectURL:  "https://localhost:4000/user/githubcallback", // URL для обработки callback от GitHub
	Scopes:       []string{"read:user", "user:email"},          // Разрешения, запрашиваемые у пользователя
	Endpoint:     github.Endpoint,
}

// Обработчик для перенаправления на страницу авторизации GitHub
func (app *application) githubLogin(w http.ResponseWriter, r *http.Request) {
	// Генерация URL для авторизации через GitHub
	authURL := githubOauth2Config.AuthCodeURL("", oauth2.AccessTypeOffline)
	// Редирект на страницу авторизации GitHub
	http.Redirect(w, r, authURL, http.StatusFound)
}

// Обработчик для обработки callback от GitHub
func (app *application) githubCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	token, err := githubOauth2Config.Exchange(r.Context(), code)
	if err != nil {
		app.serverError(w, err)
		return
	}

	client := githubOauth2Config.Client(r.Context(), token)
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		app.serverError(w, err)
		return
	}
	defer resp.Body.Close()

	userData := struct {
		ID    int64  `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
		Login string `json:"login"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&userData); err != nil {
		app.serverError(w, err)
		return
	}

	if userData.Email == "" {
		emailResp, err := client.Get("https://api.github.com/user/emails")
		if err != nil {
			app.serverError(w, err)
			return
		}
		defer emailResp.Body.Close()

		emails := []struct {
			Email   string `json:"email"`
			Primary bool   `json:"primary"`
		}{}

		if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil {
			app.serverError(w, err)
			return
		}

		for _, e := range emails {
			if e.Primary {
				userData.Email = e.Email
				break
			}
		}
	}

	userID, err := app.users.GetOrCreateOAuthUser(userData.Email, userData.Name, "github", string(userData.ID))
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.setSession(w, userID)
	app.flash(w, r, "Logged in with GitHub account!")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
