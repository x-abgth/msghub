package handlers

import (
	"fmt"
	"log"
	"msghub-server/database"
	"msghub-server/models"
	"msghub-server/template"
	"msghub-server/utils"
	jwtVar "msghub-server/utils/jwt"
	"time"

	"net/http"
)

func UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	err := database.MigrateUser(database.GormDb)
	if err != nil {
		log.Fatal("Error creating user table : ", err.Error())
	}

	// CHECKING COOKIE WEATHER THE USER IS AUTHENTICATED OR NOT
	defer func() {
		// recovers panic
		if e := recover(); e != nil {
			fmt.Println("Recovered from panic : ", e)
			alm := models.ReturnAuthErrorModel()
			err1 := template.Tpl.ExecuteTemplate(w, "index.gohtml", alm)
			if err1 != nil {
				fmt.Println("Error : ", err1)
			}
		}
	}()

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occured!")
	}

	claim := jwtVar.GetValueFromJwt(c)
	if claim.IsAuthenticated {
		http.Redirect(w, r, "/user/dashboard", http.StatusFound)
	} else {
		claims := &jwtVar.UserJwtClaim{
			IsAuthenticated: false,
		}

		token := jwtVar.SignJwtToken(claims)

		http.SetCookie(w, &http.Cookie{
			Name:  "userToken",
			Value: token,
		})

		alm := models.ReturnAuthErrorModel()
		err1 := template.Tpl.ExecuteTemplate(w, "index.gohtml", alm)
		if err1 != nil {
			fmt.Println("Error : ", err1.Error())
		}
	}
}

func UserLoginCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	ph := r.PostFormValue("signinPh")
	pass := r.PostFormValue("signinPass")

	handleExceptions(w, r, "/")

	isValid, alert := database.LoginUserWithCredentials(ph, pass)
	if alert != "" {
		alm := models.AuthErrorModel{
			ErrorStr: alert,
		}
		models.InitAuthErrorModel(alm)
	}

	if isValid {
		// assigning JWT tokens
		claims := &jwtVar.UserJwtClaim{
			UserPhone:       ph,
			IsAuthenticated: true,
		}

		token := jwtVar.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		abgth := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, abgth)

		http.Redirect(w, r, "/user/dashboard", http.StatusFound)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// This handler displays the page to enter the phone number
func UserLoginWithOtpPhonePageHandler(w http.ResponseWriter, r *http.Request) {
	data := models.ReturnPhoneErrorModel()
	err := template.Tpl.ExecuteTemplate(w, "login_with_otp.gohtml", data)
	utils.PrintError(err, "")
}

// This handler process the phone number given and check weather is valid or not
func UserLoginOtpPhoneValidateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	ph := r.PostFormValue("phone")
	user := database.UserDuplicationStatus(ph)

	if user == 1 {
		status := utils.SendOtp(ph)
		if status {
			data := models.IncorrectOtpModel{
				PhoneNumber: ph,
				IsLogin:     true,
			}
			claims := &jwtVar.UserJwtClaim{
				UserPhone:       ph,
				IsAuthenticated: false,
			}

			token := jwtVar.SignJwtToken(claims)
			//
			expire := time.Now().AddDate(0, 0, 1)
			cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
			http.SetCookie(w, cookie)
			models.InitOtpErrorModel(data)
			http.Redirect(w, r, "/login/otp/getotp", http.StatusFound)
		} else {
			errorStr := models.IncorrectPhoneModel{
				ErrorStr: "Couldn't send OTP to this number!",
			}
			models.InitPhoneErrorModel(errorStr)
			http.Redirect(w, r, "/login/otp/getphone", http.StatusSeeOther)
		}
	} else {
		errorStr := models.IncorrectPhoneModel{
			ErrorStr: "The phone number you entered is not registered!",
		}
		models.InitPhoneErrorModel(errorStr)
		http.Redirect(w, r, "/login/otp/getphone", http.StatusSeeOther)
	}
}

func UserOtpPageHandler(w http.ResponseWriter, r *http.Request) {
	data := models.ReturnOtpErrorModel()
	err := template.Tpl.ExecuteTemplate(w, "user_otp_validation.gohtml", data)
	utils.PrintError(err, "")
}

func UserVerifyLoginOtpHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	otp := r.PostFormValue("loginOtp")
	ph := r.PostFormValue("loginPhone")

	status := utils.CheckOtp(ph, otp)

	if status {
		claims := &jwtVar.UserJwtClaim{
			UserPhone:       ph,
			IsAuthenticated: true,
		}

		token := jwtVar.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/user/dashboard", http.StatusFound)
	} else {
		claims := &jwtVar.UserJwtClaim{
			UserPhone:       ph,
			IsAuthenticated: false,
		}

		token := jwtVar.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)
		data := models.IncorrectOtpModel{
			ErrorStr:    "The OTP is incorrect. Try again",
			PhoneNumber: ph,
			IsLogin:     true,
		}
		models.InitOtpErrorModel(data)
		http.Redirect(w, r, "/login/otp/getotp", http.StatusFound)
	}
}

func UserRegisterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	alert := ""
	name := r.PostFormValue("signupName")
	ph := r.PostFormValue("signupPh")
	total := database.UserDuplicationStatus(ph)
	encryptedFormPassword, err := utils.HashEncrypt(r.PostFormValue("signupPass"))

	if err != nil {
		log.Fatal("Encryption error : ", err)
		alert = "The password is too weak. Please enter a strong password."
		alm := models.AuthErrorModel{
			ErrorStr: alert,
		}
		models.InitAuthErrorModel(alm)
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		if total > 0 {
			alm := models.AuthErrorModel{
				ErrorStr: "Account already exist with this number. Try Login method.",
			}
			models.InitAuthErrorModel(alm)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			status := utils.SendOtp(ph)
			if status {
				claims := &jwtVar.UserJwtClaim{
					UserPhone:       ph,
					IsAuthenticated: false,
				}

				token := jwtVar.SignJwtToken(claims)
				//
				expire := time.Now().AddDate(0, 0, 1)
				cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
				http.SetCookie(w, cookie)
				data := models.IncorrectOtpModel{
					PhoneNumber: ph,
					IsLogin:     false,
				}
				user := models.UserModel{
					UserName:  name,
					UserPhone: ph,
					UserPass:  encryptedFormPassword,
				}
				models.InitUserModel(user)
				models.InitOtpErrorModel(data)
				http.Redirect(w, r, "/register/otp/getotp", http.StatusFound)
			} else {
				alm := models.AuthErrorModel{
					ErrorStr: "Couldn't send OTP to this number. Please check the number.",
				}
				models.InitAuthErrorModel(alm)
				http.Redirect(w, r, "/", http.StatusSeeOther)
			}
		}
	}
}

func UserVerifyRegisterOtpHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	otp := r.PostFormValue("loginOtp")

	data := models.ReturnOtpErrorModel()
	status := utils.CheckOtp(data.PhoneNumber, otp)

	if status {
		user := models.ReturnUserModel()
		done, alert := database.RegisterUser(user.UserName, user.UserPhone, user.UserPass)
		if done {
			claims := &jwtVar.UserJwtClaim{
				IsAuthenticated: true,
			}

			token := jwtVar.SignJwtToken(claims)
			//
			expire := time.Now().AddDate(0, 0, 1)
			cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
			http.SetCookie(w, cookie)
			http.Redirect(w, r, "/user/dashboard", http.StatusFound)
		} else {
			alm := models.AuthErrorModel{
				ErrorStr: alert,
			}
			models.InitAuthErrorModel(alm)
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
	} else {
		data := models.IncorrectOtpModel{
			ErrorStr:    "The OTP is incorrect. Try again",
			PhoneNumber: data.PhoneNumber,
			IsLogin:     false,
		}
		models.InitOtpErrorModel(data)
		http.Redirect(w, r, "/register/otp/getotp", http.StatusSeeOther)
	}
}

func UserDashboardHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		// recovers panic
		if e := recover(); e != nil {
			fmt.Println("Recovered from panic : ", e)
			alm := models.ReturnAuthErrorModel()
			err1 := template.Tpl.ExecuteTemplate(w, "index.gohtml", alm)
			if err1 != nil {
				fmt.Println("Error : ", err1)
			}
		}
	}()

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occured!")
	}

	claim := jwtVar.GetValueFromJwt(c)

	userDetails := database.GetUserData(claim.UserPhone)
	recentChatList := database.GetRecentChatList(claim.UserPhone)
	data := models.UserDashboardModel{
		PageTitle:      "MSG-HUB",
		UserPhone:      claim.UserPhone,
		UserDetails:    userDetails,
		RecentChatList: recentChatList,
		StoryList:      nil,
	}

	err2 := template.Tpl.ExecuteTemplate(w, "user_dashboard.gohtml", data)
	utils.PrintError(err2, "")
}

func UserLogoutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Logout pressed!")
	// assigning JWT tokens
	claims := &jwtVar.UserJwtClaim{
		IsAuthenticated: false,
	}

	token := jwtVar.SignJwtToken(claims)
	//
	abgth := &http.Cookie{Name: "userToken", Value: token, MaxAge: -1, HttpOnly: true, Path: "/"}
	http.SetCookie(w, abgth)

	http.Redirect(w, r, "/", http.StatusFound)
}
