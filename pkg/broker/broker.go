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
	"context"
	"errors"
	"github.com/SENERGY-Platform/developer-notifications/pkg/configuration"
	"github.com/SENERGY-Platform/developer-notifications/pkg/model"
	"github.com/SENERGY-Platform/developer-notifications/pkg/receiver"
	"github.com/patrickmn/go-cache"
	"log"
	"sync"
	"time"
)

func New(ctx context.Context, wg *sync.WaitGroup, config configuration.Config) (broker *Broker, err error) {
	receivers, err := receiver.New(ctx, wg, config)
	if err != nil {
		return nil, err
	}
	c := cache.New(5*time.Minute, 1*time.Minute)

	subscriptions, err := LoadSubscriptions(config)
	if err != nil {
		return nil, err
	}

	broker = &Broker{
		config:    config,
		receivers: receivers,
		cache:     c,
	}

	for _, sub := range subscriptions {
		if _, ok := receivers.Get(sub.Receiver); !ok {
			log.Printf("ignore subscription %v to unknown or unconfigured receiver %v", sub.Key, sub.Receiver)
		} else {
			broker.subscriptions = append(broker.subscriptions, sub)
		}
	}

	return broker, nil
}

type Broker struct {
	config        configuration.Config
	receivers     *receiver.Receivers
	subscriptions []model.Subscription
	cache         *cache.Cache
}

func (this *Broker) Message(msg model.Message) error {
	wg := sync.WaitGroup{}
	mux := sync.Mutex{}
	errorList := []error{}
	matches := []string{}
	distinct := []string{}
	for _, sub := range this.subscriptions {
		if sub.Match(msg) {
			matches = append(matches, sub.Key)
			if this.IsDistinctMessage(msg, sub) {
				distinct = append(distinct, sub.Key)
				wg.Add(1)
				go func(message model.Message, subscription model.Subscription) {
					defer wg.Done()
					err := this.send(message, subscription)
					if err != nil {
						mux.Lock()
						defer mux.Unlock()
						errorList = append(errorList, err)
					}
				}(msg, sub)
			}
		}
	}
	wg.Wait()
	err := errors.Join(errorList...)
	if this.config.Debug {
		log.Printf("Message(sender=%v,title=%v,tags=%v) result: matches=%#v distinct=%#v error=%v\n", msg.Sender, msg.Title, msg.Tags, matches, distinct, err)
	}
	return err
}

func (this *Broker) send(message model.Message, subscription model.Subscription) error {
	rec, found := this.receivers.Get(subscription.Receiver)
	if !found {
		return errors.New("unknown or unconfigured receiver (" + subscription.Receiver + ")")
	}
	if this.config.Debug {
		log.Printf("send message to receiver %v\n", subscription.Receiver)
	}
	return rec.Send(message, subscription.AdditionalReceiverInfo)
}

func LoadSubscriptions(config configuration.Config) (subscriptions []model.Subscription, err error) {
	subscriptions, err = AddSubscriptions(subscriptions, config.Subscriptions)
	if err != nil {
		return nil, err
	}
	if config.SubscriptionFilesDir != "" && config.SubscriptionFilesDir != "-" {
		subs, err := LoadSubscriptionFiles(config.SubscriptionFilesDir)
		if err != nil {
			return nil, err
		}
		subscriptions, err = AddSubscriptions(subscriptions, subs)
		if err != nil {
			return nil, err
		}
	}
	return subscriptions, err
}

func AddSubscriptions(list, added []model.Subscription) (result []model.Subscription, err error) {
	result = append(result, list...)
	for _, sub := range added {
		if !sub.Disabled {
			sub.DistinctTimeWindowDuration, err = time.ParseDuration(sub.DistinctTimeWindow)
			if err != nil {
				return nil, err
			}
			result = append(result, sub)
		}
	}
	return result, nil
}
