package args

func Int(i int) *int {
	return &i
}

func String(s string) *string {
	return &s
}

func Bool(b bool) *bool {
	return &b
}

func BoolString(s string) *bool {
	if s == "true" {
		return Bool(true)
	} else if s == "false" {
		return Bool(false)
	}
	return nil
}
