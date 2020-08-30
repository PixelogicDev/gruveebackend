package doesuserdocexist

// doesUserDocExistReq includes the uid of the user we are checking
type doesUserDocExistReq struct {
	UID string `json:"uid"`
}

// doesUserDocExistResp includes a result of true or false
type doesUserDocExistResp struct {
	Result bool `json:"result"`
}
