package clone

import "net/url"

// CloneObject copies the content of the given object verbatim to a new instance.
//
// If the given value is nil, nil is returned.
//
// Inner pointers are not replaced, only the top level of the object is cloned.
func CloneObject[T any](value *T) *T {
	if value == nil {
		return nil
	}
	newValue := new(T)
	*newValue = *value
	return newValue
}

// CloneURL creates a clone of the given URL object, including all of the inner pointers replaced.
func CloneURL(u *url.URL) *url.URL {
	if u == nil {
		return nil
	}
	u2 := CloneObject(u)
	u2.User = CloneObject(u2.User)
	return u2
}
