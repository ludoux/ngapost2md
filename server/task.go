package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/ludoux/ngapost2md/nga"
)

type TaskStatus struct {
	Tid          int       `json:"tid"`
	AuthorId     int       `json:"authorId"`
	Type         string    `json:"type"`   // "download" | "update"
	Status       string    `json:"status"` // "queued" | "downloading" | "processing" | "generating_markdown" | "completed" | "failed" | "cancelled"
	CurrentPage  int       `json:"currentPage"`
	TotalPage    int       `json:"totalPage"`
	CurrentFloor int       `json:"currentFloor"`
	TotalFloor   int       `json:"totalFloor"`
	Stage        string    `json:"stage"`
	Error        string    `json:"error,omitempty"`
	StartTime    time.Time `json:"startTime,omitempty"`
	EndTime      time.Time `json:"endTime,omitempty"`
	QueueTime    time.Time `json:"queueTime"` // 入队时间
	OnComplete   func(success bool)
}

// TaskManager 管理任务队列，每次只执行一个任务
type TaskManager struct {
	mu      sync.Mutex
	queue   []*TaskStatus // 等待执行的任务队列（FIFO）
	running *TaskStatus   // 当前正在执行的任务（nil 表示空闲）
	hub     *WebSocketHub
	cond    *sync.Cond // 用于通知队列处理循环
}

func NewTaskManager(hub *WebSocketHub) *TaskManager {
	tm := &TaskManager{
		queue: make([]*TaskStatus, 0),
		hub:   hub,
	}
	tm.cond = sync.NewCond(&tm.mu)
	go tm.processQueue()
	return tm
}

// processQueue 是后台循环，每次只执行一个任务，完成后自动取下一个
func (tm *TaskManager) processQueue() {
	for {
		tm.mu.Lock()
		// 等待条件：队列空（无事可做）或正在有任务执行（需等完成）
		// 仅在队列非空且无任务运行时才继续
		for len(tm.queue) == 0 || tm.running != nil {
			tm.cond.Wait()
		}

		task := tm.queue[0]
		tm.queue = tm.queue[1:]
		tm.running = task
		task.Status = "downloading"
		task.StartTime = time.Now()
		tm.mu.Unlock()

		// 广播：任务开始执行
		tm.hub.BroadcastProgress(task)

		// 同步执行任务，完成后才取下一个
		tm.executeTask(task)

		// 任务完成，标记空闲
		tm.mu.Lock()
		tm.running = nil
		tm.mu.Unlock()

		// 通知循环可以继续取下一个任务
		tm.cond.Signal()
	}
}

func (tm *TaskManager) executeTask(task *TaskStatus) {
	tie := nga.Tiezi{}

	// 设置进度回调
	tie.ProgressCallback = func(stage string, currentPage, totalPage, currentFloor, totalFloor int) {
		tm.mu.Lock()
		task.Stage = stage
		task.CurrentPage = currentPage
		task.TotalPage = totalPage
		task.CurrentFloor = currentFloor
		task.TotalFloor = totalFloor

		switch stage {
		case "downloading":
			task.Status = "downloading"
		case "processing_content":
			task.Status = "processing"
		case "generating_markdown":
			task.Status = "generating_markdown"
		case "completed":
			task.Status = "completed"
			task.EndTime = time.Now()
		}
		tm.mu.Unlock()

		tm.hub.BroadcastProgress(task)
	}

	var initErr error
	if task.Type == "update" {
		posts := ScanAllPosts()
		authorId := task.AuthorId
		for _, p := range posts {
			if p.Tid == task.Tid {
				authorId = p.AuthorId
				break
			}
		}
		initErr = tie.InitFromLocal(task.Tid, authorId)
	} else {
		initErr = tie.InitFromWeb(task.Tid, task.AuthorId)
	}

	if initErr != nil {
		tm.mu.Lock()
		task.Status = "failed"
		task.Error = initErr.Error()
		task.EndTime = time.Now()
		tm.mu.Unlock()
		tm.hub.BroadcastTaskFailed(task)
		if task.OnComplete != nil {
			task.OnComplete(false)
		}
		return
	}

	downloadErr := tie.Download()
	if downloadErr != nil {
		tm.mu.Lock()
		task.Status = "failed"
		task.Error = downloadErr.Error()
		task.EndTime = time.Now()
		tm.mu.Unlock()
		tm.hub.BroadcastTaskFailed(task)
		if task.OnComplete != nil {
			task.OnComplete(false)
		}
		return
	}

	tm.hub.BroadcastTaskComplete(task)
	if task.OnComplete != nil {
		task.OnComplete(true)
	}
}

