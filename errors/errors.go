package errors

type NoSSIDFound struct{}

func (n NoSSIDFound) Error() string {
	return "No SSID found"
}