// task_runner.go

package runner

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

// TaskHandler 是所有任务的统一接口
type TaskHandler interface {
	Run(ctx context.Context) (result interface{}, err error)
}

// VoidTask 适配 func() error 类型的任务（不接收 context，不推荐用于新代码）
type VoidTask struct {
	Fn func() error
}

func (vt *VoidTask) Run(_ context.Context) (interface{}, error) {
	return nil, vt.Fn()
}

// ContextualVoidTask 适配 func(context.Context) error 类型的任务（推荐使用）
type ContextualVoidTask struct {
	Fn func(context.Context) error
}

func (cvt *ContextualVoidTask) Run(ctx context.Context) (interface{}, error) {
	return nil, cvt.Fn(ctx)
}

// GenericTask 适配带参数和返回值的任务
type GenericTask struct {
	Fn      func(context.Context, interface{}) (interface{}, error)
	Payload interface{}
}

func (gt *GenericTask) Run(ctx context.Context) (interface{}, error) {
	return gt.Fn(ctx, gt.Payload)
}

// TaskPolicy 定义单个任务的执行策略
type TaskPolicy struct {
	MaxRetry    int           // 最大重试次数（0 = 不重试）
	RetryDelay  time.Duration // 重试间隔；若为0，使用 runner 默认值
	TaskTimeout time.Duration // 单任务超时；若为0，使用 runner 默认值
}

// ApplyDefaults 用 runner 的默认值填充 policy 中的零值
func (p *TaskPolicy) ApplyDefaults(defaultRetryDelay, defaultTaskTimeout time.Duration) TaskPolicy {
	policy := *p
	if policy.RetryDelay <= 0 {
		policy.RetryDelay = defaultRetryDelay
	}
	if policy.TaskTimeout < 0 {
		policy.TaskTimeout = 0 // 负值视为无超时
	} else if policy.TaskTimeout == 0 {
		policy.TaskTimeout = defaultTaskTimeout
	}
	return policy
}

// TaskItem 表示一个待执行的任务项
type TaskItem struct {
	ID      string
	Handler TaskHandler
	Policy  TaskPolicy // 执行策略
	index   int        // internal: 用于保持结果顺序
}

// NewTaskItem 创建任务项（推荐方式）
func NewTaskItem(id string, handler TaskHandler, policy TaskPolicy) TaskItem {
	return TaskItem{
		ID:      id,
		Handler: handler,
		Policy:  policy,
	}
}

// TaskResult 任务执行结果
type TaskResult struct {
	TaskID string
	Result interface{}
	Err    error
	index  int // internal
}

// 特殊错误：任务执行超时（不可重试）
var ErrTaskTimeout = errors.New("task timeout")

// TaskRunner 并发任务执行器（仅负责调度）
type TaskRunner struct {
	Concurrency   int           // 并发 worker 数量（≥1）
	RetryDelay    time.Duration // 全局默认重试间隔
	TaskTimeout   time.Duration // 全局默认任务超时（0 = 无限制）
	preserveOrder bool          // 是否保持输入顺序输出
}

// TaskRunnerOption 配置选项
type TaskRunnerOption func(*TaskRunner)

func WithConcurrency(conn int) TaskRunnerOption {
	return func(tr *TaskRunner) {
		if conn < 0 {
			conn = 1
		}
		tr.Concurrency = conn
	}
}

func WithRetryDelay(delay time.Duration) TaskRunnerOption {
	return func(tr *TaskRunner) {
		if delay < 0 {
			delay = 0
		}
		tr.RetryDelay = delay
	}
}

func WithTaskTimeout(timeout time.Duration) TaskRunnerOption {
	return func(tr *TaskRunner) {
		if timeout < 0 {
			timeout = 0
		}
		tr.TaskTimeout = timeout
	}
}

// WithPreserveOrder 保持任务结果与输入顺序一致（默认 false）
func WithPreserveOrder(preserve bool) TaskRunnerOption {
	return func(tr *TaskRunner) {
		tr.preserveOrder = preserve
	}
}

func NewTaskRunner(opts ...TaskRunnerOption) *TaskRunner {

	tr := &TaskRunner{
		Concurrency:   1,
		RetryDelay:    500 * time.Millisecond,
		TaskTimeout:   60 * time.Second,
		preserveOrder: false,
	}
	for _, opt := range opts {
		opt(tr)
	}
	return tr
}

