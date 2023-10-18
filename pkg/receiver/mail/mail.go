/*
 * Copyright 2023 InfAI (CC SES)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package slack

import (
	"context"
	"fmt"
	"github.com/SENERGY-Platform/developer-notifications/pkg/configuration"
	"github.com/SENERGY-Platform/developer-notifications/pkg/model"
	"github.com/SENERGY-Platform/developer-notifications/pkg/receiver/registry"
	"net/smtp"
	"strings"
	"sync"
	"text/template"
)

func init() {
	registry.ReceiverFactories = append(registry.ReceiverFactories, func(ctx context.Context, wg *sync.WaitGroup, config configuration.Config) (name string, receiver registry.Receiver, err error) {
		name = "mail"
		receiver, err = New(ctx, wg, config)
		return name, receiver, err
	})
}

func New(ctx context.Context, wg *sync.WaitGroup, config configuration.Config) (*Receiver, error) {
	if config.MailSmtpHost == "" || config.MailSmtpHost == "-" {
		return nil, fmt.Errorf("%w (missing mail smtp host)", registry.ErrNotConfigured)
	}
	if config.MailPassword == "" || config.MailPassword == "-" {
		return nil, fmt.Errorf("%w (missing mail smtp password)", registry.ErrNotConfigured)
	}
	if config.MailFrom == "" || config.MailFrom == "-" {
		return nil, fmt.Errorf("%w (missing mail smtp from)", registry.ErrNotConfigured)
	}
	if config.MailSmtpPort == "" || config.MailSmtpPort == "-" {
		return nil, fmt.Errorf("%w (missing mail smtp port)", registry.ErrNotConfigured)
	}
	auth := smtp.PlainAuth("", config.MailFrom, config.MailPassword, config.MailSmtpHost)

	tmpl, err := template.New("mailmsg").Parse(Template)
	if err != nil {
		return nil, err
	}
	return &Receiver{config: config, tmpl: tmpl, auth: auth}, nil
}

type Receiver struct {
	config configuration.Config
	tmpl   *template.Template
	auth   smtp.Auth
}

func (this *Receiver) Send(message model.Message, additionalInfo string) error {
	pl, err := this.CreatePayload(message)
	if err != nil {
		return err
	}
	return this.send(additionalInfo, message.Title, pl)
}

func (this *Receiver) send(to string, subject, body string) error {
	toList := []string{to}
	msg := []byte(fmt.Sprintf("To: %v\r\nSubject: %v\r\n\r\n%v", to, subject, body))
	return smtp.SendMail(this.config.MailSmtpHost+":"+this.config.MailSmtpPort, this.auth, this.config.MailFrom, toList, msg)
}

func (this *Receiver) CreatePayload(message model.Message) (result string, err error) {
	str := strings.Builder{}
	err = this.tmpl.Execute(&str, message)
	if err != nil {
		return "", err
	}
	return str.String(), nil
}
