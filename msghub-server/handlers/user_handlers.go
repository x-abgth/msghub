package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"msghub-server/logic"
	"msghub-server/models"
	"msghub-server/template"
	"msghub-server/utils"
	jwtPkg "msghub-server/utils/jwt"
	"os"
	"time"

	"net/http"
)

// This might helps to pass error strings from one route to other
type InformationHelper struct {
	userRepo logic.UserDb
	errorStr string
	title    string
}

func (info *InformationHelper) UserLoginHandler(w http.ResponseWriter, r *http.Request) {
	err := info.userRepo.MigrateUserDb(models.GormDb)
	if err != nil {
		log.Fatal("Error creating user table : ", err.Error())
	}

	claims := &jwtPkg.UserJwtClaim{
		IsAuthenticated: false,
	}

	token := jwtPkg.SignJwtToken(claims)

	http.SetCookie(w, &http.Cookie{
		Name:     "userToken",
		Value:    token,
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})

	alm := struct {
		ErrorStr string
	}{
		ErrorStr: info.errorStr,
	}
	err1 := template.Tpl.ExecuteTemplate(w, "index.gohtml", alm)
	if err1 != nil {
		fmt.Println("Error : ", err1.Error())
	}
}

func (info *InformationHelper) UserLoginCredentialsHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	ph := r.PostFormValue("signinPh")
	pass := r.PostFormValue("signinPass")

	handleExceptions(w, r, "/")

	isValid, alert := info.userRepo.UserLoginLogic(ph, pass)

	if isValid {
		data := models.ReturnUserModel()
		// assigning JWT tokens
		claims := &jwtPkg.UserJwtClaim{
			User:            *data,
			IsAuthenticated: true,
		}

		token := jwtPkg.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/user/dashboard", http.StatusFound)
	} else {
		info.errorStr = alert.Error()
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// This handler displays the page to enter the phone number
func (info *InformationHelper) UserLoginWithOtpPhonePageHandler(w http.ResponseWriter, r *http.Request) {
	data := models.ReturnOtpErrorModel()
	err := template.Tpl.ExecuteTemplate(w, "login_with_otp.gohtml", data)
	utils.PrintError(err, "")
}

// This handler process the phone number given and check weather is valid or not
func (info *InformationHelper) UserLoginOtpPhoneValidateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	ph := r.PostFormValue("phone")

	if user := info.userRepo.UserDuplicationStatsAndSendOtpLogic(ph); user {
		userData := models.UserModel{UserPhone: ph}
		claims := &jwtPkg.UserJwtClaim{
			User:            userData,
			IsAuthenticated: false,
		}

		token := jwtPkg.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/login/otp/getotp", http.StatusFound)
	} else {
		http.Redirect(w, r, "/login/otp/getphone", http.StatusSeeOther)
	}
}

func (info *InformationHelper) UserOtpPageHandler(w http.ResponseWriter, r *http.Request) {
	data := models.ReturnOtpErrorModel()
	err := template.Tpl.ExecuteTemplate(w, "user_otp_validation.gohtml", data)
	utils.PrintError(err, "")
}

func (info *InformationHelper) UserVerifyLoginOtpHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	otp := r.PostFormValue("loginOtp")
	ph := r.PostFormValue("loginPhone")

	status := info.userRepo.UserValidateOtpLogic(ph, otp)

	if status {
		// TODO: Need more data for user
		userData := models.UserModel{UserPhone: ph}
		claims := &jwtPkg.UserJwtClaim{
			User:            userData,
			IsAuthenticated: true,
		}

		token := jwtPkg.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/user/dashboard", http.StatusFound)
	} else {
		userData := models.UserModel{UserPhone: ph}
		claims := &jwtPkg.UserJwtClaim{
			User:            userData,
			IsAuthenticated: false,
		}

		token := jwtPkg.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/login/otp/getotp", http.StatusFound)
	}
}

