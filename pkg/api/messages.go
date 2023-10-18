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

package api

import (
	"encoding/json"
	"github.com/SENERGY-Platform/developer-notifications/pkg/configuration"
	"github.com/SENERGY-Platform/developer-notifications/pkg/model"
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func init() {
	endpoints = append(endpoints, MessagesEndpoint)
}

func MessagesEndpoint(router *httprouter.Router, _ configuration.Config, broker Broker) {
	router.POST("/messages", func(writer http.ResponseWriter, request *http.Request, params httprouter.Params) {
		msg := model.Message{}
		err := json.NewDecoder(request.Body).Decode(&msg)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		err = broker.Message(msg)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
}