//		tasks := []TaskItem{
//			NewTaskItem("api-call", &ContextualVoidTask{
//				Fn: func(ctx context.Context) error { /* ... */ },
//			}, TaskPolicy{
//				MaxRetry:    2,
//				TaskTimeout: 1 * time.Second,
//			}),
//	 GenericTask or ContextualVoidTask
//
// default: concurrency = 6, retryDelay 300ms ,timeout = 60s, total timeout 300s
func DefaultRunConcurrency(tasks []TaskItem) []TaskResult {
	totalTime := 300 * time.Second
	concurrency := 6
	retryDelay := 300 * time.Millisecond
	timeout := 60 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), totalTime)
	defer cancel()

	runner := NewTaskRunner(
		WithConcurrency(concurrency),
		WithRetryDelay(retryDelay),
		WithTaskTimeout(timeout),
	)

	return runner.Run(ctx, tasks)
}
func (tr *TaskRunner) Run(ctx context.Context, tasks []TaskItem) []TaskResult {
	if len(tasks) == 0 {
		return nil
	}

	// 复制任务并记录索引（用于保序）
	tasksCopy := make([]TaskItem, len(tasks))
	for i, t := range tasks {
		t.index = i
		tasksCopy[i] = t
	}

	// 限制 channel 缓冲区大小，防止内存爆炸
	const maxChanBuffer = 100
	taskBuf := len(tasksCopy)
	if taskBuf > maxChanBuffer {
		taskBuf = maxChanBuffer
	}
	taskCh := make(chan TaskItem, taskBuf)
	resultCh := make(chan TaskResult, taskBuf)

	var wg sync.WaitGroup

	for i := 0; i < tr.Concurrency; i++ {
		wg.Add(1)
		go tr.runWorker(ctx, taskCh, resultCh, &wg)
	}

	go tr.sendTasks(ctx, tasksCopy, taskCh)

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	results := tr.collectResults(resultCh)

	if tr.preserveOrder {
		sort.Slice(results, func(i, j int) bool {
			return results[i].index < results[j].index
		})
	}

	return results
}

func (tr *TaskRunner) runWorker(ctx context.Context, taskCh <-chan TaskItem, resultCh chan<- TaskResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range taskCh {
		select {
		case <-ctx.Done():
			resultCh <- TaskResult{TaskID: task.ID, Err: ctx.Err(), index: task.index}
			return
		default:
		}
		result := tr.executeTaskWithRetry(ctx, task)
		resultCh <- result
	}
}

func (tr *TaskRunner) executeTaskWithRetry(ctx context.Context, task TaskItem) TaskResult {
	policy := task.Policy.ApplyDefaults(tr.RetryDelay, tr.TaskTimeout)
	maxAttempts := policy.MaxRetry + 1
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := tr.executeOnceWithTimeout(ctx, task.Handler, policy.TaskTimeout, task.ID, attempt)
		if err == nil {
			return TaskResult{TaskID: task.ID, Result: result, Err: nil, index: task.index}
		}
		lastErr = err

		// 不可重试的错误
		if errors.Is(err, context.Canceled) ||
			errors.Is(err, context.DeadlineExceeded) ||
			errors.Is(err, ErrTaskTimeout) {
			return TaskResult{TaskID: task.ID, Result: nil, Err: err, index: task.index}
		}

		if attempt < maxAttempts {
			select {
			case <-time.After(policy.RetryDelay):
				// continue
			case <-ctx.Done():
				return TaskResult{TaskID: task.ID, Result: nil, Err: ctx.Err(), index: task.index}
			}
		}
	}

	return TaskResult{TaskID: task.ID, Result: nil, Err: lastErr, index: task.index}
}

func (tr *TaskRunner) executeOnceWithTimeout(
	ctx context.Context,
	handler TaskHandler,
	timeout time.Duration,
	taskID string,
	attempt int,
) (result interface{}, err error) {
	done := make(chan struct{})
	var res interface{}
	var e error

	go func() {
		defer func() {
			if r := recover(); r != nil {
				e = fmt.Errorf("task %s panicked on attempt %d: %v", taskID, attempt, r)
			}
			close(done)
		}()
		res, e = handler.Run(ctx)
	}()

	var timeoutCh <-chan time.Time
	if timeout > 0 {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		timeoutCh = timer.C
	}

	select {
	case <-done:
		return res, e
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timeoutCh:
		return nil, fmt.Errorf("%w: task %s (attempt %d) timed out after %v", ErrTaskTimeout, taskID, attempt, timeout)
	}
}

func (tr *TaskRunner) sendTasks(ctx context.Context, tasks []TaskItem, taskCh chan<- TaskItem) {
	defer close(taskCh)
	for _, task := range tasks {
		select {
		case taskCh <- task:
		case <-ctx.Done():
			return
		}
	}
}

func (tr *TaskRunner) collectResults(resultCh <-chan TaskResult) []TaskResult {
	var results []TaskResult
	for res := range resultCh {
		results = append(results, res)
	}
	return results
}
