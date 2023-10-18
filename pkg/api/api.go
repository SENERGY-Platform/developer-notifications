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
	"context"
	"github.com/SENERGY-Platform/developer-notifications/pkg/api/util"
	"github.com/SENERGY-Platform/developer-notifications/pkg/configuration"
	"github.com/SENERGY-Platform/developer-notifications/pkg/model"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync"
)

var endpoints []func(*httprouter.Router, configuration.Config, Broker)

type Broker interface {
	Message(msg model.Message) error
}

func Start(ctx context.Context, wg *sync.WaitGroup, config configuration.Config, broker Broker) error {
	log.Println("start api")
	router := httprouter.New()
	for _, e := range endpoints {
		log.Println("add endpoints: " + runtime.FuncForPC(reflect.ValueOf(e).Pointer()).Name())
		e(router, config, broker)
	}
	log.Println("add logging and cors")
	corsHandler := util.NewCors(router)
	logger := util.NewLogger(corsHandler)
	server := &http.Server{Addr: ":" + config.ApiPort, Handler: logger}
	go func() {
		log.Println("listening on ", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			debug.PrintStack()
			log.Fatal("FATAL:", err)
		}
	}()
	wg.Add(1)
	go func() {
		<-ctx.Done()
		log.Println("api shutdown", server.Shutdown(context.Background()))
		wg.Done()
	}()
	return nil
}
