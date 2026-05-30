package server

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/cast"
)
var cronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)




type ScheduleTask struct {
	ID            string       `json:"id"`
	Tid           int          `json:"tid"`
	AuthorId      int          `json:"authorId"`
	Cron          string       `json:"cron"`
	Enabled       bool         `json:"enabled"`
	LastRunTime   string       `json:"lastRunTime,omitempty"`
	NextRunTime   string       `json:"nextRunTime,omitempty"`
	LastRunStatus string       `json:"lastRunStatus,omitempty"` // "completed" | "failed" | ""
	CreatedTime   string       `json:"createdTime"`
	CronEntryID   cron.EntryID // 内部使用，不序列化
}

type ScheduleData struct {
	Tasks []ScheduleTask `json:"tasks"`
}

type ScheduleManager struct {
	mu       sync.Mutex
	data     ScheduleData
	filePath string
	cron     *cron.Cron
	nextID   int
	tm       *TaskManager
}

func NewScheduleManager(tm *TaskManager) *ScheduleManager {
	return &ScheduleManager{
		filePath: "schedules.json",
		nextID:   1,
		tm:       tm,
	}
}

func (sm *ScheduleManager) Load() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, err := os.Stat(sm.filePath); os.IsNotExist(err) {
		sm.data = ScheduleData{Tasks: []ScheduleTask{}}
		return nil
	}

	data, err := os.ReadFile(sm.filePath)
	if err != nil {
		return fmt.Errorf("读取 schedules.json 失败: %v", err)
	}

	if err := json.Unmarshal(data, &sm.data); err != nil {
		return fmt.Errorf("解析 schedules.json 失败: %v", err)
	}

	// 计算 nextID
	for _, task := range sm.data.Tasks {
		id := cast.ToInt(task.ID)
		if id >= sm.nextID {
			sm.nextID = id + 1
		}
	}

	return nil
}

func (sm *ScheduleManager) Start() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.cron = cron.New()

	for i, task := range sm.data.Tasks {
		if task.Enabled {
			// 捕获值而非循环变量，避免闭包共享导致所有定时任务都指向最后一个
			taskTid := task.Tid
			taskID := task.ID
			entryID, err := sm.cron.AddFunc(task.Cron, func() {
				sm.runScheduleTask(taskTid, taskID)
			})
			if err != nil {
				log.Printf("添加定时任务 %s 的 cron job 失败: %v，已自动禁用该任务", task.ID, err)
				sm.data.Tasks[i].Enabled = false
				continue
			}
			sm.data.Tasks[i].CronEntryID = entryID
		}
		// 计算 NextRunTime
		sm.updateNextRunTime(i)
	}

	sm.cron.Start()
	return nil
}

func (sm *ScheduleManager) Stop() {
	if sm.cron != nil {
		sm.cron.Stop()
	}
}

func (sm *ScheduleManager) runScheduleTask(tid int, scheduleID string) {
	log.Printf("定时任务 %s 触发: tid=%d", scheduleID, tid)

	_, err := sm.tm.EnqueueUpdate(tid, func(success bool) {
		if success {
			sm.updateLastRun(scheduleID, "completed")
			log.Printf("定时任务 %s 执行成功: tid=%d", scheduleID, tid)
		} else {
			sm.updateLastRun(scheduleID, "failed")
			log.Printf("定时任务 %s 执行失败: tid=%d", scheduleID, tid)
		}
	})
	if err != nil {
		log.Printf("定时任务 %s 入队失败: %v", scheduleID, err)
		sm.updateLastRun(scheduleID, "failed")
		return
	}

	log.Printf("定时任务 %s 已入队: tid=%d", scheduleID, tid)
}

func (sm *ScheduleManager) GetAllSchedules() []ScheduleTask {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 更新所有任务的 NextRunTime
	for i := range sm.data.Tasks {
		sm.updateNextRunTime(i)
	}

	return sm.data.Tasks
}

