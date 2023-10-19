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

package model

import (
	"log"
	"slices"
	"time"
)

type Message struct {
	Sender string   `json:"sender"`
	Title  string   `json:"title"`
	Body   string   `json:"body"`
	Tags   []string `json:"tags"`
}

type Subscription struct {
	Key                        string          `json:"key"` //used to identify the subscription
	Receiver                   string          `json:"receiver"`
	DistinctTimeWindow         string          `json:"distinct_time_window"`
	DistinctTimeWindowDuration time.Duration   `json:"-"`
	Filter                     []MessageFilter `json:"filter"`                   //subscription is a match if no filter is not a match
	AdditionalReceiverInfo     string          `json:"additional_receiver_info"` //it is the receivers concern to interpret this field however it needs to
	Disabled                   bool            `json:"disabled"`
}

type MessageFilterType string

const SenderFilter MessageFilterType = "sender"
const TagFilter MessageFilterType = "tag"

var KnownTags = struct {
	Error        string
	Warning      string
	Notification string
}{
	Error:        "error",
	Warning:      "warning",
	Notification: "notification",
}

type MessageFilter struct {
	Type  MessageFilterType
	Value string
}

func (this *Subscription) Match(message Message) bool {
	for _, filter := range this.Filter {
		if !filter.Match(message) {
			return false
		}
	}
	return true
}

func (this *MessageFilter) Match(message Message) bool {
	switch this.Type {
	case SenderFilter:
		return message.Sender == this.Value
	case TagFilter:
		return slices.Contains(message.Tags, this.Value)
	default:
		log.Println("ERROR: unknown filter type")
		return false
	}
}
