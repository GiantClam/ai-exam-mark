package handlers

import (
	"log"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/GiantClam/homework_marking/services"
	"github.com/GiantClam/homework_marking/utils"
)

// TaskHandler 处理任务相关请求
type TaskHandler struct {
	taskQueue *services.TaskQueue
}

// NewTaskHandler 创建任务处理器
func NewTaskHandler(taskQueue *services.TaskQueue) *TaskHandler {
	return &TaskHandler{
		taskQueue: taskQueue,
	}
}

// GetTaskStatus 获取任务状态
func (h *TaskHandler) GetTaskStatus(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		utils.RespondWithError(c, http.StatusBadRequest, "任务ID不能为空")
		return
	}

	log.Printf("查询任务状态: %s", taskID)

	task, exists := h.taskQueue.GetTask(taskID)
	if !exists {
		utils.RespondWithError(c, http.StatusNotFound, "任务不存在")
		return
	}

	// 计算进度时避免除以零
	var progress float64 = 0
	if task.TotalStudents > 0 {
		progress = float64(task.ProcessedCount) / float64(task.TotalStudents)
		
		// 确保进度值在0-1之间，避免出现NaN或Infinity
		if math.IsNaN(progress) || math.IsInf(progress, 0) {
			progress = 0
		} else if progress > 1 {
			progress = 1
		}
	}

	// 根据任务状态返回不同的响应
	switch task.Status {
	case services.TaskStatusPending, services.TaskStatusProcessing:
		// 返回任务进度
		c.JSON(http.StatusOK, gin.H{
			"status":         string(task.Status),
			"message":        "任务正在处理中",
			"task_id":        task.ID,
			"total_students": task.TotalStudents,
			"processed":      task.ProcessedCount,
			"progress":       progress, // 使用安全计算的进度值
			"start_time":     task.StartTime,
			// 如果有部分结果，可以返回已处理的结果
			"partial_results": task.Results,
		})
	case services.TaskStatusCompleted:
		// 返回完整结果
		c.JSON(http.StatusOK, gin.H{
			"status":   "completed",
			"message":  "任务处理完成",
			"task_id":  task.ID,
			"results":  task.Results,
			"end_time": task.EndTime,
			"total_students": task.TotalStudents,
			"processed":      task.ProcessedCount,
			"progress":       1.0, // 已完成的任务进度始终为100%
		})
	case services.TaskStatusFailed:
		// 返回错误信息
		c.JSON(http.StatusOK, gin.H{
			"status":   "failed",
			"message":  "任务处理失败",
			"task_id":  task.ID,
			"error":    task.Error,
			"end_time": task.EndTime,
			"progress": 0.0, // 失败的任务进度设为0
			"is_error": true,
		})
	default:
		utils.RespondWithError(c, http.StatusInternalServerError, "未知任务状态")
	}
}

// GetAllTasks 获取所有任务
func (h *TaskHandler) GetAllTasks(c *gin.Context) {
	// 获取任务计数
	counts := h.taskQueue.GetTasksCount()
	
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "任务统计",
		"counts":  counts,
	})
} 