func (sm *ScheduleManager) AddSchedule(tid int, authorId int, cronExpr string) (*ScheduleTask, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 验证 cron 表达式

	schedule, err := cronParser.Parse(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("cron 表达式不合法: %v", err)
	}

	task := ScheduleTask{
		ID:          cast.ToString(sm.nextID),
		Tid:         tid,
		AuthorId:    authorId,
		Cron:        cronExpr,
		Enabled:     true,
		CreatedTime: time.Now().Format(time.RFC3339),
	}
	sm.nextID++

	// 添加 cron job
	entryID, err := sm.cron.AddFunc(cronExpr, func() {
		sm.runScheduleTask(tid, task.ID)
	})
	if err != nil {
		return nil, fmt.Errorf("添加 cron job 失败: %v", err)
	}
	task.CronEntryID = entryID

	// 计算 NextRunTime
	nextRun := schedule.Next(time.Now())
	task.NextRunTime = nextRun.Format(time.RFC3339)

	sm.data.Tasks = append(sm.data.Tasks, task)
	sm.save()

	return &task, nil
}

func (sm *ScheduleManager) UpdateSchedule(id string, cronExpr string, enabled bool) (*ScheduleTask, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	idx := -1
	for i, task := range sm.data.Tasks {
		if task.ID == id {
			idx = i
			break
		}
	}

	if idx == -1 {
		return nil, fmt.Errorf("定时任务 %s 不存在", id)
	}

	task := &sm.data.Tasks[idx]

	// 如果 cron 表达式变了或启禁状态变了，需要重新注册 cron job
	if cronExpr != "" && cronExpr != task.Cron || enabled != task.Enabled {
		// 先移除旧的 cron job
		if task.Enabled && task.CronEntryID != 0 {
			sm.cron.Remove(task.CronEntryID)
			task.CronEntryID = 0
		}

		if cronExpr != "" {
			task.Cron = cronExpr
		}
		task.Enabled = enabled

		// 如果启用，添加新的 cron job
		if task.Enabled {
			// 捕获值而非 slice 元素指针，避免 slice 扩容后指针失效
			cronTid := task.Tid
			cronID := task.ID
			entryID, err := sm.cron.AddFunc(task.Cron, func() {
				sm.runScheduleTask(cronTid, cronID)
			})
			if err != nil {
				return nil, fmt.Errorf("更新 cron job 失败: %v", err)
			}
			task.CronEntryID = entryID
		}
	}

	sm.updateNextRunTime(idx)
	sm.save()

	return task, nil
}

func (sm *ScheduleManager) DeleteSchedule(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	idx := -1
	for i, task := range sm.data.Tasks {
		if task.ID == id {
			idx = i
			break
		}
	}

	if idx == -1 {
		return fmt.Errorf("定时任务 %s 不存在", id)
	}

	task := sm.data.Tasks[idx]

	// 移除 cron job
	if task.Enabled && task.CronEntryID != 0 {
		sm.cron.Remove(task.CronEntryID)
	}

	sm.data.Tasks = append(sm.data.Tasks[:idx], sm.data.Tasks[idx+1:]...)
	sm.save()

	return nil
}

func (sm *ScheduleManager) updateLastRun(scheduleID string, status string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, task := range sm.data.Tasks {
		if task.ID == scheduleID {
			sm.data.Tasks[i].LastRunTime = time.Now().Format(time.RFC3339)
			sm.data.Tasks[i].LastRunStatus = status
			sm.updateNextRunTime(i)
			break
		}
	}
	sm.save()
}

func (sm *ScheduleManager) updateNextRunTime(idx int) {
	if sm.cron == nil {
		return
	}
	task := sm.data.Tasks[idx]
	if task.Enabled && task.Cron != "" {
	
		schedule, err := cronParser.Parse(task.Cron)
		if err == nil {
			nextRun := schedule.Next(time.Now())
			sm.data.Tasks[idx].NextRunTime = nextRun.Format(time.RFC3339)
		}
	} else {
		sm.data.Tasks[idx].NextRunTime = ""
	}
}

func (sm *ScheduleManager) save() {
	data, err := json.MarshalIndent(sm.data, "", "  ")
	if err != nil {
		log.Println("序列化 schedules.json 失败:", err)
		return
	}

	if err := os.WriteFile(sm.filePath, data, 0666); err != nil {
		log.Println("保存 schedules.json 失败:", err)
	}
}
