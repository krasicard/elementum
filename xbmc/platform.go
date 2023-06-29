package xbmc

// Platform ...
type Platform struct {
	OS      string
	Arch    string
	Version string
	Kodi    int
	Build   string
}

// GetPlatform ...
func (h *XBMCHost) GetPlatform() *Platform {
	retVal := Platform{}
	h.executeJSONRPCEx("GetPlatform", &retVal, nil)
	return &retVal
}
