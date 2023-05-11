# Implementing Collectors

Any collector must conform to the collector interface (TODO: link to collector interface). It should use the callback to expose collected information to the user.
Once you have filled out your collector. Any arguments should be added to the `CollectionConstuctor` and method should also be defined on the `CollectionConstuctor`.
You will then need to add a call to the new method in the `setupCollectors` function in the runner package.
As well as implementing your custom collector you will also need to extend `CollectionConstructor`, `setupCollectors()` and `collectorNames` to integrate it into the tool and allow the tool to use your new collector.

An example of a very simple collector:

In `collectors/collectors.go` any arguments additional should be added to the `CollectionConstuctor`
```go
...

type CollectionConstuctor struct {
    ...
    Msg string
}

...
```

In `collectors/anouncement_collector.go` you should define your collector and a constuctor method on `CollectionConstuctor`
```go
package collectors

import (
	"fmt"

	"github.com/redhat-partner-solutions/vse-sync-testsuite/pkg/callbacks"
)

type AnouncementCollector struct {
	callback callbacks.Callback
	key      string
	msg      string
}

func (anouncer *AnouncementCollector) Start(key string) error {
	anouncer.key = key
	return nil
}

func (anouncer *AnouncementCollector) ShouldPoll() bool {
	// We always want to annouce ourselves
	return true
}

func (anouncer *AnouncementCollector) Poll() []error {
	errs := make([]error, 0)
	err := anouncer.callback.Call(
		fmt.Sprintf("%T", anouncer),
		anouncer.key,
		anouncer.msg,
	)
	if err != nil {
		errs = append(errs, err)
	}
	return errs
}

func (anouncer *AnouncementCollector) CleanUp(key string) error {
	return nil
}

func (constuctor *CollectionConstuctor) NewAnouncementCollector() (*AnouncementCollector, error) {
	anouncer := AnouncementCollector{callback: constuctor.Callback, msg: constuctor.Msg}
	return &anouncer, nil
}

```
In runner/runner.go Call the `NewAnouncementCollector` constuctor in the `initialise` method of CollectorRunner and append `"Anouncer"` to `collectorNames` in the `NewCollectorRunner` function.
```go
func NewCollectorRunner() CollectorRunner {
	...
    collectorNames = append(collectorNames, "PTP", "Anouncer")
	...
}

func (runner *CollectorRunner) initialise(...){
	...
	for _, constuctorName := range runner.collectorNames {
		var newCollector collectors.Collector
		switch constuctorName {
            ...
		case "Anouncer": //nolint: goconst // This is just for ilustrative purposes
			NewAnouncerCollector, err := constuctor.NewAnouncementCollector()
			// Handle error...
            utils.IfErrorPanic(err)
			newCollector = NewAnouncerCollector
			log.Debug("Anouncer Collector")
		...
        }
        ...
    }
    ...
}

```
