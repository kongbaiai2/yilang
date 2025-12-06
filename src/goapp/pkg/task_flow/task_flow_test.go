
package task

import (
        "encoding/json"
        "fmt"
        "log"
        "runtime"
        "testing"
        "time"

        "gorm.io/driver/mysql"
        "gorm.io/gorm"
)

type Dst struct {
        HostId   string
        HostName string
}
type TaskOne struct {
        d Dst
        T Task
}

func (to *TaskOne) GetTask() *Task {
        to.T.Action = "TaskOne"
        return &to.T
}
func (to *TaskOne) PreDo(t_er Tasker) error {
        to.d.HostName = "TaskOne"
        if to.T.Status == StatusRetry { // 模拟重试成功
                log.Print(to.d.HostName, "--", to.T.Retry)
                return nil
        }
        log.Println("TaskOne PreDo err")
        t_er.GetTask().OptGlobalMap(t_er, "TaskOnekey", "TaskOnevalue")
        return fmt.Errorf("TaskOne predo error")
}

func (to *TaskOne) Do(t_er Tasker) error {
        log.Print(to.d.HostName)
        return nil
}
func (to *TaskOne) PostDo(t_er Tasker) error {
        log.Println("TaskOne PostDo")
        return nil
}
func (to *TaskOne) Schedule() (bool, error) {
        return false, nil
}
func (to *TaskOne) GetContext() (string, error) {
        jsonBytes, err := json.Marshal(to.d)
        if err != nil {
                return "", err
        }
        return string(jsonBytes), nil
}

type TaskTwo struct {
        Name string
        T    Task
}

func (to *TaskTwo) GetTask() *Task {
        to.T.Action = "TaskTwo"
        return &to.T
}
func (to *TaskTwo) PreDo(t_er Tasker) error {
        to.Name = "TaskTwo"
        log.Println("TaskTwo PreDo")
        return nil
}

func (to *TaskTwo) Do(t_er Tasker) error {
        log.Print(to.Name)
        if to.T.Status == StatusRetry { // 模拟重试成功
                log.Print(to.Name, "--", to.T.Retry)
                return nil
        }
        return fmt.Errorf("taskTwo do error") // 模拟失败
        // return nil
}
func (to *TaskTwo) PostDo(t_er Tasker) error {
        log.Println("TaskTwo PostDo")
        return nil
}
func (to *TaskTwo) Schedule() (bool, error) {
        return false, nil
}
func (to *TaskTwo) GetContext() (string, error) {
        jsonBytes, err := json.Marshal(to.Name)
        if err != nil {
                return "", err
        }
        return string(jsonBytes), nil
}

type TaskThree struct {
        Name string
        T    Task
}

func (to *TaskThree) GetTask() *Task {
        to.T.Action = "TaskThree"
        return &to.T
}
func (to *TaskThree) PreDo(t_er Tasker) error {
        to.Name = "TaskThree"
        // return fmt.Errorf("format string, a ...any")
        log.Println("TaskThree PreDo")
        return nil
}

func (to *TaskThree) Do(t_er Tasker) error {
        log.Print(to.Name)
        mapPool := t_er.GetTask().OptGlobalMap(t_er, "TaskFourkey", "TaskThreevalue")
        log.Println(mapPool)
        return nil
}
func (to *TaskThree) PostDo(t_er Tasker) error {
        if to.T.Status == StatusRetry { // 模拟重试成功
                log.Print(to.Name, "--", to.T.Retry)
                return nil
        }
        log.Println("TaskThree PostDo")
        return fmt.Errorf("taskThree postdo error")
}
func (to *TaskThree) Schedule() (bool, error) {
        return false, nil
}
func (to *TaskThree) GetContext() (string, error) {
        jsonBytes, err := json.Marshal(to.Name)
        if err != nil {
                return "", err
        }
        return string(jsonBytes), nil
}

type TaskFour struct {
        Name string
        T    Task
}

func (to *TaskFour) GetTask() *Task {
        to.T.Action = "TaskFour"
        return &to.T
}
func (to *TaskFour) PreDo(t_er Tasker) error {
        to.Name = "TaskFour"
        if to.T.Status == StatusRetry { // 模拟重试成功
                log.Print(to.Name, "--", to.T.Retry)
                return nil
        }
        log.Println("TaskFour PreDo err")
        return fmt.Errorf("TaskFour err")
        // return nil
}

func (to *TaskFour) Do(t_er Tasker) error {
        log.Print(to.Name)
        mapPool := t_er.GetTask().OptGlobalMap(t_er, "", "")
        log.Println(mapPool)
        t_er.GetTask().OptGlobalMap(t_er, "TaskFourkey", "TaskFourvalue")
        log.Println(mapPool)
        return nil
}
func (to *TaskFour) PostDo(t_er Tasker) error {
        return nil
}
func (to *TaskFour) Schedule() (bool, error) {
        return false, nil
}

func (to *TaskFour) GetContext() (string, error) {
        jsonBytes, err := json.Marshal(to.Name)
        if err != nil {
                return "", err
        }
        return string(jsonBytes), nil
}

type TaskFive struct {
        Name string
        T    Task
}

