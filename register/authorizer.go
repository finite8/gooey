package register

// the authorizer provides a page with authorization capabilities. 

type PageAuthorizer interface {
	IsAuthorized(page Page, ctx PageContext) 
}

type AuthResult struct {
	// The reason for the auth failure. If empty, it is assumed successful
	Reason string
	// Where to redirect the request to. Leave nil if you want the standard mechanics to handle it
	Redirect interface{} 
	// Set to true to direct the user to login instead. The request will be redirected to the relevant authorizer with a page rediect on success.
	RequestAuth bool
}