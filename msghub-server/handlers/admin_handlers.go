package handlers

import (
	"log"
	"msghub-server/logic"
	"msghub-server/models"
	"msghub-server/template"
	"msghub-server/utils"
	"net/http"
)

func AdminLoginPageHandler(w http.ResponseWriter, r *http.Request) {
	err1 := logic.MigrateAdminDb(models.GormDb)
	if err1 != nil {
		log.Fatal("Error creating user table : ", err1.Error())
	}

	err := template.Tpl.ExecuteTemplate(w, "admin_login.gohtml", nil)
	utils.PrintError(err, "")
}

func AdminAuthenticateHandler(w http.ResponseWriter, r *http.Request) {
	//r.ParseForm()
	//
	//name := r.PostFormValue("signinName")
	//pass := r.PostFormValue("signinPass")

	//isValid, alert := .LoginAdmin(name, pass)
	//if alert != "" {
	//	alm := models.AuthErrorModel{
	//		ErrorStr: alert,
	//	}
	//	models.InitAuthErrorModel(alm)
	//}
	//
	//if isValid {
	//	http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
	//} else {
	//	http.Redirect(w, r, "/admin/login-page", http.StatusSeeOther)
	//}
}
