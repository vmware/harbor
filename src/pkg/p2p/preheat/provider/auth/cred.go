package auth

const (
	// AuthModeNone means no auth required
	AuthModeNone = "NONE"
	// AuthModeBasic is basic mode
	AuthModeBasic = "BASIC"
	// AuthModeOAuth is OAuth mode
	AuthModeOAuth = "OAUTH"
	// AuthModeCustom is custom mode
	AuthModeCustom = "CUSTOM"
)

// Credential stores the related data for authorization.
type Credential struct {
	Mode string

	// Keep the auth data.
	// If authMode is 'BASIC', then 'username' and 'password' are stored;
	// If authMode is 'OAUTH', then 'token' is stored'
	// If authMode is 'CUSTOM', then 'header_key' with corresponding header value are stored.
	Data map[string]string
}
