package utils

import "encoding/json"

// Optional is a PATCH field for a nullable column. It tells a field that was
// left out of the JSON body (keep the current value) apart from one sent as
// null (clear it), which a plain pointer cannot do.
type Optional[T any] struct {
	Set   bool // the field was present in the body
	Value *T   // nil when the field was null
}

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Set = true
	if string(data) == "null" {
		o.Value = nil
		return nil
	}

	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	o.Value = &v
	return nil
}

// Apply returns the field's new value: current if the field was left out,
// otherwise whatever the client sent.
func (o Optional[T]) Apply(current *T) *T {
	if !o.Set {
		return current
	}
	return o.Value
}
