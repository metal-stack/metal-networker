package net

// Validator is an interface to apply common validation.
type Validator interface {
	Validate() error
}
