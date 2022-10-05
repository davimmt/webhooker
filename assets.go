package main

func stringInSlice(list []string, s string) bool {
	for _, i := range list {
		if i == s {
			return true
		}
	}
	return false
}
