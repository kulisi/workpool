package workpool

import (
	"context"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/kulisi/queue"
)

func New(max int) *WorkPool {
	if max < 1 {
		// 最小的最大工作线程数为 1
		max = 1
	}
	wp := &WorkPool{
		task:         make(chan TaskHandler, 2*max),
		errChan:      make(chan error, 1),
		waitingQueue: queue.New(),
	}

	// 启动队列
	go func() {
		wp.isQueTask = 1
		for {
			// 从队列中取任务
			tmp := wp.waitingQueue.Pop()
			// 如果工作池关闭，对应关闭队列，并退出循环
			if wp.IsClose() {
				wp.waitingQueue.Close()
				break
			}

			// 如果取出来的值为nil则退出循环，否则发送任务 TaskHandler
			if tmp != nil {
				task := tmp.(TaskHandler)
				if task != nil {
					wp.task <- task
				}
			} else {
				break
			}
		}
		atomic.StoreInt32(&wp.isQueTask, 0)
	}()

	// 启动 max 个 worker
	wp.wg.Add(max)
	for i := 0; i < max; i++ {
		go func() {
			defer wp.wg.Done()

			for wt := range wp.task {
				// 有 err 立即返回
				if wt == nil || atomic.LoadInt32(&wp.closed) == 1 {
					continue
				}

				closed := make(chan struct{}, 1)

				// 如果设置了超时
				if wp.timeout > 0 {
					ct, cancel := context.WithTimeout(context.Background(), wp.timeout)
					go func() {
						select {
						case <-ct.Done():
							wp.errChan <- ct.Err()
							atomic.StoreInt32(&wp.closed, 1)
							cancel()
						case <-closed:
						}
					}()
				}

				err := wt()

				close(closed)

				if err != nil {
					select {
					case wp.errChan <- err:
						atomic.StoreInt32(&wp.closed, 1)
					default:
					}
				}
			}
		}()
	}

	return wp
}

// SetTimeout 设置超时时间
func (wp *WorkPool) SetTimeout(timeout time.Duration) {
	wp.timeout = timeout
}

// Do 添加到工作池，并立即返回
func (wp *WorkPool) Do(task TaskHandler) {
	if wp.IsClose() {
		// 若工作池已关闭，立即返回
		return
	}
	// 将任务添加到队列
	wp.waitingQueue.Push(task)
}

// DoWait 添加到工作池，并等待执行完成之后再返回
func (wp *WorkPool) DoWait(task TaskHandler) {
	if wp.IsClose() {
		// 若工作池已关闭，立即返回
		return
	}
	// 用于接收是否完成的通道
	doneChan := make(chan struct{})
	wp.waitingQueue.Push(TaskHandler(func() error {
		defer close(doneChan)
		return task()
	}))
	<-doneChan
}

// 等待工作线程执行结束
func (wp *WorkPool) Wait() error {
	var c int
	// 让队列等待消费完成
	wp.waitingQueue.Wait()
	wp.waitingQueue.Close()
	for {
		c++
		if wp.IsDone() {
			if atomic.LoadInt32(&wp.isQueTask) == 0 {
				break
			}
		}

		if c <= 100 {
			runtime.Gosched()
			continue
		}
		time.Sleep(20 * time.Microsecond)
	}
	close(wp.task)
	wp.wg.Wait()
	select {
	case err := <-wp.errChan:
		return err
	default:
		return nil
	}
}

// IsDone 判断是否完成 (非阻塞)
func (wp *WorkPool) IsDone() bool {
	if wp == nil || wp.task == nil {
		return true
	}
	return wp.waitingQueue.Len() == 0 && len(wp.task) == 0
}

// IsClose 返回是否已经关闭
func (wp *WorkPool) IsClose() bool {
	return atomic.LoadInt32(&wp.closed) == 1
}
