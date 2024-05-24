package merror

// Append an error to a multi-error.
// If `to` is `nil` it will just assign `err`.
// If `to` is not a `*MError` it will create a new `*MError` and append both errors.
// If `err` is `nil` it will just return `to`.
// Otherwise it will just append to the existing `*MError`.
func Append(to, err error) error {
	if err == nil {
		return to
	}
	if to == nil {
		return err
	}

	if merr, ok := to.(*MError); ok {
		merr.Append(err)
		return merr
	}

	merr := New()
	merr.Append(to)
	merr.Append(err)
	return merr
}
