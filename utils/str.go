package utils

func StringToPtr(n string) *string {
	return &n
}

func StringPtrToString(n *string) string {
	if n == nil {
		return ""
	}
	return *n
}
