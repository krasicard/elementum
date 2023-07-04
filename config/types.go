package config

import "github.com/elgatito/elementum/xbmc"

type ConfigFormat string

type ConfigBundle struct {
	Info     *xbmc.AddonInfo
	Platform *xbmc.Platform
	Settings XbmcSettings
	Language string
	Region   string
}

const (
	JsonConfigFormat ConfigFormat = "json"
	YamlConfigFormat ConfigFormat = "yaml"
)

var (
	// Storages ...
	Storages = []string{
		"File",
		"Memory",
	}
)
