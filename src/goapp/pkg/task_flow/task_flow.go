package task

import (
	"encoding/json"
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/alibabacloud-go/tea/tea"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Tasker interface {
	GetTask() *Task
	PreDo(Tasker) error
	Do(Tasker) error
	PostDo(Tasker) error
	Schedule() (bool, error)     // 多父任务使用
	GetContext() (string, error) // 获取当前任务详情
}
type TaskErr struct {
	TaskId string
	Exec   string
	Err    string
}
type TaskStatus string

var (
	Gdb *gorm.DB
)

const (
	StatusInit    TaskStatus = "init"
	StatusRunning TaskStatus = "running"
	StatusSuccess TaskStatus = "success"
	StatusFailed  TaskStatus = "failed"
	StatusRetry   TaskStatus = "retry"
	StatusPending TaskStatus = "pending" // call outward api failed
	StatusSkip    TaskStatus = "skip"
	GetTask                  = "GetTask"
	PreDo                    = "PreDo"
	Do                       = "Do"
	PostDo                   = "PostDo"
	IfErr                    = "IfErr"
	GetContext               = "GetContext"
	Schedule                 = "Schedule"
)

// 任务结构体（包含持久化字段）
type Task struct {
	ID         uint              `gorm:"primaryKey;autoIncrement:true"`
	Action     string            `gorm:"column:action;type:varchar(50)"`
	TaskId     string            `gorm:"column:task_id;type:varchar(50);uniqueIndex:task_id"`
	SubTask    []Tasker          `gorm:"-"`
	Context    string            `json:"-" gorm:"column:context;type:text;comment:'context'"` // json format
	Status     TaskStatus        `gorm:"column:status;type:varchar(20)"`
	Retry      int               `gorm:"column:retry;type:int"`
	MaxRetry   int               `gorm:"column:max_retry;type:int"`
	ErrMessage string            `gorm:"column:err_message;type:text;comment:'err_message'"`
	CreatedAt  time.Time         `json:"-" gorm:"autoCreateTime"`
	UpdatedAt  time.Time         `json:"-" gorm:"autoUpdateTime"`
	GlobalMap  map[string]string `json:"-" gorm:"-"`
	GlobalMaps string            `gorm:"column:global_maps;type:text;comment:'global_maps'"` // json format
}

// key传空返回map。传非空赋值后，返回map
func (t *Task) OptGlobalMap(parent Tasker, key, value string) map[string]string {
	// log.Printf("t.Action: %v, parent.Action: %v, key: %v", t.Action, parent.GetTask().Action, key)
	if t.GlobalMap == nil {
		t.GlobalMap = make(map[string]string)
	}

	if len(t.GlobalMap) < len(parent.GetTask().GlobalMap) {
		t.GlobalMap = parent.GetTask().GlobalMap
		t.GetGlobalMap()
	}
	if key != "" {
		t.GlobalMap[key] = value
		t.GetGlobalMap()
	}
	if t.TaskId != parent.GetTask().TaskId && t.GlobalMap[key] != parent.GetTask().GlobalMap[key] {
		parent.GetTask().OptGlobalMap(t, key, value)
	}

	return t.GlobalMap
}

func PrintDepth(t Task, depth int, format string) {
	msg := fmt.Sprintf("task:%v, action: %v, status: %v, %v", t.TaskId, t.Action, t.Status, format)
	log.Output(depth, msg)
}
func PrintStack() {
	var buf [4096]byte
	n := runtime.Stack(buf[:], false)
	log.Printf("==> %s", string(buf[:n]))

}
func (t *Task) TableName() (name string) {
	name = "vfw_task"
	return
}

func (t *Task) Register(sub Tasker) {
	if sub == nil {
		return
	}
	if t.TaskId == "" {
		t.GetTask().TaskId = uuid.New().String()
		t.GetTask().Status = StatusInit
	}

	if sub.GetTask().TaskId == "" {
		sub.GetTask().TaskId = uuid.New().String()
		sub.GetTask().Status = StatusInit
		sub.GetContext()
	}

	t.SubTask = append(t.SubTask, sub)
	data, _ := json.Marshal(t.SubTask)

	t.Context = string(data)
}

func (t *Task) GetTask() *Task {
	return t
}
func (t *Task) PreDo(sub Tasker) error {
	if t.GlobalMap == nil {
		t.GlobalMap = make(map[string]string)
	}
	if sub.GetTask().Status == StatusSkip {
		return nil
	}
	return sub.GetTask().IfErr(sub, PreDo, true, StatusFailed, sub.PreDo(t))
}
func (t *Task) Do(sub Tasker) error {
	if sub.GetTask().Status == StatusSkip {
		return nil
	}
	return sub.GetTask().IfErr(sub, Do, true, StatusPending, sub.Do(t))
}

func (t *Task) PostDo(sub Tasker) error {
	if sub.GetTask().Status == StatusSkip {
		return nil
	}
	return sub.GetTask().IfErr(sub, PostDo, true, StatusFailed, sub.PostDo(t))
}

