package flows

const (
	OpenIdScope       = "openid"
	RefreshTokenScope = "offline_access"
	EmailScope        = "email"
)

var Scopes = []string{EmailScope, OpenIdScope, RefreshTokenScope}
