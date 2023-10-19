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

package tests

import (
	"context"
	"github.com/SENERGY-Platform/developer-notifications/pkg"
	"github.com/SENERGY-Platform/developer-notifications/pkg/broker"
	"github.com/SENERGY-Platform/developer-notifications/pkg/client"
	"github.com/SENERGY-Platform/developer-notifications/pkg/configuration"
	"github.com/SENERGY-Platform/developer-notifications/pkg/model"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestBroker(t *testing.T) {
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := configuration.Load("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}

	receivedMessages := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		msg, err := io.ReadAll(request.Body)
		if err != nil {
			t.Error(err)
			http.Error(writer, err.Error(), 500)
			return
		}
		receivedMessages = append(receivedMessages, string(msg))
		writer.WriteHeader(200)
	}))
	defer server.Close()

	config.SlackWebhookUrl = server.URL

	config.SubscriptionFilesDir = "./testdata/subscriptions"

	port, err := GetFreePort()
	if err != nil {
		t.Error(err)
		return
	}
	config.ApiPort = strconv.Itoa(port)

	err = pkg.Start(ctx, wg, config)
	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(time.Second)

	c := client.New("http://localhost:" + config.ApiPort)

	err = c.SendMessage(model.Message{
		Sender: "test",
		Title:  "test-title-1",
		Body:   "test-body",
		Tags:   nil,
	})
	if err != nil {
		t.Error(err)
		return
	}
	if len(receivedMessages) != 1 {
		t.Errorf("%#v", len(receivedMessages))
		return
	}

	err = c.SendMessage(model.Message{
		Sender: "test",
		Title:  "test-title-error-1",
		Body:   "test-error-body",
		Tags:   []string{model.KnownTags.Error},
	})
	if err != nil {
		t.Error(err)
		return
	}
	if len(receivedMessages) != 3 {
		t.Errorf("%#v", len(receivedMessages))
		return
	}

	err = c.SendMessage(model.Message{
		Sender: "test",
		Title:  "test-title-warning-1",
		Body:   "test-warning-body",
		Tags:   []string{model.KnownTags.Warning},
	})
	if err != nil {
		t.Error(err)
		return
	}
	if len(receivedMessages) != 5 {
		t.Errorf("%#v", len(receivedMessages))
		return
	}

	err = c.SendMessage(model.Message{
		Sender: "test",
		Title:  "test-title-warning-and-error-1",
		Body:   "test-warning-and-error-body",
		Tags:   []string{model.KnownTags.Warning, model.KnownTags.Error},
	})
	if err != nil {
		t.Error(err)
		return
	}
	if len(receivedMessages) != 8 {
		t.Errorf("%#v", len(receivedMessages))
		return
	}

	err = c.SendMessage(model.Message{
		Sender: "unknown",
		Title:  "unknown",
		Body:   "unknown",
		Tags:   []string{},
	})
	if err != nil {
		t.Error(err)
		return
	}
	//increase not because of 'unknown' subscription but because of 'default' slack
	if len(receivedMessages) != 9 {
		t.Errorf("%#v", len(receivedMessages))
		return
	}

}

func TestSubscriptionDirLoad(t *testing.T) {
	config, err := configuration.Load("../../config.json")
	if err != nil {
		t.Error(err)
		return
	}
	config.SubscriptionFilesDir = "./testdata/subscriptions"

	subscriptions, err := broker.LoadSubscriptions(config)
	if err != nil {
		t.Error(err)
		return
	}

	keys := map[string]bool{}
	for _, sub := range subscriptions {
		keys[sub.Key] = true
		if sub.DistinctTimeWindowDuration == 0 {
			t.Error(sub.Key, "DistinctTimeWindowDuration=", sub.DistinctTimeWindowDuration)
			return
		}
	}

	if !reflect.DeepEqual(keys, map[string]bool{
		"default":       true,
		"slack-errors":  true,
		"mail-errors":   true,
		"slack-warning": true,
		"mail-warning":  true,
		"unknown":       true,
	}) {
		t.Error(keys)
		return
	}
}

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}
