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
	"sync"
	"testing"
)

func TestReceiver_send(t *testing.T) {
	t.Skip("manual test: needs smtp config and target mail address")
	receiver, err := New(context.Background(), &sync.WaitGroup{}, configuration.Config{
		MailSmtpHost: "",
		MailSmtpPort: "",
		MailFrom:     "",
		MailPassword: "",
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = receiver.send("", "Test", "my test body")
	if err != nil {
		t.Error(err)
		return
	}
}
