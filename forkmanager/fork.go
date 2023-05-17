package forkmanager

type ForkHandlerName string

type ForkName string

const (
	BaseFork      ForkName = "basev1"
	LondonFork    ForkName = "london"
	NikaragvaFork ForkName = "nikaragva"
)

type Fork struct {
	Name            ForkName
	FromBlockNumber uint64
}

type ForkHandler struct {
	FromBlockNumber uint64
	Handler         interface{}
}
