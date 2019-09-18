package melody

type envelope struct {
	t      int
	msg    []byte
	filter filterFunc
}

type closesession struct {
	t           int
	key         string
	value       interface{}
	keepSession *Session
}
