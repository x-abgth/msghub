package handlers

import (
	"msghub-server/database"
	"msghub-server/models"
	"msghub-server/template"
	"msghub-server/utils"
	"net/http"
)

func AdminLoginPageHandler(w http.ResponseWriter, r *http.Request) {
	err := template.Tpl.ExecuteTemplate(w, "admin_login.gohtml", nil)
	utils.PrintError(err, "")
}

func AdminAuthenticateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	name := r.PostFormValue("signinName")
	pass := r.PostFormValue("signinPass")

	isValid, alert := database.LoginAdmin(name, pass)
	if alert != "" {
		alm := models.AuthErrorModel{
			ErrorStr: alert,
		}
		models.InitAuthErrorModel(alm)
	}

	if isValid {
		http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
	} else {
		http.Redirect(w, r, "/admin/login-page", http.StatusSeeOther)
	}
}
