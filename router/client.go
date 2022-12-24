package router

type Client interface {
	Run(cmd string) (string, error)
	CloseListenerSession(cmd string)
	Close()
}