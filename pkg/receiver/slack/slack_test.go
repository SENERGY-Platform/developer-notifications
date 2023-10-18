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
	"github.com/SENERGY-Platform/developer-notifications/pkg/configuration"
	"github.com/SENERGY-Platform/developer-notifications/pkg/model"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestReceiver_send(t *testing.T) {
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

	receiver, err := New(context.Background(), &sync.WaitGroup{}, configuration.Config{SlackWebhookUrl: server.URL})
	if err != nil {
		t.Error(err)
		return
	}

	err = receiver.send("test-message")
	if err != nil {
		t.Error(err)
		return
	}

	expected := []string{`{"text":"test-message"}`}

	if !reflect.DeepEqual(expected, receivedMessages) {
		t.Errorf("\n%#v\n%#v\n", expected, receivedMessages)
		return
	}

}

func TestReceiver_CreatePayload(t *testing.T) {
	receiver, err := New(context.Background(), &sync.WaitGroup{}, configuration.Config{SlackWebhookUrl: "placeholder"})
	if err != nil {
		t.Error(err)
		return
	}

	sender := "github.com/SENERGY-Platform/developer-notifications"
	title := "Test Message"
	body := `Lorem ipsum dolor sit amet, consetetur sadipscing elitr, 
sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, 
sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. 

Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet. 
Lorem ipsum dolor sit amet, consetetur sadipscing elitr, 
sed diam nonumy eirmod tempor invidunt ut labore et dolore magna aliquyam erat, 
sed diam voluptua. At vero eos et accusam et justo duo dolores et ea rebum. 

Stet clita kasd gubergren, no sea takimata sanctus est Lorem ipsum dolor sit amet.`

	tags := []string{model.KnownTags.Warning, model.KnownTags.Notification}

	t.Run("title and body", func(t *testing.T) {
		pl, err := receiver.CreatePayload(model.Message{
			Sender: sender,
			Title:  title,
			Body:   body,
		})
		if err != nil {
			t.Error(err)
			return
		}

		if !strings.Contains(pl, sender) {
			t.Error("missing sender in payload")
		}

		if !strings.Contains(pl, title) {
			t.Error("missing title in payload")
		}

		if !strings.Contains(pl, body) {
			t.Error("missing body in payload")
		}
	})

	t.Run("title, body and tags", func(t *testing.T) {
		pl, err := receiver.CreatePayload(model.Message{
			Sender: sender,
			Title:  title,
			Body:   body,
			Tags:   tags,
		})
		if err != nil {
			t.Error(err)
			return
		}

		if !strings.Contains(pl, sender) {
			t.Error("missing sender in payload")
		}

		if !strings.Contains(pl, title) {
			t.Error("missing title in payload")
		}

		if !strings.Contains(pl, body) {
			t.Error("missing body in payload")
		}

		for _, tag := range tags {
			if !strings.Contains(pl, tag) {
				t.Error("missing tag in payload")
			}
		}

	})

}
