package xbmc

// UpNextNotify is used to ask JSONRPC from Python part to send notification to UpNext plugin
func (h *XBMCHost) UpNextNotify(payload Args) string {
	var retVal string
	h.executeJSONRPCEx("UpNext_Notify", &retVal, payload)
	return retVal
}
