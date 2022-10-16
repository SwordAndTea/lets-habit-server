package service

import (
	"fmt"
	"net/smtp"
)

type MailService interface {
	Sender() string
	SendMail(toMail []string, content []byte) error
}

// mailService default implement of MailService
type mailService struct {
	identity     string
	fromMail     string
	password     string
	mailHost     string
	mailPort     uint32
	mailHostPort string
	auth         smtp.Auth
}

var defaultMailService MailService

func GetMailExecutor() MailService {
	return defaultMailService
}

// InitMailService initialize mail service, currently no error returned
func InitMailService(identity string, fromMail string, password string, mailHost string, mailPort uint32) error {
	defaultMailService = &mailService{
		identity:     identity,
		fromMail:     fromMail,
		password:     password,
		mailHost:     mailHost,
		mailPort:     mailPort,
		mailHostPort: fmt.Sprintf("%s:%d", mailHost, mailPort),
		auth:         smtp.PlainAuth(identity, fromMail, password, mailHost),
	}
	return nil
}

// SendMail send email to target mail list with content
func (m *mailService) SendMail(toMail []string, content []byte) error {
	err := smtp.SendMail(m.mailHostPort, m.auth, m.fromMail, toMail, content)
	if err != nil {
		return err
	}
	return nil
}

func (m *mailService) Sender() string {
	return m.fromMail
}
