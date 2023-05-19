package utils

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
)

/*定义一个slice队列*/
type SliceQueue struct {
	data []interface{}
}
type FlagInfo struct {
	AbnormalTime int64
}

func NewSliceQueue(n int) (q *SliceQueue) {
	return &SliceQueue{data: make([]interface{}, 0, n)}
}

// Enqueue 把值放在队尾
func (q *SliceQueue) Enqueue(v interface{}) {
	q.data = append(q.data, v)
}

// Dequeue 移去队头并返回
func (q *SliceQueue) Dequeue() interface{} {
	if len(q.data) == 0 {
		return nil
	}
	v := q.data[0]
	q.data = q.data[1:]
	return v
}

func (q *SliceQueue) Len() int {
	return len(q.data)
}

func (q *SliceQueue) Get(index int) interface{} {
	return q.data[index]
}

func (q *SliceQueue) Clear() {
	q.data = make([]interface{}, 0, 0)
}

func (q *SliceQueue) Sub(start, end int64) int64 {
	if start >= end {
		return start - end
	}
	return end - start
}

func (q *SliceQueue) IsValid(cron int64, warnId uint, structKey string) bool {
	if uint(q.Len()) >= warnId {
		// 第一个队列中，取结构体中名为structKey的值
		firstTime, err := q.UseReflectGetValue(q.Get(0), structKey)
		if err != nil {
			log.Panic("1v ...any", err)
		}
		// 最后一个队列中，取结构体中名为structKey的值
		endTime, err := q.UseReflectGetValue(q.Get(q.Len()-1), structKey)
		if err != nil {
			log.Panic("2v ...any", err)
		}

		// log.Printf("%v-%v <%v", endTime, firstTime, cron)
		// 相减
		interval := q.Sub(firstTime, endTime)
		if interval < cron {
			// 清空,命中返回false
			log.Println(q.data...)
			q.Clear()
			return false
		} else {
			// 去头，返回true，go on
			q.Dequeue()
			return true
		}
	}
	return true
}

func (q *SliceQueue) UseReflectGetValue(b interface{}, keyName string) (int64, error) {
	ele := reflect.ValueOf(b).Elem()
	t := ele.Type()
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Name == keyName {
			str := fmt.Sprintf("%v", ele.Field(i))
			ret, err := strconv.ParseInt(str, 10, 64)
			if err != nil {
				return 0, err
			}
			return ret, nil
		}
	}
	return 0, fmt.Errorf("key not found")
}

// 判断指针类型为空指针
func IsNil(i interface{}) bool {
	vi := reflect.ValueOf(i)
	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}
	return false
}

func (q *SliceQueue) IsWithinFiveMinutes(cron int64, structKey string) bool {
	if q.Len()-1 <= 0 {
		return true
	}
	firstTime, err := q.UseReflectGetValue(q.Get(0), structKey)
	if err != nil {
		log.Panic("1v ...any", err)
	}
	endTime, err := q.UseReflectGetValue(q.Get(q.Len()-1), structKey)
	if err != nil {
		log.Panic("2v ...any", err)
	}

	interval := q.Sub(firstTime, endTime)
	if interval < cron {
		return true
	} else {
		// log.Println(q.data...)
		q.Clear()
		return false
	}
}

// func test() {
// 	var slbTrafficRecordQ *SliceQueue
// 	// 判断，retNm次错误，并且在cron秒内的，则返回err，否返回nil
// 	slbTrafficRecordQ.Enqueue(&slbFlagInfo{sshHost: iplist, abnormalTime: curTime})
// 	if !slbTrafficRecordQ.IsValid(cron, retNm, "abnormalTime") {
// 		return fmt.Errorf("slb-lvs check traffic double 0.0, host: %s, err:%v", iplist, errArr)
// 	}
// }
