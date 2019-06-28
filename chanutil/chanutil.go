package chanutil

import "sync"

// Merge TODO
func Merge(input ...chan int) chan int {

	out := make(chan int)
	var shutdownSignal sync.WaitGroup

	for _, inputChan := range input {

		shutdownSignal.Add(1)
		go func(inputChan chan int) {
			defer shutdownSignal.Done()
			for num := range inputChan {
				out <- num
			}
		}(inputChan)
	}

	/*
		A goroutine will be responsible for monitoring when all upstream
		channels. When all the upstream channels are closed, it'll signal
		the downstream channel
	*/
	go func() {
		shutdownSignal.Wait()
		close(out)
	}()

	return out
}

// Split TODO
func Split(input chan int, outputs ...chan int) {

	go func() {

		/*
			Create n WaitGroup, each corresponding to goroutines spawned for
			a single output channel
		*/
		syncGoroutines := make([]sync.WaitGroup, len(outputs))

		for index, outputChan := range outputs {
			syncGoroutines[index] = sync.WaitGroup{}
			syncGoroutines[index].Add(1)

			/*
				The following goroutine monitors if all the goroutines spawned to publish
				to the respective output channel have closed using 'wg' and the 'input'
				channel has been closed, to close the 'output' channel
			*/
			go func(outputChan chan int, wg *sync.WaitGroup) {
				wg.Wait()
				close(outputChan)
			}(outputChan, &syncGoroutines[index])

		}

		for inputVal := range input {
			for index, outputChan := range outputs {

				syncGoroutines[index].Add(1)

				go func(inputVal int, outputChan chan int, wg *sync.WaitGroup) {
					outputChan <- inputVal
					wg.Done()
				}(inputVal, outputChan, &syncGoroutines[index])

			}
		}

		/*
			Since the earlier for-range loop is blocking and will only release
			once the 'input' channel is closed. We can define code here to be
			executed once the 'input' channel is closed
		*/
		for index := range syncGoroutines {
			syncGoroutines[index].Done()
		}

	}()
}

// Pipeline TODO
func Pipeline(input chan int, filters ...func(task int) bool) chan int {

	out := make(chan int)

	filter := filters[0]

	go func(out chan int, filter func(task int) bool) {
		defer close(out)
		for value := range input {
			if filter(value) {
				out <- value
			}
		}
	}(out, filter)

	if len(filters) != 1 {
		filters = filters[1:]
		return Pipeline(out, filters...)
	}

	return out
}