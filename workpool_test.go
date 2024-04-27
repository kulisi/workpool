package workpool

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	wp := New(10)
	for i := 0; i < 100000; i++ {
		ii := i
		wp.Do(func() error {

			fmt.Println(ii)
			return nil
		})
	}
	wp.Wait()
	fmt.Println("down")
}

// 支持最大任务数, 放到工作池里面 并等待全部完成
func TestWorkerPoolMax(t *testing.T) {
	// 设置最大线程数
	wp := New(10)
	// 开启20个请求
	for i := 0; i < 20; i++ {
		ii := i
		wp.Do(func() error {
			for j := 0; j < 10; j++ {
				fmt.Println(fmt.Sprintf("%v->\t%v", ii, j))
				time.Sleep(1 * time.Second)
			}
			return nil
		})
	}
	wp.Wait()
	fmt.Println("down")
}

// 支持错误返回

func TestRetError(t *testing.T) {
	wp := New(10)
	for i := 0; i < 20; i++ {
		ii := i
		wp.Do(func() error {
			for j := 0; j < 10; j++ {
				fmt.Println(fmt.Sprintf("%v->\t%v", ii, j))
				if ii == 1 {
					return errors.New("my test err")
				}
				time.Sleep(1 * time.Second)
			}
			return nil
		})
	}
	err := wp.Wait()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("down")
}

// 支持超时退出

func TestTimeout(t *testing.T) {
	wp := New(5)
	wp.SetTimeout(time.Millisecond)
	for i := 0; i < 10; i++ {
		ii := i
		wp.DoWait(func() error {
			for j := 0; j < 5; j++ {
				fmt.Println(fmt.Sprintf("%v->\t%v", ii, j))
				time.Sleep(1 * time.Second)
			}

			return nil
		})
	}
	err := wp.Wait()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("down")
}
