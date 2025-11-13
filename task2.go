package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// 指针-整数指针，增加10
func addTen(p *int) {
	*p += 10
}

// 指针-整数切片，每个元素乘以2
func mul2(slice *[]int) {
	for k, v := range *slice {
		(*slice)[k] = v * 2
	}
}

// Goroutine-打印奇偶数
func printEvenOdd() {
	go func() {
		for i := 1; i < 10; i += 2 {
			fmt.Println("1-10的奇数", i)
		}
	}()

	go func() {
		for i := 2; i <= 10; i += 2 {
			fmt.Println("2-10的偶数", i)
		}
	}()
}

// Goroutine- Person 结构体
type Person struct {
	Name string
	Age  int
}

type Employee struct {
	Person
	EmployeeID int
}

func (ep *Employee) PrintInfo() {
	fmt.Printf("Employee Name: %s\nEmployee Age: %d\nEmployee ID: %d\n", ep.Name, ep.Age, ep.EmployeeID)
}

// Channel-两个协程之间的通信
func ChSend10Receive() {
	ch := make(chan int)
	go func() {
		for i := 0; i < 10; i++ {
			ch <- i
		}
		close(ch)
	}()

	go func() {
		for v := range ch {
			fmt.Println(v)
		}
		fmt.Println("receive ending")
	}()
}

// 通道的缓冲机制
func ChSend100Receive() {
	bufferedCh := make(chan int, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			bufferedCh <- i
		}
		close(bufferedCh)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for num := range bufferedCh {
			fmt.Printf("%d\n", num)
		}
	}()

	wg.Wait()
	fmt.Println("END!")
}

// 锁机制-计数器
type Counter struct {
	mu    sync.Mutex
	count int
}

func Count10000() {
	counter := &Counter{}
	var wg sync.WaitGroup

	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				counter.mu.Lock()
				counter.count += 1
				counter.mu.Unlock()
			}
		}()
	}

	wg.Wait()
	counter.mu.Lock()
	fmt.Println(counter.count)
	counter.mu.Unlock()
}

// 锁机制-计数器，原子操作，无锁
type AtomicCounter struct {
	count int64
}

func Count10000Atomic() {
	counter := &AtomicCounter{}
	var wg sync.WaitGroup

	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				// counter.count += 1
				atomic.AddInt64(&counter.count, 1)
			}
		}()
	}

	wg.Wait()
	fmt.Println(atomic.LoadInt64(&counter.count))
}

// func main() {
// 	// val := 1
// 	// addTen(&val)
// 	// fmt.Println(val)

// 	// slice := []int{1, 2, 3}
// 	// mul2(&slice)
// 	// fmt.Println(slice)

// 	// printEvenOdd()
// 	// time.Sleep(2 * time.Second)

// 	// ep := &Employee{
// 	// 	Person{"Bob", 21},
// 	// 	1001,
// 	// }
// 	// ep.PrintInfo()

// 	// ChSend10Receive()
// 	// time.Sleep(2 * time.Second)

// 	// ChSend100Receive()

// 	// Count10000()

// 	// Count10000Atomic()
// }
