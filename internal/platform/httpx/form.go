package httpx

// FormErrors holds transport-neutral validation presentation for web adapters.
type FormErrors struct {
	Global string
	Fields map[string]string
}

func (e FormErrors) Has() bool {
	return e.Global != "" || len(e.Fields) > 0
}

func (e FormErrors) Field(name string) string {
	if e.Fields == nil {
		return ""
	}
	return e.Fields[name]
}

// NewFieldErrors constructs FormErrors with field messages.
func NewFieldErrors(fields map[string]string) FormErrors {
	return FormErrors{Fields: fields}
}

// GlobalError constructs FormErrors with only a global message.
func GlobalError(msg string) FormErrors {
	return FormErrors{Global: msg}
}