// return false is ignore
func (t *Task) Schedule() (isDo bool, err error) {
	PrintDepth(*t, 2, "")
	text, _ := t.GetContext()
	t.UpdateStatus(text, StatusRunning)

	PrintDepth(*t, 2, fmt.Sprintf("task len:%v", len(t.SubTask)))
	for _, sub := range t.SubTask {
		if sub.GetTask().Status != StatusInit && sub.GetTask().Status != StatusRetry {
			continue
		}
		sub.GetTask().GlobalMap = t.GlobalMap
		isDo, err = sub.Schedule()
		if t.IfErr(sub, Schedule, true, StatusFailed, err) != nil {
			for k, v := range sub.GetTask().GlobalMap {
				t.GetTask().OptGlobalMap(t, k, v)
			}
			return
		}

		if isDo {
			for k, v := range sub.GetTask().GlobalMap {
				t.GetTask().OptGlobalMap(t, k, v)
			}
			continue
		}

		err = func() error {
			var iserr TaskErr
			exec := PreDo
			if sub.GetTask().ErrMessage != "" && sub.GetTask().Status == StatusRetry {
				err := json.Unmarshal([]byte(sub.GetTask().ErrMessage), &iserr)
				if err != nil {
					return err
				}

				if iserr.Exec != "" {
					exec = iserr.Exec
				}
			}

			PrintDepth(*sub.GetTask(), 2, "")
			for sub.GetTask().Retry = 0; sub.GetTask().Retry <= sub.GetTask().MaxRetry; sub.GetTask().Retry++ {

				if sub.GetTask().Status != StatusRetry {
					text, _ := sub.GetContext()
					sub.GetTask().UpdateStatus(text, StatusRunning)
				}
				PrintDepth(*sub.GetTask(), 2, fmt.Sprintf("retry: %d", sub.GetTask().Retry))

				if exec == PreDo {
					err = t.PreDo(sub)
					if t.IfErr(sub, PreDo, false, StatusFailed, err) != nil {
						return err
					}
					exec = Do
				}

				if exec == Do {
					err = t.Do(sub)
					if t.IfErr(sub, Do, false, StatusPending, err) != nil {
						return err
					}
					exec = PostDo
				}

				if exec == PostDo {
					err = t.PostDo(sub)
					if t.IfErr(sub, PostDo, false, StatusFailed, err) != nil {
						return err
					}
				}
				time.Sleep(30 * time.Second)
			}
			if sub.GetTask().Status == StatusSkip {
				sub.GetTask().UpdateStatus(text, StatusSkip)
				return nil
			}
			sub.GetTask().ErrMessage = ""
			text, _ := sub.GetContext()
			sub.GetTask().UpdateStatus(text, StatusSuccess)
			return nil
		}()
		if err != nil {
			return
		}

		PrintDepth(*sub.GetTask(), 2, string(StatusSuccess))
	}

	t.ErrMessage = ""
	text, _ = t.GetContext()
	t.UpdateStatus(text, StatusSuccess)

	return true, nil
}
func (t *Task) setStatusToRetry() {
	if t.Status == StatusPending || t.Status == StatusFailed {
		t.Status = StatusRetry
	}
	if len(t.SubTask) <= 0 {
		return
	}
	for _, sub := range t.SubTask {
		if sub.GetTask().Status == StatusPending || sub.GetTask().Status == StatusFailed {
			sub.GetTask().Status = StatusRetry
		}
		sub.GetTask().setStatusToRetry()
	}
}
func (t *Task) RetryDo() error {
	t.setStatusToRetry()

	_, err := t.Schedule()
	if err != nil {
		return err
	}
	return nil
}
func (t *Task) IfErr(sub Tasker, exec string, print bool, status TaskStatus, err error) error {
	if err != nil {
		e := &TaskErr{TaskId: sub.GetTask().TaskId, Exec: exec, Err: err.Error()}
		eStr := tea.Prettify(e)
		t.ErrMessage = eStr
		if t.Status != StatusPending && t.Status != StatusSkip {
			text, _ := t.GetContext()
			t.UpdateStatus(text, status)
		}

		if print {
			PrintDepth(*t, 3, fmt.Sprintf("[ERROR]: %v", t.ErrMessage))
			// PrintStack()
		}
		return err
	}
	return nil
}
func (t *Task) UpdateStatus(context string, status TaskStatus) {
	t.GetTask().Status = status

	t.GetTask().Context = context
	// log.Print(tea.Prettify(context))
	err := t.GetTask().WriteDb(*t.GetTask())
	if err != nil {
		PrintDepth(*t, 2, fmt.Sprintf("[ERROR]: %v", t.ErrMessage))
		return
	}

}
func (t *Task) WriteDb(v Task) error {
	return CreateNfvTask(Gdb, v)
	// return t.SaveDb(v)
}
func (t *Task) GetContext() (string, error) {
	jsonBytes, _ := json.Marshal(t.SubTask)
	return string(jsonBytes), nil
}

func (t *Task) GetGlobalMap() (string, error) {
	jsonBytes, _ := json.Marshal(t.GlobalMap)
	t.GlobalMaps = string(jsonBytes)
	return t.GlobalMaps, nil
}
func (t *Task) SaveDb(v Task) error {
	return Gdb.Save(&v).Where("TaskId = ?", v.GetTask().TaskId).Error
}
func CreateNfvTask(db *gorm.DB, t Task) error {
	return db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: t.TaskId}, {Name: "ID"}}, // 冲突字段
		DoUpdates: clause.AssignmentColumns([]string{"status", "updated_at",
			"context", "retry", "err_message", "global_maps"}), // 更新字段
	}).Create(&t).Error
}

func UpdateVfwTaskStatus(db *gorm.DB, taskId string, status TaskStatus) error {
	if err := db.Model(&Task{}).Where("task_id = ? ", taskId).Update("status", status).Error; err != nil {
		return err
	}

	return nil
}

// parent
func (t *Task) GenerateTask(task_er ...Tasker) error {
	if len(task_er) == 0 {
		return fmt.Errorf("task is nil")
	}
	for _, t_er := range task_er {
		t.Register(t_er)
		t.WriteDb(*t_er.GetTask())
	}

	return t.WriteDb(*t)
}
