package util

// ConvertCamelToDash converts a camelCase string to dash-case
func ConvertCamelToDash(s string) string {
	buf := make([]byte, 0, len(s)+len(s)/2)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			if i > 0 && s[i-1] >= 'a' && s[i-1] <= 'z' {
				buf = append(buf, '-')
			}
			buf = append(buf, c+('a'-'A'))
		} else {
			buf = append(buf, c)
		}
	}
	return string(buf)
}
