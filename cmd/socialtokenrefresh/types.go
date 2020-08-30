package socialtokenrefresh

// spotifyRefreshTokenRes contains the response from Spotify when trying to refresh the access token
type spotifyRefreshTokenRes struct {
	PlatformName string `json:"platformName"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
}
