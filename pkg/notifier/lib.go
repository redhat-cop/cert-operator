package notifier

type Chat interface {
	Kind() string
	Send(message string) error
}
