package main

import (
	"encoding/json"
	"fmt"
	"slices"
)

const (
	maxStoreRetries = 10
)

func (p *Plugin) readArray(key string) ([]string, error) {
	raw, err := p.API.KVGet(key)
	if err != nil {
		return nil, err
	}

	var arr []string
	err2 := json.Unmarshal(raw, &arr)
	if err2 != nil {
		return nil, err2
	}
	return arr, nil
}

func (p *Plugin) addArrayValue(key string, value string) error {
	return p.addArrayValueWithRetries(key, value, maxStoreRetries)
}

func (p *Plugin) addArrayValueWithRetries(key string, value string, retries int) error {
	if retries == 0 {
		return fmt.Errorf("Failed to add value to array after 10 retries")
	}
	oldValue, err := p.API.KVGet(key)
	if err != nil {
		return err
	}
	var arr []string
	_ = json.Unmarshal(oldValue, &arr)
	if arr == nil {
		arr = []string{}
	}

	arr = append(arr, value)
	slices.Sort(arr)
	arr = slices.Compact(arr)
	newValue, err2 := json.Marshal(arr)
	if err2 != nil {
		return err2
	}

	inserted, err := p.API.KVCompareAndSet(key, oldValue, newValue)
	if inserted {
		return nil
	}
	if err != nil {
		return err
	}
	return p.addArrayValueWithRetries(key, value, retries - 1)
}

func (p *Plugin) removeArrayValue(key string, value string) error {
	return p.removeArrayValueWithRetries(key, value, maxStoreRetries)
}

func (p *Plugin) removeArrayValueWithRetries(key string, value string, retries int) error {
	if retries == 0 {
		return fmt.Errorf("Failed to remove value from array after 10 retries")
	}
	oldValue, err := p.API.KVGet(key)
	if err != nil {
		return err
	}
	var arr []string
	err2 := json.Unmarshal(oldValue, &arr)
	if err2 != nil {
		return err2
	}

	arr = slices.DeleteFunc(arr, func(s string) bool {
		return s == value
	})
	newValue, err2 := json.Marshal(arr)
	if err2 != nil {
		return err2
	}

	inserted, err := p.API.KVCompareAndSet(key, oldValue, newValue)
	if inserted {
		return nil
	}
	if err != nil {
		return err
	}
	return p.removeArrayValueWithRetries(key, value, retries - 1)
}

func (p *Plugin) getChannels(teamId string) ([]string, error) {
	key := fmt.Sprintf("channels_%s", teamId)
	return p.readArray(key)
}

func (p *Plugin) getTeams(channelId string) ([]string, error) {
	key := fmt.Sprintf("teams_%s", channelId)
	return p.readArray(key)
}

func (p *Plugin) linkTeamToChannel(channelId string, teamId string) error {
	channelsKey := fmt.Sprintf("teams_%s", channelId)
	teamsKey := fmt.Sprintf("channels_%s", teamId)
	err := p.addArrayValue(channelsKey, teamId)
	if err != nil {
		return err
	}
	return p.addArrayValue(teamsKey, channelId)
}

