package usernameavailable

// usernameAvailableReq includes the username to query the user collection
type usernameAvailableReq struct {
	Username string `json:"username"`
}

// usernameAvailableResp includes a result of true or false
type usernameAvailableResp struct {
	Result bool `json:"result"`
}
