package services

import (
	"log"
	"sync"
	"time"
)

// TaskStatus 表示任务状态
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"   // 待处理
	TaskStatusProcessing TaskStatus = "processing" // 处理中
	TaskStatusCompleted TaskStatus = "completed" // 已完成
	TaskStatusFailed    TaskStatus = "failed"    // 失败
)

// HomeworkTask 表示一个作业处理任务
type HomeworkTask struct {
	ID              string                 `json:"id"`              // 任务ID
	FilePath        string                 `json:"filePath"`        // 文件路径
	HomeworkType    string                 `json:"homeworkType"`    // 作业类型
	PagesPerStudent int                    `json:"pagesPerStudent"` // 每个学生的页数
	Layout          string                 `json:"layout"`          // 布局方式
	TotalStudents   int                    `json:"totalStudents"`   // 学生总数
	ProcessedCount  int                    `json:"processedCount"`  // 已处理学生数
	StartTime       time.Time              `json:"startTime"`       // 开始时间
	EndTime         *time.Time             `json:"endTime"`         // 结束时间
	Status          TaskStatus             `json:"status"`          // 任务状态
	Error           string                 `json:"error,omitempty"` // 错误信息
	Results         []string               `json:"results"`         // 每个学生的处理结果
	Params          map[string]interface{} `json:"params"`          // 其他参数
	ProcessFunc     TaskProcessFunc        `json:"-"`              // 处理函数，不导出到JSON
}

// TaskQueue 简单的内存任务队列
type TaskQueue struct {
	tasks     map[string]*HomeworkTask
	tasksChan chan *HomeworkTask
	mutex     sync.RWMutex
	wg        sync.WaitGroup
	workerCount int
}

// NewTaskQueue 创建一个新的任务队列
func NewTaskQueue(workerCount int) *TaskQueue {
	q := &TaskQueue{
		tasks:       make(map[string]*HomeworkTask),
		tasksChan:   make(chan *HomeworkTask, 100), // 缓冲大小
		workerCount: workerCount,
	}

	// 启动工作协程
	for i := 0; i < workerCount; i++ {
		go q.worker(i)
	}

	return q
}

// worker 处理任务的工作协程
func (q *TaskQueue) worker(id int) {
	log.Printf("[INFO] 启动工作协程 #%d", id)
	
	for task := range q.tasksChan {
		log.Printf("[INFO] 工作协程 #%d 开始处理任务: %s", id, task.ID)
		q.processTask(task)
	}
}

// processTask 处理单个任务
func (q *TaskQueue) processTask(task *HomeworkTask) {
	// 更新任务状态为处理中
	q.updateTaskStatus(task.ID, TaskStatusProcessing)

	// 避免panic导致工作协程退出
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[ERROR] 任务处理出现panic: %v", r)
			now := time.Now()
			
			q.mutex.Lock()
			task.Status = TaskStatusFailed
			task.Error = "处理过程中出现未知错误"
			task.EndTime = &now
			q.mutex.Unlock()
		}
		
		q.wg.Done() // 完成一个任务
	}()

	// 在这里调用实际的处理函数
	// 这里我们需要在外部注入实际的处理函数
	if task.ProcessFunc != nil {
		task.ProcessFunc(task)
	} else {
		log.Printf("[ERROR] 任务 %s 没有设置处理函数", task.ID)
		q.updateTaskStatus(task.ID, TaskStatusFailed)
		task.Error = "未设置处理函数"
	}
}

// 添加处理函数类型
type TaskProcessFunc func(task *HomeworkTask)

// 为任务添加处理函数字段
func (q *TaskQueue) RegisterProcessFunc(taskID string, processFunc TaskProcessFunc) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if task, exists := q.tasks[taskID]; exists {
		task.ProcessFunc = processFunc
	}
}

// 为HomeworkTask添加处理函数字段（不导出）
func (t *HomeworkTask) WithProcessFunc(fn TaskProcessFunc) *HomeworkTask {
	t.ProcessFunc = fn
	return t
}

// ProcessFunc 任务处理函数
type ProcessFunc func(task *HomeworkTask)

// AddTask 添加新任务到队列
func (q *TaskQueue) AddTask(task *HomeworkTask) {
	q.mutex.Lock()
	q.tasks[task.ID] = task
	q.mutex.Unlock()

	task.Status = TaskStatusPending
	task.StartTime = time.Now()
	task.ProcessedCount = 0

	// 增加等待计数
	q.wg.Add(1)
	
	// 发送到任务通道
	q.tasksChan <- task
	
	log.Printf("[INFO] 添加任务到队列: %s", task.ID)
}

// GetTask 获取任务信息
func (q *TaskQueue) GetTask(taskID string) (*HomeworkTask, bool) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	
	task, exists := q.tasks[taskID]
	return task, exists
}

// updateTaskStatus 更新任务状态
func (q *TaskQueue) updateTaskStatus(taskID string, status TaskStatus) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if task, exists := q.tasks[taskID]; exists {
		task.Status = status
		if status == TaskStatusCompleted || status == TaskStatusFailed {
			now := time.Now()
			task.EndTime = &now
		}
	}
}

// UpdateTaskProgress 更新任务处理进度
func (q *TaskQueue) UpdateTaskProgress(taskID string, processedCount int, result string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if task, exists := q.tasks[taskID]; exists {
		task.ProcessedCount = processedCount
		if result != "" {
			task.Results = append(task.Results, result)
		}
	}
}

