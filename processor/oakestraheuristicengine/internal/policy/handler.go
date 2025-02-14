package policy

type PolicyHandler interface {
	Enforce() error
}
