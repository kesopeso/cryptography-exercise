package assert

import "fmt"

// True panics with the given message if the condition is false.
func True(cond bool, msg string, args ...any) {
	if !cond {
		panic(fmt.Sprintf(msg, args...))
	}
}

// NoError panics with the error message if err is not nil.
func NoError(err error) {
	if err != nil {
		panic(err)
	}
}