// EnqueueDownload 将下载任务加入队列
func (tm *TaskManager) EnqueueDownload(tid int, authorId int, onComplete func(bool)) (*TaskStatus, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.isTidInUse(tid) {
		return nil, fmt.Errorf("tid %d 已在队列或正在执行中", tid)
	}

	task := &TaskStatus{
		Tid:        tid,
		AuthorId:   authorId,
		Type:       "download",
		Status:     "queued",
		QueueTime:  time.Now(),
		OnComplete: onComplete,
	}
	tm.queue = append(tm.queue, task)
	tm.cond.Signal() // 通知 processQueue 有新任务
	return task, nil
}

// EnqueueUpdate 将更新任务加入队列
func (tm *TaskManager) EnqueueUpdate(tid int, onComplete func(bool)) (*TaskStatus, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.isTidInUse(tid) {
		return nil, fmt.Errorf("tid %d 已在队列或正在执行中", tid)
	}

	task := &TaskStatus{
		Tid:        tid,
		Type:       "update",
		Status:     "queued",
		QueueTime:  time.Now(),
		OnComplete: onComplete,
	}
	tm.queue = append(tm.queue, task)
	tm.cond.Signal()
	return task, nil
}

// isTidInUse 检查 tid 是否在队列中或正在执行（需在 mu.Lock 内调用）
func (tm *TaskManager) isTidInUse(tid int) bool {
	if tm.running != nil && tm.running.Tid == tid {
		return true
	}
	for _, t := range tm.queue {
		if t.Tid == tid {
			return true
		}
	}
	return false
}

// CancelTask 取消/移除任务。正在运行的任务无法取消（nga 无取消机制），队列中的任务可以移除。
func (tm *TaskManager) CancelTask(tid int) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// 正在运行的任务不允许取消
	if tm.running != nil && tm.running.Tid == tid {
		return fmt.Errorf("tid %d 正在执行中，无法取消", tid)
	}

	// 从队列中移除
	for i, t := range tm.queue {
		if t.Tid == tid {
			t.Status = "cancelled"
			t.EndTime = time.Now()
			tm.hub.BroadcastTaskFailed(t)
			tm.queue = append(tm.queue[:i], tm.queue[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("tid %d 不在队列中", tid)
}

// GetAllTasks 返回当前正在运行的任务 + 队列中等待的任务
func (tm *TaskManager) GetAllTasks() []TaskStatus {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	var result []TaskStatus
	if tm.running != nil {
		result = append(result, *tm.running)
	}
	for _, t := range tm.queue {
		result = append(result, *t)
	}
	return result
}

// PostInfo 表示一个已下载的帖子
type PostInfo struct {
	Tid         int    `json:"tid"`
	AuthorId    int    `json:"authorId"`
	Title       string `json:"title"`
	FolderName  string `json:"folderName"`
	MaxPage     int    `json:"maxPage"`
	MaxFloor    int    `json:"maxFloor"`
	FloorCount  int    `json:"floorCount,omitempty"`
	HasMarkdown bool   `json:"hasMarkdown"`
	CreatedTime string `json:"createdTime"`
	UpdatedTime string `json:"updatedTime"`
}
