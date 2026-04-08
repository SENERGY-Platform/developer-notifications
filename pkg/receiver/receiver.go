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

package receiver

import (
	"context"
	"errors"
	"sync"

	"github.com/SENERGY-Platform/developer-notifications/pkg/configuration"
	_ "github.com/SENERGY-Platform/developer-notifications/pkg/receiver/mail"
	"github.com/SENERGY-Platform/developer-notifications/pkg/receiver/registry"
	_ "github.com/SENERGY-Platform/developer-notifications/pkg/receiver/slack"
)

func New(ctx context.Context, wg *sync.WaitGroup, config configuration.Config) (result *Receivers, err error) {
	result = &Receivers{reg: map[string]registry.Receiver{}}
	for _, f := range registry.ReceiverFactories {
		name, rec, err := f(ctx, wg, config)
		if errors.Is(err, registry.ErrNotConfigured) {
			config.GetLogger().Warn("skip unconfigured receiver", "name", name, "error", err)
			continue
		}
		config.GetLogger().Info("init receiver", "name", name)
		if err != nil {
			config.GetLogger().Error("unable to init receiver", "error", err)
			return nil, err
		}
		result.reg[name] = rec
	}
	return result, nil
}

type Receivers struct {
	reg map[string]registry.Receiver
}

func (this *Receivers) Get(name string) (receiver registry.Receiver, found bool) {
	receiver, found = this.reg[name]
	return receiver, found
}
