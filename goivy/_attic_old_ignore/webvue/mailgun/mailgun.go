package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

func main() {

	apiKey := os.Getenv("MAILGUN_FENCEBUNT_SIGNUP_API_KEY")
	if apiKey == "" {
		apiKey = "API_KEY"
	}

	id, err := SendSimpleMessage("mg.fencebunt.com", apiKey)
	fmt.Println(id)
	fmt.Println(err)

}

func SendSimpleMessage(domain, apiKey string) (string, error) {
	mg := mailgun.NewMailgun(domain, apiKey)
	//When you have an EU-domain, you must specify the endpoint:
	// mg.SetAPIBase("https://api.eu.mailgun.net")
	m := mg.NewMessage(
		"Mailgun Sandbox <postmaster@mg.fencebunt.com>",
		"Hello Jason Aten",
		"Congratulations Jason Aten, you just sent an email with Mailgun! You are truly awesome!",
		"Jason Aten <j.e.aten@gmail.com>",
	)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	_, id, err := mg.Send(ctx, m)
	return id, err
}
