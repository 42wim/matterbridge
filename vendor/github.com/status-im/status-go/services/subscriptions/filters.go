package subscriptions

type filter interface {
	getID() string
	getChanges() ([]interface{}, error)
	uninstall() error
}
