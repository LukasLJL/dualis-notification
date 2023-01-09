package dualis

import (
	"bytes"
	"log"
	"net/http"
	"net/http/cookiejar"
	"text/template"
	"time"

	"github.com/lukasljl/dualis-notification/config"
	"gopkg.in/gomail.v2"
)

const (
	baseURL         = "https://dualis.dhbw.de/"
	loginPath       = "scripts/mgrqispi.dll?APPNAME=CampusNet&PRGNAME=EXTERNALPAGES&ARGUMENTS=-N000000000000001,-N000324,-Awelcome"
	loginScriptPath = "/scripts/mgrqispi.dll"
	userAgent       = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.108 Safari/537.36"
)

func InitDualis() {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
		Timeout: time.Duration(10 * time.Second),
	}

	dualis := Dualis{
		Client: client,
	}

	homeUrl, _ := dualis.login(config.Dualis.Username, config.Dualis.Password)
	dualis.initStructs(homeUrl)
	dualis.pollGrades()
}

func (dualis *Dualis) pollGrades() {
	log.Printf("Scheduled polling for grades (Every %v minute(s)).\n", config.Dualis.UpdateIntervalMinutes)

	for {
		time.Sleep(time.Duration(config.Dualis.UpdateIntervalMinutes) * time.Minute)
		log.Println("Polling for new grades.")
		updatedModules := dualis.updateModules()

		if len(updatedModules) > 0 {
			log.Println("New grades discovered.")
			dualis.sendNotification(&updatedModules)
		}
	}
}

func (dualis *Dualis) sendNotification(modules *[]Module) {
	log.Printf("Sending notification for %v modules.\n", len(*modules))

	tpl, err := template.ParseFiles("templates/notification.tpl")
	if err != nil {
		panic(err)
	}

	var body bytes.Buffer

	err = tpl.Execute(&body, modules)
	if err != nil {
		panic(err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", config.SMTP.SMTPUsername)
	m.SetHeader("To", config.SMTP.NotificationRecipient)
	m.SetHeader("Subject", "New grades available!")
	m.SetBody("text/html", body.String())

	d := gomail.NewDialer(config.SMTP.SMTPHost, config.SMTP.SMTPPort, config.SMTP.SMTPUsername, config.SMTP.SMTPPassword)

	if err := d.DialAndSend(m); err != nil {
		panic(err)
	}
}
