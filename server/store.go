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

	arr := []string{}
	if len(raw) == 0 {
		return arr, nil
	}
	if err2 := json.Unmarshal(raw, &arr); err2 != nil {
		return nil, err2
	}
	return arr, nil
}

func (p *Plugin) addArrayValue(key string, value string) error {
	return p.addArrayValueWithRetries(key, value, maxStoreRetries)
}

func (p *Plugin) addArrayValueWithRetries(key string, value string, retries int) error {
	if retries == 0 {
		return fmt.Errorf("failed to add value to array after 10 retries")
	}
	oldValue, err := p.API.KVGet(key)
	if err != nil {
		return err
	}
	var arr []string
	if len(oldValue) > 0 {
		_ = json.Unmarshal(oldValue, &arr)
	}

	arr = append(arr, value)
	slices.Sort(arr)
	arr = slices.Compact(arr)
	newValue, err2 := json.Marshal(arr)
	if err2 != nil {
		return err2
	}

	inserted, err3 := p.API.KVCompareAndSet(key, oldValue, newValue)
	if inserted {
		return nil
	}
	if err3 != nil {
		return err3
	}
	return p.addArrayValueWithRetries(key, value, retries-1)
}

func (p *Plugin) removeArrayValue(key string, value string) error {
	return p.removeArrayValueWithRetries(key, value, maxStoreRetries)
}

func (p *Plugin) removeArrayValueWithRetries(key string, value string, retries int) error {
	if retries == 0 {
		return fmt.Errorf("failed to remove value from array after 10 retries")
	}
	oldValue, err := p.API.KVGet(key)
	if err != nil {
		return err
	}
	var arr []string
	if len(oldValue) == 0 {
		return nil
	}
	err2 := json.Unmarshal(oldValue, &arr)
	if err2 != nil {
		return err2
	}

	arr = slices.DeleteFunc(arr, func(s string) bool {
		return s == value
	})
	newValue, err3 := json.Marshal(arr)
	if err3 != nil {
		return err2
	}

	inserted, err4 := p.API.KVCompareAndSet(key, oldValue, newValue)
	if inserted {
		return nil
	}
	if err4 != nil {
		return err4
	}
	return p.removeArrayValueWithRetries(key, value, retries-1)
}

func (p *Plugin) getChannels(teamID string) ([]string, error) {
	key := fmt.Sprintf("channels_%s", teamID)
	return p.readArray(key)
}

func (p *Plugin) getTeams(channelID string) ([]string, error) {
	key := fmt.Sprintf("teams_%s", channelID)
	return p.readArray(key)
}

func (p *Plugin) linkTeamToChannel(channelID string, teamID string) error {
	channelsKey := fmt.Sprintf("teams_%s", channelID)
	teamsKey := fmt.Sprintf("channels_%s", teamID)
	err := p.addArrayValue(channelsKey, teamID)
	if err != nil {
		return err
	}
	return p.addArrayValue(teamsKey, channelID)
}

func (p *Plugin) unlinkTeamFromChannel(channelID string, teamID string) error {
	channelsKey := fmt.Sprintf("teams_%s", channelID)
	teamsKey := fmt.Sprintf("channels_%s", teamID)
	err := p.removeArrayValue(channelsKey, teamID)
	if err != nil {
		return err
	}
	return p.removeArrayValue(teamsKey, channelID)
}
