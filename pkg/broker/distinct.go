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
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/SENERGY-Platform/developer-notifications/pkg/model"
)

func (this *Broker) IsDistinctMessage(msg model.Message, sub model.Subscription) bool {
	key := sub.Key + "_" + hash(msg)
	if this.existsInCache(key) {
		return false
	}
	this.addToCache(key, sub.DistinctTimeWindowDuration)
	return true
}

func hash(msg model.Message) string {
	hashArr := sha256.Sum256([]byte(fmt.Sprintf("%#v", msg)))
	return base64.StdEncoding.EncodeToString(hashArr[:])
}
