package utils

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/twilio/twilio-go"
	openapi "github.com/twilio/twilio-go/rest/verify/v2"
)

var TWILIO_ACCOUNT_SID string
var TWILIO_AUTH_TOKEN string
var VERIFY_SERVICE_SID string
var client *twilio.RestClient

func getCredentials() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println(err.Error())
		os.Exit(0)
	}

	TWILIO_ACCOUNT_SID = os.Getenv("TWILIO_SID")
	TWILIO_AUTH_TOKEN = os.Getenv("TWILIO_TOKEN")
	VERIFY_SERVICE_SID = os.Getenv("TWILIO_SERVICE")
	client = twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: TWILIO_ACCOUNT_SID,
		Password: TWILIO_AUTH_TOKEN,
	})
}

func SendOtp(toPhone string) bool {
	getCredentials()
	toPhone = "+91" + toPhone
	params := &openapi.CreateVerificationParams{}
	params.SetTo(toPhone)
	params.SetChannel("sms")

	_, err := client.VerifyV2.CreateVerification(VERIFY_SERVICE_SID, params)

	if err != nil {
		fmt.Println(err.Error())
		return false
	} else {
		return true
	}
}

func CheckOtp(toPhone, code string) bool {
	getCredentials()
	toPhone = "+91" + toPhone
	params := &openapi.CreateVerificationCheckParams{}
	params.SetTo(toPhone)
	params.SetCode(code)

	resp, err := client.VerifyV2.CreateVerificationCheck(VERIFY_SERVICE_SID, params)

	if err != nil {
		fmt.Println(err.Error())
		return false
	} else if *resp.Status == "approved" {
		fmt.Println("Correct!")
		return true
	} else {
		fmt.Println("Incorrect!")
		return false
	}
}
