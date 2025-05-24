package rapidus

import (
	"github.com/fouched/toolkit/v2"
	"net/url"
)

// Validation wraps the toolkit Validation type
type Validation = toolkit.Validation

// Field wraps the toolkit Field type
type Field = toolkit.Field

// Validator creates an instance of a Validation based on form values.
// You can pass nil to this constructor to use the validation functions without an Http form
// It is essentially a wrapper for toolkit.Validator
func (r *Rapidus) Validator(data url.Values) *Validation {
	return &Validation{
		Data:   data,
		Errors: make(map[string]string),
	}
}
