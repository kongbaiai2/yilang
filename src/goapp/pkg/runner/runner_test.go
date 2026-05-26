// main_test.go

package runner

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"
)

func init() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lshortfile)
}

// 模拟一个可能失败的无参任务
func doSomething(name string) error {
	log.Printf("[%s] 开始执行...\n", name)
	// time.Sleep(200 * time.Millisecond)
	if name == "task-fail" {

		log.Println("开始睡眠 ...", name)
		time.Sleep(2 * time.Second)
		log.Println("结束睡眠 ...", name)
		return fmt.Errorf("模拟失败: %s", name)
	}
	log.Printf("[%s] 执行成功!\n", name)
	return nil
}

// 模拟一个带参数的任务
func fetchResource(ctx context.Context, payload interface{}) (interface{}, error) {
	id := payload.(int)
	log.Printf("[资源-%d] 正在获取...\n", id)
	// time.Sleep(300 * time.Millisecond)
	if id == 999 {
		time.Sleep(1 * time.Second)
		return nil, fmt.Errorf("资源 %d 不可用", id)
	}
	return fmt.Sprintf("data_of_%d", id), nil
}

func TestMain(t *testing.T) {
	// tasks := []TaskItem{

	// 	NewTaskItem("api-call", &ContextualVoidTask{
	// 		Fn: func(ctx context.Context) error { /* ... */ return nil },
	// 	}, TaskPolicy{
	// 		MaxRetry:    2,
	// 		TaskTimeout: 1 * time.Second,
	// 	}),

	// 	NewTaskItem("db-query", &GenericTask{
	// 		Fn:      fetchResource,
	// 		Payload: "user_123",
	// 	}, TaskPolicy{}),

	// 	{
	// 		ID: "t4",
	// 		Handler: &GenericTask{
	// 			Fn:      fetchResource,
	// 			Payload: 999, // 会失败
	// 		},
	// 		Policy: TaskPolicy{},
	// 	},
	// }
	tasks := []TaskItem{
		{
			ID: "t1",
			Handler: &VoidTask{
				Fn: func() error {

					return doSomething("task-ok")
				},
			},
		},
		{
			ID: "t2",
			Handler: &VoidTask{
				Fn: func() error { return doSomething("task-fail") },
			},
			Policy: TaskPolicy{
				MaxRetry:    3,
				TaskTimeout: 3 * time.Second,
			},
		},
		{
			ID: "t3",
			Handler: &GenericTask{
				Fn:      fetchResource,
				Payload: 101,
			},
			Policy: TaskPolicy{},
		},
		{
			ID: "t4",
			Handler: &GenericTask{
				Fn:      fetchResource,
				Payload: 999, // 会失败
			},
			Policy: TaskPolicy{
				MaxRetry:    3,
				TaskTimeout: 3 * time.Second,
			},
		},
	}

	// 整体超时：5 秒
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	runner := NewTaskRunner(WithConcurrency(4),
		WithRetryDelay(300*time.Millisecond),
		WithTaskTimeout(3*time.Second), // 单任务最多 2 秒
	)

	log.Println("🚀 开始并发执行任务...")
	results := runner.Run(ctx, tasks)

	log.Println("📊 执行结果:")
	for _, r := range results {
		if r.Err != nil {
			log.Printf("❌ [%s] 失败: %v\n", r.TaskID, r.Err)
		} else {
			log.Printf("✅ [%s] 成功: %v\n", r.TaskID, r.Result)
		}
	}

	log.Println("\n🔚 测试结束")
}
