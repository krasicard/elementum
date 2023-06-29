package xbmc

// SetResolvedURL ...
func (h *XBMCHost) SetResolvedURL(url string) {
	retVal := -1
	h.executeJSONRPCEx("SetResolvedUrl", &retVal, Args{url})
}