// UpdateTaskTotalStudents 更新任务的学生总数
func (q *TaskQueue) UpdateTaskTotalStudents(taskID string, totalStudents int) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if task, exists := q.tasks[taskID]; exists {
		task.TotalStudents = totalStudents
		log.Printf("[INFO] 更新任务 %s 的学生总数: %d", taskID, totalStudents)
	} else {
		log.Printf("[ERROR] 更新学生总数失败: 任务 %s 不存在", taskID)
	}
}

// AddTaskResult 添加任务处理结果
func (q *TaskQueue) AddTaskResult(taskID string, result string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if task, exists := q.tasks[taskID]; exists {
		task.Results = append(task.Results, result)
		log.Printf("[INFO] 添加任务 %s 的结果，目前共 %d 个结果", taskID, len(task.Results))
	} else {
		log.Printf("[ERROR] 添加结果失败: 任务 %s 不存在", taskID)
	}
}

// IncrementProcessedCount 增加已处理学生数量
func (q *TaskQueue) IncrementProcessedCount(taskID string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	if task, exists := q.tasks[taskID]; exists {
		task.ProcessedCount++
		log.Printf("[INFO] 更新任务 %s 的处理进度: %d/%d", 
			taskID, task.ProcessedCount, task.TotalStudents)
	} else {
		log.Printf("[ERROR] 增加处理计数失败: 任务 %s 不存在", taskID)
	}
}

// CompleteTask 将任务标记为完成
func (q *TaskQueue) CompleteTask(taskID string, result string) {
	q.mutex.Lock()
	if task, exists := q.tasks[taskID]; exists {
		task.Status = TaskStatusCompleted
		now := time.Now()
		task.EndTime = &now
		if result != "" {
			// 添加最终结果
			task.Results = append(task.Results, result)
		}
	}
	q.mutex.Unlock()
	
	log.Printf("[INFO] 任务已完成: %s", taskID)
}

// FailTask 将任务标记为失败
func (q *TaskQueue) FailTask(taskID string, err string) {
	q.mutex.Lock()
	if task, exists := q.tasks[taskID]; exists {
		task.Status = TaskStatusFailed
		task.Error = err
		now := time.Now()
		task.EndTime = &now
	}
	q.mutex.Unlock()
	
	log.Printf("[ERROR] 任务失败: %s, 错误: %s", taskID, err)
}

// Wait 等待所有任务完成
func (q *TaskQueue) Wait() {
	q.wg.Wait()
}

// Close 关闭任务队列
func (q *TaskQueue) Close() {
	close(q.tasksChan)
}

// CleanupTasks 清理旧任务
func (q *TaskQueue) CleanupTasks(ageHours int) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	cutoff := time.Now().Add(-time.Duration(ageHours) * time.Hour)
	
	for id, task := range q.tasks {
		if task.EndTime != nil && task.EndTime.Before(cutoff) {
			delete(q.tasks, id)
			log.Printf("[INFO] 清理了旧任务: %s", id)
		}
	}
}

// GetTasksCount 获取队列中的任务数量
func (q *TaskQueue) GetTasksCount() map[TaskStatus]int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	
	counts := map[TaskStatus]int{
		TaskStatusPending:    0,
		TaskStatusProcessing: 0,
		TaskStatusCompleted:  0,
		TaskStatusFailed:     0,
	}
	
	for _, task := range q.tasks {
		counts[task.Status]++
	}
	
	return counts
}

// CreateTask 创建一个新的任务并返回任务ID
func (q *TaskQueue) CreateTask(taskType, message string) string {
	taskID := GenerateTaskID()
	task := &HomeworkTask{
		ID:        taskID,
		Status:    TaskStatusPending,
		StartTime: time.Now(),
		Results:   make([]string, 0),
		Params:    map[string]interface{}{"type": taskType, "message": message},
	}
	
	q.mutex.Lock()
	q.tasks[taskID] = task
	q.mutex.Unlock()
	
	log.Printf("[INFO] 创建任务: %s, 类型: %s", taskID, taskType)
	return taskID
}

// UpdateTaskStatus 更新任务状态和消息
func (q *TaskQueue) UpdateTaskStatus(taskID, status, message string) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	
	task, exists := q.tasks[taskID]
	if !exists {
		log.Printf("[ERROR] 更新状态失败: 任务 %s 不存在", taskID)
		return
	}
	
	// 将字符串状态转换为TaskStatus类型
	var taskStatus TaskStatus
	switch status {
	case "pending":
		taskStatus = TaskStatusPending
	case "processing":
		taskStatus = TaskStatusProcessing
	case "completed":
		taskStatus = TaskStatusCompleted
		now := time.Now()
		task.EndTime = &now
	case "error", "failed":
		taskStatus = TaskStatusFailed
		now := time.Now()
		task.EndTime = &now
		task.Error = message
	default:
		log.Printf("[WARN] 未知任务状态: %s", status)
		return
	}
	
	task.Status = taskStatus
	
	// 如果是完成状态，将消息添加到结果中
	if taskStatus == TaskStatusCompleted && message != "" {
		task.Results = append(task.Results, message)
	}
	
	log.Printf("[INFO] 更新任务状态: %s -> %s", taskID, status)
}

// GenerateTaskID 生成唯一的任务ID
func GenerateTaskID() string {
	return "task_" + time.Now().Format("20060102150405") + 
		"_" + RandStringRunes(6)
}

// RandStringRunes 生成随机字符串
func RandStringRunes(n int) string {
	// 简化实现，实际使用时可以使用uuid等库
	return time.Now().Format("150405")
}

// GetAllTasks 获取所有任务的列表
func (q *TaskQueue) GetAllTasks() []*HomeworkTask {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	
	tasks := make([]*HomeworkTask, 0, len(q.tasks))
	for _, task := range q.tasks {
		tasks = append(tasks, task)
	}
	
	return tasks
} 