package inarr

func String(needle string, haystack []string) bool {
	for _, el := range haystack {
		if el == needle {
			return true
		}
	}
	return false
}
