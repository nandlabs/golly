package clients

type Client[RQ any, RS any] interface {
	SetOptions(options *ClientOptions)
	Execute(req RQ) (RS, error)
}
