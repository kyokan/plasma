package node

type Service interface {
	Start() error
	Stop() error
}
