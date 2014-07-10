// Copyright 2013-2014 Bowery, Inc.
package broome

// LoginReq contains fields for request bodies that user common login info
// such as email/password.
type LoginReq struct {
	Name     string `json:"name,omitempty"` // Only some use name.
	Email    string `json:"email"`
	Password string `json:"password"`
}
