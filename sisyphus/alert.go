// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sisyphus

import (
	"fmt"
	"log"
	"net/smtp"
	"time"
)

const (
	gmailSMTPServer    = "smtp.gmail.com"
	gmailSMTPPort      = 587
	defaultTimeZoneLoc = "America/Los_Angeles"
)

// AlertConfig is optional. It customizes the alert email message.
type AlertConfig struct {
	TimeZoneLocation string
	Subject          string
	Prologue         string
	Epilogue         string
}

type alert struct {
	gmailAppPass  string
	identity      string
	senderAddr    string
	receiverAddrs []string
	alertConfig   *AlertConfig
	location      *time.Location
}

func NewAlert(gmailAppPass, identity, senderAddr, receiverAddr string,
	alertConfig *AlertConfig) (*alert, error) {
	alert := &alert{
		gmailAppPass:  gmailAppPass,
		identity:      identity,
		senderAddr:    senderAddr,
		receiverAddrs: []string{receiverAddr},
		alertConfig:   alertConfig,
	}
	timeZoneLocation := alertConfig.TimeZoneLocation
	if timeZoneLocation == "" {
		timeZoneLocation = defaultTimeZoneLoc
	}
	var err error
	alert.location, err = time.LoadLocation(timeZoneLocation)
	return alert, err
}

// Send uses gmail smtp server to send out email.
func (a *alert) Send(body string) error {
	timestamp, err := a.now()
	if err != nil {
		log.Printf("Failed to read current time when sending alert\n")
		return err
	}
	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n\n%s\n",
		a.senderAddr, a.receiverAddrs,
		a.alertConfig.Subject+timestamp,
		a.alertConfig.Prologue+body+a.alertConfig.Epilogue)
	gmailSMTPAddr := fmt.Sprintf("%s:%d", gmailSMTPServer, gmailSMTPPort)
	smtpAuth := smtp.PlainAuth(a.identity, a.senderAddr, a.gmailAppPass, gmailSMTPServer)
	err = smtp.SendMail(gmailSMTPAddr, smtpAuth, a.senderAddr, a.receiverAddrs, []byte(msg))
	if err != nil {
		log.Printf("Alert failed to be sent\n")
		return err
	}
	return nil
}

func (a *alert) now() (string, error) {
	return time.Now().In(a.location).Format("2006-01-02 15:04:05 PST"), nil
}
