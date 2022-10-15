package consulclient

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
	envManager "github.com/punk-link/environment-variable-manager"
)

type ConsulClient struct {
}

func New(config ConsulConfig) (*ConsulClient, error) {
	_fullStorageName = getFullStorageName(config.StorageName)

	client, err := api.NewClient(&api.Config{
		Address: config.Address,
		Scheme:  config.Scheme,
		Token:   config.Token,
	})
	if err == nil {
		_kvClient = client.KV()
	}

	return &ConsulClient{}, err
}

func (service *ConsulClient) Get(key string) (any, error) {
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

func (service *ConsulClient) GetOrSet(key string, period time.Duration) (any, error) {
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

func getEnvironmentName() string {
	isExist, name := envManager.TryGetEnvironmentVariable("GO_ENVIRONMENT")
	if !isExist {
		return "Development"
	}

	return name
}

func getFullStorageName(storageName string) string {
	name := getEnvironmentName()
	lowerCasedName := strings.ToLower(name)
	lowerCasedStorageName := strings.ToLower(storageName)

	return fmt.Sprintf("/%s/%s", lowerCasedName, lowerCasedStorageName)
}

var _fullStorageName string
var _kvClient *api.KV
var _localStorage map[string]CacheEntry = make(map[string]CacheEntry)

const DEFAULT_CACHING_INTERVAL = time.Minute * 5
