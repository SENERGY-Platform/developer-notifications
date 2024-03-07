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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SENERGY-Platform/developer-notifications/pkg/configuration"
	"github.com/SENERGY-Platform/developer-notifications/pkg/model"
	"github.com/SENERGY-Platform/developer-notifications/pkg/receiver/registry"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"text/template"
)

func init() {
	registry.ReceiverFactories = append(registry.ReceiverFactories, func(ctx context.Context, wg *sync.WaitGroup, config configuration.Config) (name string, receiver registry.Receiver, err error) {
		name = "slack"
		receiver, err = New(ctx, wg, config)
		return name, receiver, err
	})
}

func New(ctx context.Context, wg *sync.WaitGroup, config configuration.Config) (*Receiver, error) {
	if config.SlackWebhookUrl == "" || config.SlackWebhookUrl == "-" {
		return nil, fmt.Errorf("%w (missing slack webhook)", registry.ErrNotConfigured)
	}
	tmpl, err := template.New("slackmsg").Parse(Template)
	if err != nil {
		return nil, err
	}
	return &Receiver{config: config, tmpl: tmpl}, nil
}

type Receiver struct {
	config configuration.Config
	tmpl   *template.Template
}

func (this *Receiver) Send(message model.Message, additionalInfo string) error {
	pl, err := this.CreatePayload(message)
	if err != nil {
		return err
	}
	return this.send(pl)
}

func (this *Receiver) send(pl string) error {
	b, err := json.Marshal(map[string]string{"text": pl})
	if err != nil {
		return err
	}
	resp, err := http.Post(this.config.SlackWebhookUrl, "application/json", bytes.NewBuffer(b))
	if err != nil {
		log.Println("ERROR: ", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		err = errors.New("unexpected status code " + strconv.Itoa(resp.StatusCode) + ": " + string(body))
		log.Println("ERROR: ", err)
		return err
	}
	return nil
}

func (this *Receiver) CreatePayload(message model.Message) (result string, err error) {
	str := strings.Builder{}
	err = this.tmpl.Execute(&str, message)
	if err != nil {
		return "", err
	}
	return str.String(), nil
}