func (to *TaskFive) GetTask() *Task {
        to.T.Action = "TaskFive"
        return &to.T
}
func (to *TaskFive) PreDo(t_er Tasker) error {
        to.Name = "TaskFive"
        // return fmt.Errorf("format string, a ...any")
        return nil
}

func (to *TaskFive) Do(t_er Tasker) error {
        log.Print(to.Name)
        mapPool := t_er.GetTask().OptGlobalMap(t_er, "", "")
        log.Println(mapPool)
        return nil
}
func (to *TaskFive) PostDo(t_er Tasker) error {
        return nil
}
func (to *TaskFive) Schedule() (bool, error) {
        return false, nil
}
func (to *TaskFive) GetContext() (string, error) {
        jsonBytes, err := json.Marshal(to.Name)
        if err != nil {
                return "", err
        }
        return string(jsonBytes), nil
}

type TaskSix struct {
        Name string
        T    Task
}

func (to *TaskSix) GetTask() *Task {
        to.T.Action = "TaskSix"
        return &to.T
}
func (to *TaskSix) PreDo(t_er Tasker) error {
        to.Name = "TaskSix"
        // return fmt.Errorf("format string, a ...any")
        return nil
}

func (to *TaskSix) Do(t_er Tasker) error {
        log.Print(to.Name)
        mapPool := t_er.GetTask().OptGlobalMap(t_er, "", "")
        log.Println(mapPool)
        mapPool = t_er.GetTask().OptGlobalMap(t_er, "TaskFourkey", "TaskSixvalue")
        log.Println(mapPool)
        return nil
}
func (to *TaskSix) PostDo(t_er Tasker) error {
        return nil
}
func (to *TaskSix) Schedule() (bool, error) {
        return false, nil
}
func (to *TaskSix) GetContext() (string, error) {
        jsonBytes, err := json.Marshal(to.Name)
        if err != nil {
                return "", err
        }
        return string(jsonBytes), nil
}

func InitDB(mySqlConf string) *gorm.DB {

        var err error
        db, err := gorm.Open(mysql.Open(mySqlConf), &gorm.Config{})
        if err != nil {
                log.Fatal(err.Error())
        }

        sqlDb, err := db.DB()
        if err != nil {
                log.Fatal(err.Error())
        }

        sqlDb.SetMaxIdleConns(50)
        sqlDb.SetMaxOpenConns(100)

        err = db.AutoMigrate(&Task{})
        if err != nil {
                log.Printf("%+v", err)
        }

        return db
}
func GetGoroutineID() int {
        var buf [64]byte
        runtime.Stack(buf[:], false)
        var id int
        fmt.Sscanf(string(buf[:]), "goroutine %d", &id)
        return id
}

func Test_task_flow(test *testing.T) {
        dns := `root:zabbix123.com@tcp(127.0.0.1:3306)/testmac?charset=utf8&parseTime=True&loc=Asia%2FShanghai`
        Gdb = InitDB(dns)
        var ttt5 Task
        ttt5.Action = "Task-ttt5"
        ttt5.GenerateTask(&TaskThree{})

        var tttt Task
        tttt.Action = "Task-tttt"
        tttt.GenerateTask(&TaskFour{}, &TaskFive{}, &ttt5)

        var ttt Task
        ttt.Action = "Task-ttt"
        ttt.GenerateTask(&TaskSix{}, &tttt)

        var tt Task
        tt.Action = "Task-tt"
        tt.GenerateTask(&TaskFour{}, &TaskFive{}, &ttt)

        var t Task
        t.Action = "Task"
        err := t.GenerateTask(&tt, &TaskOne{}, &TaskTwo{}, &TaskThree{})

        if err != nil {
                log.Print(err)
                return
        }
        // log.Println(t.SubTaskDb)

        _, err = t.Schedule()
        if err != nil {
                log.Print(err)
                // return
        }
        log.Println(t.GlobalMap)
        time.Sleep(5 * time.Second)
        err = t.RetryDo()
        if err != nil {
                log.Print(err)

        }
        log.Println(t.GlobalMap)
        time.Sleep(5 * time.Second)
        err = t.RetryDo()
        if err != nil {
                log.Print(err)

        }
        log.Println(t.GlobalMap)
        time.Sleep(5 * time.Second)
        err = t.RetryDo()
        if err != nil {
                log.Print(err)

        }
        log.Println(t.GlobalMap)
        time.Sleep(5 * time.Second)
        err = t.RetryDo()
        if err != nil {
                log.Print(err)

        }
        log.Println(t.GlobalMap)
        log.Println(t.GlobalMap)
        time.Sleep(5 * time.Second)
        err = t.RetryDo()
        if err != nil {
                log.Print(err)
                PrintStack()
        }
        log.Println(t.GlobalMap)
        time.Sleep(5 * time.Second)
        err = t.RetryDo()
        if err != nil {
                log.Print(err)

        }
        log.Println(t.GlobalMap)

        time.Sleep(5 * time.Second)
        err = t.RetryDo()
        if err != nil {
                log.Print(err)

        }
        log.Println(t.GlobalMap)
}


