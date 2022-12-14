package consulclient

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

type ConsulClientService struct {
	config *ConsulConfig
}

func New(config *ConsulConfig) (ConsulClient, error) {
	_fullStorageName = fmt.Sprintf("/%s/%s", strings.ToLower(config.EnvironmentName), strings.ToLower(config.StorageName))

	var scheme string
	if config.Scheme == "" {
		scheme = "http"
	}

	client, err := api.NewClient(&api.Config{
		Address: config.Address,
		Scheme:  scheme,
		Token:   config.Token,
	})
	if err == nil {
		_kvClient = client.KV()
	}

	return &ConsulClientService{
		config: config,
	}, err
}

func (service *ConsulClientService) Get(key string) (any, error) {
	pair, _, err := _kvClient.Get(_fullStorageName, nil)
	if err != nil {
		return new(any), err
	}

	var results map[string]any
	if err := json.Unmarshal(pair.Value, &results); err != nil {
		return new(any), err
	}

	return results[key], nil
}

func (service *ConsulClientService) GetOrSet(key string, period time.Duration) (any, error) {
	now := time.Now().UTC()
	if container, ok := _localStorage[key]; ok {
		if now.Before(container.Expired) {
			return container.Value, nil
		}
	}

	value, err := service.Get(key)
	if err != nil {
		return value, err
	}

	if period == 0 {
		period = DEFAULT_CACHING_INTERVAL
	}

	_localStorage[key] = CacheEntry{
		Expired: now.Add(period),
		Value:   value,
	}

	return value, nil
}

var _fullStorageName string
var _kvClient *api.KV
var _localStorage map[string]CacheEntry = make(map[string]CacheEntry)

const DEFAULT_CACHING_INTERVAL = time.Minute * 5
