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

package broker

import (
	"encoding/json"
	"github.com/SENERGY-Platform/developer-notifications/pkg/model"
	"log"
	"os"
	"path/filepath"
)

func LoadSubscriptionFiles(dir string) (subscriptions []model.Subscription, err error) {
	subscriptions = []model.Subscription{}
	files, err := os.ReadDir(dir)
	if err != nil {
		return subscriptions, err
	}
	for _, file := range files {
		p := filepath.Join(dir, file.Name())
		if file.IsDir() {
			temp, err := LoadSubscriptionFiles(p)
			if err != nil {
				return subscriptions, err
			}
			subscriptions = append(subscriptions, temp...)
		} else {
			ext := filepath.Ext(file.Name())
			switch ext {
			case ".md":
				//ignore and do not warn
			case ".json":
				temp, err := LoadJson(p)
				if err != nil {
					log.Println("WARNING: unable to load", p, err)
					continue
				}
				subscriptions = append(subscriptions, temp...)
			default:
				log.Println("WARNING: unknown file type in topic-descriptions directory", ext, file.Name())
			}
		}
	}
	return subscriptions, nil
}

func LoadJson(location string) (topicDescriptions []model.Subscription, err error) {
	file, err := os.Open(location)
	if err != nil {
		log.Println("error on config load:\n", location, "\n", err)
		return topicDescriptions, err
	}
	err = json.NewDecoder(file).Decode(&topicDescriptions)
	if err != nil {
		log.Println("error on config load:\n", location, "\n", err)
		return topicDescriptions, err
	}
	return topicDescriptions, nil
}
