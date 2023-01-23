package ports

type Publisher interface {
	PublishState()
	CheckTransactionStatus()
}
