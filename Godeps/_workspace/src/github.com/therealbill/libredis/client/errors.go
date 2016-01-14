package client

type RedisError struct {
	Name    string
	Class   string
	Raw     string
	Message string
	Command string
}
