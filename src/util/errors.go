package util

func BailOnError(err error) {
	if err != nil {
		Stderr(err, true)
	}
}