func (info *InformationHelper) UserRegisterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	name := r.PostFormValue("signupName")
	ph := r.PostFormValue("signupPh")
	pass := r.PostFormValue("signupPass")

	status := info.userRepo.UserRegisterLogic(name, ph, pass)
	if status {
		userData := models.UserModel{UserPhone: ph}
		claims := &jwtPkg.UserJwtClaim{
			User:            userData,
			IsAuthenticated: false,
		}

		token := jwtPkg.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)

		http.Redirect(w, r, "/register/otp/getotp", http.StatusFound)
	} else {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func (info *InformationHelper) UserVerifyRegisterOtpHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	otp := r.PostFormValue("loginOtp")

	user := models.ReturnUserModel()

	if done, flag := info.userRepo.CheckUserRegisterOtpLogic(otp, user.UserName, user.UserPhone, user.UserPass); done {
		userData := models.UserModel{
			UserName:  user.UserName,
			UserPhone: user.UserPhone,
		}
		claims := &jwtPkg.UserJwtClaim{
			User:            userData,
			IsAuthenticated: true,
		}

		token := jwtPkg.SignJwtToken(claims)
		//
		expire := time.Now().AddDate(0, 0, 1)
		cookie := &http.Cookie{Name: "userToken", Value: token, Expires: expire, HttpOnly: true, Path: "/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/user/dashboard", http.StatusFound)
	} else {
		if flag == "login" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else if flag == "otp" {
			http.Redirect(w, r, "/register/otp/getotp", http.StatusSeeOther)
		}
	}
}

func (info *InformationHelper) UserDashboardHandler(w http.ResponseWriter, r *http.Request) {
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
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	fmt.Println("In dashboard handler ", claim.User.UserPhone)
	data := info.userRepo.GetDataForDashboardLogic(claim.User.UserPhone)
	info.errorStr = ""
	err2 := template.Tpl.ExecuteTemplate(w, "user_dashboard.gohtml", data)
	utils.PrintError(err2, "")
}

func (info *InformationHelper) UserShowPeopleHandler(w http.ResponseWriter, r *http.Request) {
	err := template.Tpl.ExecuteTemplate(w, "user_show_people.gohtml", nil)
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}
}

func (info *InformationHelper) UserCreateGroup(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 24)
	if err != nil {
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}

	groupName := r.PostFormValue("groupName")
	groupAbout := r.PostFormValue("group-about")

	file, _, _ := r.FormFile("profile_photo")

	if file != nil {
		defer file.Close()

		imageName := fmt.Sprintf("%s*.png", groupName)
		out, pathError := os.CreateTemp("../msghub-client/assets/db_files/", imageName)
		if pathError != nil {
			log.Println("Error creating a file for writing", pathError)
			return
		}

		defer out.Close()

		_, copyError := io.Copy(out, file)
		if copyError != nil {
			log.Println("Error copying", copyError)
		}
	}

	var groupImage string
	if file == nil {
		groupImage = ""
	} else {
		groupImage = groupName
	}

	data := models.GroupModel{
		Image: groupImage,
		Name:  groupName,
		About: groupAbout,
	}
	claims := &jwtPkg.UserJwtClaim{
		GroupModel: data,
	}

	token := jwtPkg.SignJwtToken(claims)
	expire := time.Now().AddDate(0, 0, 1)
	cookie := &http.Cookie{Name: "userGroupDetails", Value: token, Expires: expire, HttpOnly: true, Path: "/user/dashboard/"}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/user/dashboard/add-group-members", http.StatusSeeOther)
}

func (info *InformationHelper) UserAddGroupMembers(w http.ResponseWriter, r *http.Request) {
	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	data := info.userRepo.GetAllUsersLogic(claim.User.UserPhone)

	err2 := template.Tpl.ExecuteTemplate(w, "add_group_members.gohtml", data)
	if err2 != nil {
		log.Println(err2.Error())
		cookie := &http.Cookie{Name: "userGroupDetails", MaxAge: -1, HttpOnly: true, Path: "/user/dashboard/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}
}

func (info *InformationHelper) UserGroupCreationHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Created group reached!")
	type groupMembers struct {
		Data []string `json:"data"`
	}
	var val groupMembers

	a, _ := io.ReadAll(r.Body)
	err := json.Unmarshal(a, &val)
	if err != nil {
		log.Println("ERROR happened -- ", err.Error())
		cookie := &http.Cookie{Name: "userGroupDetails", MaxAge: -1, HttpOnly: true, Path: "/user/dashboard/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}

	// INSERT GROUP DATA TO THE DATABASE
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode("success")

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (info *InformationHelper) UserLogoutHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Logout pressed!")
	// assigning JWT tokens
	claims := &jwtPkg.UserJwtClaim{
		IsAuthenticated: false,
	}

	token := jwtPkg.SignJwtToken(claims)
	//
	cookie := &http.Cookie{Name: "userToken", Value: token, MaxAge: -1, HttpOnly: true, Path: "/"}
	http.SetCookie(w, cookie)

	http.Redirect(w, r, "/", http.StatusFound)
}
