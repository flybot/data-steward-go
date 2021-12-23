package queue

type SQSClient struct {
	client int
}

func InitService() Maker {
	cl := 1
	return &SQSClient{client: cl}
}

func (service *SQSClient) NewTask(uid int, payload *Payload) error {
	//
	return nil
} 
