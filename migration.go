package main

import (
	"time"

	"github.com/elgatito/elementum/repository"
	"github.com/elgatito/elementum/xbmc"
)

func checkRepository(xbmcHost *xbmc.XBMCHost) bool {
	if xbmcHost == nil {
		return false
	}

	if xbmcHost.IsAddonInstalled("repository.elementum") {
		if !xbmcHost.IsAddonEnabled("repository.elementum") {
			xbmcHost.SetAddonEnabled("repository.elementum", true)
		}
		return true
	}

	log.Info("Creating Elementum repository add-on...")
	if err := repository.MakeElementumRepositoryAddon(); err != nil {
		log.Errorf("Unable to create repository add-on: %s", err)
		return false
	}

	xbmcHost.UpdateLocalAddons()
	for _, addon := range xbmcHost.GetAddons("xbmc.addon.repository", "unknown", "all", []string{"name", "version", "enabled"}).Addons {
		if addon.ID == "repository.elementum" && addon.Enabled == true {
			log.Info("Found enabled Elementum repository add-on")
			return false
		}
	}
	log.Info("Elementum repository not installed, installing...")
	xbmcHost.InstallAddon("repository.elementum")
	for timeout := 0; timeout < 10; timeout++ {
		if xbmcHost.IsAddonInstalled("repository.elementum") {
			break
		}
		log.Info("Sleeping 1 second while waiting for Elementum repository add-on to be installed")
		time.Sleep(1 * time.Second)
	}
	xbmcHost.SetAddonEnabled("repository.elementum", true)
	xbmcHost.UpdateLocalAddons()
	xbmcHost.UpdateAddonRepos()

	return true
}
