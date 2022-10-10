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
	userRepo  logic.UserDb
	groupRepo logic.GroupDataLogicModel
	errorStr  string
	title     string
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

	if ok, flag := info.userRepo.CheckUserRegisterOtpLogic(otp, user.UserName, user.UserPhone, user.UserPass); ok {
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

	data, err := info.userRepo.GetDataForDashboardLogic(claim.User.UserPhone)
	if err != nil {
		info.errorStr = err.Error()
		panic(err.Error())
	}
	info.errorStr = ""
	err2 := template.Tpl.ExecuteTemplate(w, "user_dashboard.gohtml", data)
	utils.PrintError(err2, "")
}

func (info *InformationHelper) UserProfilePageHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		// recovers panic
		if e := recover(); e != nil {
			fmt.Println("Recovered from panic : ", e)
			err1 := template.Tpl.ExecuteTemplate(w, "index.gohtml", nil)
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

	// Take values from the database
	userInfo, err2 := info.userRepo.GetUserDataLogic(claim.User.UserPhone)
	if err2 != nil {
		panic(err2.Error())
	}

	data := struct {
		Name  string
		About string
		Phone string
		Image string
	}{
		Name:  userInfo.UserName,
		About: userInfo.UserAbout,
		Phone: userInfo.UserPhone,
		Image: userInfo.UserAvatarUrl,
	}

	err := template.Tpl.ExecuteTemplate(w, "user_profile_update.gohtml", data)
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}
}

func (info *InformationHelper) UserProfileUpdateHandler(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if e := recover(); e != nil {
			log.Println("ERROR HAPPENED -- ", e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
		}
	}()

	err := r.ParseMultipartForm(10 << 24)
	if err != nil {
		panic(err.Error())
	}

	userName := r.PostFormValue("name")
	userAbout := r.PostFormValue("about")

	file, _, _ := r.FormFile("user_photo")

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	var imageName, imageNameA string
	if file != nil {
		defer file.Close()

		// Check weather the string is same as in db
		imageName = fmt.Sprintf("../msghub-client/assets/db_files/user_dp_img/%s.png", claim.User.UserPhone)
		imageNameA = fmt.Sprintf("%s.png", claim.User.UserPhone)
		out, pathError := os.Create(imageName)
		if pathError != nil {
			log.Println("Error creating a file for writing", pathError)
			panic(pathError.Error())
		}
		imageName = out.Name()
		fmt.Println("Name of the file, ", imageName)

		defer out.Close()

		_, copyError := io.Copy(out, file)
		if copyError != nil {
			panic(copyError.Error())
		}
	}

	var userImage string
	if file == nil {
		userImage = ""
	} else {
		userImage = imageNameA
	}

	// Update data to the database
	err2 := info.userRepo.UpdateUserProfileDataLogic(userName, userAbout, userImage, claim.User.UserPhone)
	if err2 != nil {
		panic(err2.Error())
	}

	http.Redirect(w, r, "/user/dashboard", http.StatusFound)
}

func (info *InformationHelper) UserShowPeopleHandler(w http.ResponseWriter, r *http.Request) {
	err := template.Tpl.ExecuteTemplate(w, "user_show_people.gohtml", nil)
	if err != nil {
		fmt.Println(err.Error())
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}
}

func (info *InformationHelper) UserCreateGroup(w http.ResponseWriter, r *http.Request) {
	// Group migration statements
	groupMigrationError := info.groupRepo.MigrateGroupDb(models.GormDb)
	if groupMigrationError != nil {
		log.Fatal("Can't migrate group - ", groupMigrationError.Error())
	}

	groupUserMigrationError := info.groupRepo.MigrateUserGroupDb(models.GormDb)
	if groupUserMigrationError != nil {
		log.Fatal("Can't migrate group - ", groupUserMigrationError.Error())
	}

	// Parse form to get data
	err := r.ParseMultipartForm(10 << 24)
	if err != nil {
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}

	groupName := r.PostFormValue("groupName")
	groupAbout := r.PostFormValue("group-about")

	file, _, _ := r.FormFile("profile_photo")

	c, err1 := r.Cookie("userToken")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	claim := jwtPkg.GetValueFromJwt(c)

	var imageName, imageNameA string
	if file != nil {
		defer file.Close()

		// Check weather the string is same as in db
		imageName = fmt.Sprintf("../msghub-client/assets/db_files/user_dp_img/%s%s.png", groupName, claim.User.UserPhone)
		imageNameA = fmt.Sprintf("%s%s.png", groupName, claim.User.UserPhone)
		out, pathError := os.Create(imageName)
		if pathError != nil {
			log.Println("Error creating a file for writing", pathError)
			return
		}
		imageName = out.Name()

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
		groupImage = imageNameA
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

	defer func() {
		if e := recover(); e != nil {
			log.Println(e)
			http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
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

	data, err := info.userRepo.GetAllUsersLogic(claim.User.UserPhone)
	if err != nil {
		panic(err.Error())
	}

	err2 := template.Tpl.ExecuteTemplate(w, "add_group_members.gohtml", data)
	if err2 != nil {
		log.Println(err2.Error())
		cookie := &http.Cookie{Name: "userGroupDetails", MaxAge: -1, HttpOnly: true, Path: "/user/dashboard/"}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/user/dashboard", http.StatusSeeOther)
	}
}

func (info *InformationHelper) UserGroupCreationHandler(w http.ResponseWriter, r *http.Request) {
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

	// Claim to get group data
	gc, err1 := r.Cookie("userGroupDetails")
	if err1 != nil {
		if err1 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	groupClaim := jwtPkg.GetValueFromJwt(gc)

	// Claim to get user phone number
	uc, err2 := r.Cookie("userToken")
	if err2 != nil {
		if err2 == http.ErrNoCookie {
			panic("Cookie not found!")
		}
		panic("Unknown error occurred!")
	}

	userClaim := jwtPkg.GetValueFromJwt(uc)

	data := models.GroupModel{
		Owner:   userClaim.User.UserPhone,
		Name:    groupClaim.GroupModel.Name,
		About:   groupClaim.GroupModel.About,
		Image:   groupClaim.GroupModel.Image,
		Members: val.Data,
	}

	status, err3 := info.groupRepo.CreateGroupAndInsertDataLogic(data)

	if status {
		fmt.Println("Success - Redirect to dashboard")
		http.Redirect(w, r, "/user/dashboard", http.StatusFound)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err3.Error())
	}
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
