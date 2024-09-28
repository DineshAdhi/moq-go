package main

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

func main() {
	mut := sync.Mutex{}
	cond := sync.NewCond(&mut)

	var data []int

	for i := 0; i < 2; i++ {

		go func(i int) {
			itr := 0
			for {
				cond.L.Lock()
				cond.Wait()

				for itr < len(data) {
					log.Printf("Consumer %d - Data %d", i, data[itr])
					itr++
				}

				cond.L.Unlock()
			}
		}(i)
	}

	itr := 0

	for {
		data = append(data, itr)
		itr++

		if itr%5 == 0 {
			cond.Broadcast()
		}

		time.Sleep(time.Second)
	}

}
