package endpoint

// User returns a User API's endpoint url.
func (ep *Endpoints) User(name string) string {
	return ep.users + "/" + name
}

// Users returns a User API's endpoint url.
func (ep *Endpoints) Users() string {
	return ep.users
}
