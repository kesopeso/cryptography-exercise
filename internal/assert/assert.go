package assert

// True panics with the given message if the condition is false.
func True(cond bool, msg string) {
	if !cond {
		panic(msg)
	}
}
