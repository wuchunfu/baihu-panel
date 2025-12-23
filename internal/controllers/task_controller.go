package controllers

import (
	"path/filepath"
	"strconv"

	"baihu/internal/constant"
	"baihu/internal/services"
	"baihu/internal/utils"

	"github.com/gin-gonic/gin"
)

type TaskController struct {
	taskService *services.TaskService
	cronService *services.CronService
}

func NewTaskController(taskService *services.TaskService, cronService *services.CronService) *TaskController {
	return &TaskController{
		taskService: taskService,
		cronService: cronService,
	}
}

// resolveWorkDir 将相对路径转换为绝对路径
func resolveWorkDir(workDir string) string {
	if workDir == "" {
		// 空则使用默认 scripts 目录
		absPath, err := filepath.Abs(constant.ScriptsWorkDir)
		if err != nil {
			return constant.ScriptsWorkDir
		}
		return absPath
	}
	// 如果已经是绝对路径，直接返回
	if filepath.IsAbs(workDir) {
		return workDir
	}
	// 相对路径，基于 scripts 目录
	fullPath := filepath.Join(constant.ScriptsWorkDir, workDir)
	absPath, err := filepath.Abs(fullPath)
	if err != nil {
		return fullPath
	}
	return absPath
}

func (tc *TaskController) CreateTask(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Command     string `json:"command" binding:"required"`
		Schedule    string `json:"schedule" binding:"required"`
		Timeout     int    `json:"timeout"`
		WorkDir     string `json:"work_dir"`
		CleanConfig string `json:"clean_config"`
		Envs        string `json:"envs"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if err := tc.cronService.ValidateCron(req.Schedule); err != nil {
		utils.BadRequest(c, "无效的cron表达式: "+err.Error())
		return
	}

	// 转换为绝对路径
	workDir := resolveWorkDir(req.WorkDir)

	task := tc.taskService.CreateTask(req.Name, req.Command, req.Schedule, req.Timeout, workDir, req.CleanConfig, req.Envs)
	tc.cronService.AddTask(task)

	utils.Success(c, task)
}

func (tc *TaskController) GetTasks(c *gin.Context) {
	p := utils.ParsePagination(c)
	name := c.DefaultQuery("name", "")

	tasks, total := tc.taskService.GetTasksWithPagination(p.Page, p.PageSize, name)
	utils.PaginatedResponse(c, tasks, total, p)
}

func (tc *TaskController) GetTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	task := tc.taskService.GetTaskByID(id)
	if task == nil {
		utils.NotFound(c, "任务不存在")
		return
	}

	utils.Success(c, task)
}

func (tc *TaskController) UpdateTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Command     string `json:"command"`
		Schedule    string `json:"schedule"`
		Timeout     int    `json:"timeout"`
		WorkDir     string `json:"work_dir"`
		CleanConfig string `json:"clean_config"`
		Envs        string `json:"envs"`
		Enabled     bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	if req.Schedule != "" {
		if err := tc.cronService.ValidateCron(req.Schedule); err != nil {
			utils.BadRequest(c, "无效的cron表达式: "+err.Error())
			return
		}
	}

	task := tc.taskService.UpdateTask(id, req.Name, req.Command, req.Schedule, req.Timeout, resolveWorkDir(req.WorkDir), req.CleanConfig, req.Envs, req.Enabled)
	if task == nil {
		utils.NotFound(c, "任务不存在")
		return
	}

	if task.Enabled {
		tc.cronService.AddTask(task)
	} else {
		tc.cronService.RemoveTask(task.ID)
	}

	utils.Success(c, task)
}

func (tc *TaskController) DeleteTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.BadRequest(c, "无效的任务ID")
		return
	}

	tc.cronService.RemoveTask(uint(id))

	success := tc.taskService.DeleteTask(id)
	if !success {
		utils.NotFound(c, "任务不存在")
		return
	}

	utils.SuccessMsg(c, "删除成功")
}
