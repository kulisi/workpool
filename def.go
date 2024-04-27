package workpool

import (
	"sync"
	"time"

	"github.com/kulisi/queue"
)

type TaskHandler func() error

type WorkPool struct {
	task         chan TaskHandler // 任务传输通道
	errChan      chan error       // 错误信息传输通道
	waitingQueue *queue.Queue     // 自定义队列实例

	isQueTask int32 // 标记是否队列取出任务
	closed    int32 // 线程池关闭标识

	wg      sync.WaitGroup // 用户精准启动线程池 worker 个数
	timeout time.Duration  // 定义超时时间
}
