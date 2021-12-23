package queue

type Maker interface {
	NewTask(uid int, payload *Payload) error
}
