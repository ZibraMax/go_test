Código para correr cosas en paralelo con un wait group

```go
package main

import (
	"strconv"
	"sync"
	"time"
)

func printTime(index int) {
	time.Sleep(10 * time.Second)
	println(time.Now().String() + " i=" + strconv.Itoa(index))
	wg.Done()
}

var wg = sync.WaitGroup{}

func main() {
	println("Multi thread")

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go printTime(i)
	}
	wg.Wait()
	println("Oh si, single thread")
}
```
