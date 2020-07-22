package auth

// Static handler registry
var knownHandlers = map[string]Handler{
	AuthModeNone:   &NoneAuthHandler{},
	AuthModeBasic:  &BasicAuthHandler{&BaseHandler{}},
	AuthModeCustom: &CustomAuthHandler{&BaseHandler{}},
	AuthModeOAuth:  &TokenAuthHandler{&BaseHandler{}},
}

// GetAuthHandler gets the handler per the mode
func GetAuthHandler(mode string) (Handler, bool) {
	h, ok := knownHandlers[mode]

	return h, ok
}
