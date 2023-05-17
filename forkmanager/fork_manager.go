package forkmanager

import (
	"fmt"
	"sort"
	"sync"
)

/*
Regarding whether it is okay to use the Singleton pattern in Go, it's a topic of some debate.
The Singleton pattern can introduce global state and make testing and concurrent programming more challenging.
It can also make code tightly coupled and harder to maintain.
In general, it's recommended to favor dependency injection and explicit collaboration over singletons.

However, there might be scenarios where the Singleton pattern is still useful,
such as managing shared resources or maintaining a global configuration.
Just be mindful of the potential drawbacks and consider alternative patterns when appropriate.
*/

var (
	forkManagerInstance     *forkManager
	forkManagerInstanceLock sync.Mutex
)

type forkManager struct {
	lock sync.Mutex

	forkMap     map[ForkName]Fork
	forks       []Fork
	handlersMap map[ForkHandlerName][]ForkHandler
}

func GetInstance() *forkManager {
	forkManagerInstanceLock.Lock()
	defer forkManagerInstanceLock.Unlock()

	if forkManagerInstance == nil {
		fork := Fork{
			Name:            BaseFork,
			FromBlockNumber: 0,
		}
		forkManagerInstance = &forkManager{
			forks: []Fork{
				fork,
			},
			forkMap: map[ForkName]Fork{
				fork.Name: fork,
			},
			handlersMap: map[ForkHandlerName][]ForkHandler{},
		}
	}

	return forkManagerInstance
}

func (fm *forkManager) RegisterFork(name ForkName, fromBlock uint64) {
	fm.lock.Lock()
	defer fm.lock.Unlock()

	fork := Fork{Name: name, FromBlockNumber: fromBlock}
	fm.forkMap[name] = fork

	if len(fm.forks) == 0 {
		fm.forks = append(fm.forks, fork)
	} else {
		// keep everything in sorted order
		index := sort.Search(len(fm.forks), func(i int) bool {
			return fm.forks[i].FromBlockNumber >= fromBlock
		})
		fm.forks = append(fm.forks, Fork{})
		copy(fm.forks[index+1:], fm.forks[index:])
		fm.forks[index] = fork
	}
}

func (fm *forkManager) RegisterHandler(forkName ForkName, handlerName ForkHandlerName, handler interface{}) error {
	fm.lock.Lock()
	defer fm.lock.Unlock()

	fork, exists := fm.forkMap[forkName]
	if !exists {
		return fmt.Errorf("fork does not exist: %s", forkName)
	}

	if handlers, exists := fm.handlersMap[handlerName]; !exists {
		fm.handlersMap[handlerName] = []ForkHandler{
			{
				FromBlockNumber: fork.FromBlockNumber,
				Handler:         handler,
			},
		}
	} else {
		// keep everything in sorted order
		index := sort.Search(len(handlers), func(i int) bool {
			return handlers[i].FromBlockNumber >= fork.FromBlockNumber
		})
		handlers = append(handlers, ForkHandler{})
		copy(handlers[index+1:], handlers[index:])
		handlers[index] = ForkHandler{
			FromBlockNumber: fork.FromBlockNumber,
			Handler:         handler,
		}
		fm.handlersMap[handlerName] = handlers
	}

	return nil
}

func (fm *forkManager) GetHandler(name ForkHandlerName, blockNumber uint64) interface{} {
	fm.lock.Lock()
	defer fm.lock.Unlock()

	handlers, exists := fm.handlersMap[name]
	if !exists {
		panic(fmt.Errorf("handlers not registered for %s", name)) //nolint:gocritic
	}

	// binary search to find first position inside []*ForkHandler where FromBlockNumber >= blockNumber
	pos := sort.Search(len(handlers), func(i int) bool {
		return blockNumber < handlers[i].FromBlockNumber
	}) - 1

	return handlers[pos].Handler
}

func (fm *forkManager) IsForkSupported(name ForkName) bool {
	fm.lock.Lock()
	defer fm.lock.Unlock()

	_, exists := fm.forkMap[name]

	return exists
}

func (fm *forkManager) IsForkEnabled(name ForkName, blockNumber uint64) bool {
	fm.lock.Lock()
	defer fm.lock.Unlock()

	fork, exists := fm.forkMap[name]
	if !exists {
		return false
	}

	return fork.FromBlockNumber <= blockNumber
}

func (fm *forkManager) GetForkBlock(name ForkName) (uint64, error) {
	fm.lock.Lock()
	defer fm.lock.Unlock()

	fork, exists := fm.forkMap[name]
	if !exists {
		return 0, fmt.Errorf("fork does not exists: %s", name)
	}

	return fork.FromBlockNumber, nil
}
