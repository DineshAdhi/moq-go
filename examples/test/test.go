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
				length := len(data)
				cond.L.Unlock()

				for itr < length {
					log.Printf("Consumer %d - Data %d", i, data[itr])
					itr++
				}
			}
		}(i)
	}

	itr := 0

	for {
		cond.L.Lock()
		data = append(data, itr)
		itr++

		cond.Broadcast()
		cond.L.Unlock()

		time.Sleep(time.Second)
	}

}
