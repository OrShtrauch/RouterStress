package errors

type NoSSIDFound struct{}

func (n NoSSIDFound) Error() string {
	return "No SSID found"
}

type NoFilesFound struct{}

func (n NoFilesFound) Error() string {
	return "No Files Found With given file Index"
}
