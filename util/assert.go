package util

func AssertWithMsg(condition bool, message string) {
	if !condition {
		panic(message)
	}
}
func Assert(condition bool) {
	if !condition {
		panic("condition is false")
	}
}
