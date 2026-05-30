package server

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ludoux/ngapost2md/config"
	"github.com/ludoux/ngapost2md/nga"
	"github.com/spf13/cast"
	"gopkg.in/ini.v1"
)

type APIHandler struct {
	taskManager     *TaskManager
	scheduleManager *ScheduleManager
	cfg             *ini.File
	password        string
	sessionManager  *SessionManager
}

func NewAPIHandler(taskManager *TaskManager, scheduleManager *ScheduleManager, cfg *ini.File, password string, sessionManager *SessionManager) *APIHandler {
	return &APIHandler{
		taskManager:     taskManager,
		scheduleManager: scheduleManager,
		cfg:             cfg,
		password:        password,
		sessionManager:  sessionManager,
	}
}

// GET /api/version - 返回版本号
func (h *APIHandler) HandleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"version": nga.VERSION})
}

// POST /api/login - 登录
func (h *APIHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "无效的请求体: "+err.Error())
		return
	}

	if req.Username != "admin" || req.Password != h.password {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "用户名或密码错误"})
		return
	}

	session := h.sessionManager.CreateSession()
	h.sessionManager.SetSessionCookie(w, session.Token)

	writeJSON(w, http.StatusOK, map[string]string{"message": "登录成功"})
}

// POST /api/logout - 登出
func (h *APIHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	session := h.sessionManager.GetSessionFromRequest(r)
	if session != nil {
		h.sessionManager.DeleteSession(session.Token)
	}
	h.sessionManager.ClearSessionCookie(w)

	writeJSON(w, http.StatusOK, map[string]string{"message": "已登出"})
}

func writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}

func readBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("读取请求体失败: %v", err)
	}
	return body, nil
}

// POST /api/download - 下载帖子
func (h *APIHandler) HandleDownload(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		Tid      int `json:"tid"`
		AuthorId int `json:"authorId"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "无效的请求体: "+err.Error())
		return
	}

	if req.Tid == 0 {
		writeError(w, http.StatusBadRequest, "tid 不能为空或0")
		return
	}

	task, err := h.taskManager.EnqueueDownload(req.Tid, req.AuthorId, nil)
	if err != nil {
		if strings.Contains(err.Error(), "已在队列") || strings.Contains(err.Error(), "正在执行") {
			writeError(w, http.StatusConflict, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusAccepted, task)
}

// POST /api/update - 更新帖子
func (h *APIHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		Tid int `json:"tid"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "无效的请求体: "+err.Error())
		return
	}

	if req.Tid == 0 {
		writeError(w, http.StatusBadRequest, "tid 不能为空或0")
		return
	}

	task, err := h.taskManager.EnqueueUpdate(req.Tid, nil)
	if err != nil {
		if strings.Contains(err.Error(), "已在队列") || strings.Contains(err.Error(), "正在执行") {
			writeError(w, http.StatusConflict, err.Error())
		} else {
			writeError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusAccepted, task)
}

// GET /api/tasks - 获取任务列表
func (h *APIHandler) HandleTasks(w http.ResponseWriter, r *http.Request) {
	tasks := h.taskManager.GetAllTasks()
	writeJSON(w, http.StatusOK, tasks)
}

// DELETE /api/tasks/{tid} - 取消任务
func (h *APIHandler) HandleTaskCancel(w http.ResponseWriter, r *http.Request) {
	tidStr := r.PathValue("tid")
	tid := cast.ToInt(tidStr)
	if tid == 0 {
		writeError(w, http.StatusBadRequest, "无效的 tid")
		return
	}

	if err := h.taskManager.CancelTask(tid); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "任务已取消"})
}

// GET /api/posts - 获取帖子列表
func (h *APIHandler) HandlePosts(w http.ResponseWriter, r *http.Request) {
	posts := ScanAllPosts()
	writeJSON(w, http.StatusOK, posts)
}

// DELETE /api/posts/{tid} - 删除帖子文件夹
func (h *APIHandler) HandlePostDelete(w http.ResponseWriter, r *http.Request) {
	tidStr := r.PathValue("tid")
	tid := cast.ToInt(tidStr)
	if tid == 0 {
		writeError(w, http.StatusBadRequest, "无效的 tid")
		return
	}

	folderName := findPostFolder(tid)
	if folderName == "" {
		writeError(w, http.StatusNotFound, "未找到对应帖子文件夹")
		return
	}

	folderPath := filepath.Join(nga.CFGFILE_OUTPUT_PATH, folderName)
	if err := os.RemoveAll(folderPath); err != nil {
		writeError(w, http.StatusInternalServerError, "删除文件夹失败: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "帖子文件夹已删除"})
}

// GET /api/posts/{tid}/download - 下载帖子文件
func (h *APIHandler) HandlePostDownload(w http.ResponseWriter, r *http.Request) {
	tidStr := r.PathValue("tid")
	tid := cast.ToInt(tidStr)
	if tid == 0 {
		writeError(w, http.StatusBadRequest, "无效的 tid")
		return
	}

	// 查找匹配 tid 的文件夹
	folderName := findPostFolder(tid)
	if folderName == "" {
		writeError(w, http.StatusNotFound, "未找到对应帖子文件夹")
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, folderName))

	zw := zip.NewWriter(w)
	defer zw.Close()

	zipBasePath := filepath.Join(nga.CFGFILE_OUTPUT_PATH, folderName)
	filepath.WalkDir(zipBasePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		// 排除 . 开头的文件和目录
		if strings.HasPrefix(filepath.Base(path), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		// 根目录本身跳过
		if path == zipBasePath {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return nil
		}
		// zip 内使用相对路径（相对于 output 目录），如 "folderName/file.md"
		relPath, _ := filepath.Rel(nga.CFGFILE_OUTPUT_PATH, path)
		header.Name = strings.ReplaceAll(relPath, `\`, `/`)
		header.Method = zip.Deflate

		if d.IsDir() {
			header.Name += "/"
			_, err := zw.CreateHeader(header)
			return err
		}

		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		io.Copy(writer, f)
		return nil
	})
}

// findPostFolder 查找匹配 tid 的文件夹名
func findPostFolder(tid int) string {
	allPosts := ScanAllPosts()
	for _, p := range allPosts {
		if p.Tid == tid {
			return p.FolderName
		}
	}
	return ""
}

// GET /api/schedules - 获取定时任务列表
func (h *APIHandler) HandleSchedulesGet(w http.ResponseWriter, r *http.Request) {
	schedules := h.scheduleManager.GetAllSchedules()
	writeJSON(w, http.StatusOK, schedules)
}

// POST /api/schedules - 创建定时任务
func (h *APIHandler) HandleSchedulesCreate(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		Tid      int    `json:"tid"`
		AuthorId int    `json:"authorId"`
		Cron     string `json:"cron"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "无效的请求体: "+err.Error())
		return
	}

	if req.Tid == 0 {
		writeError(w, http.StatusBadRequest, "tid 不能为空或0")
		return
	}
	if req.Cron == "" {
		writeError(w, http.StatusBadRequest, "cron 表达式不能为空")
		return
	}

	schedule, err := h.scheduleManager.AddSchedule(req.Tid, req.AuthorId, req.Cron)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, schedule)
}

// PUT /api/schedules/{id} - 更新定时任务
func (h *APIHandler) HandleSchedulesUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	body, err := readBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req struct {
		Cron    string `json:"cron,omitempty"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, http.StatusBadRequest, "无效的请求体: "+err.Error())
		return
	}

	schedule, err := h.scheduleManager.UpdateSchedule(id, req.Cron, req.Enabled)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, schedule)
}

// DELETE /api/schedules/{id} - 删除定时任务
func (h *APIHandler) HandleSchedulesDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := h.scheduleManager.DeleteSchedule(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "定时任务已删除"})
}

// GET /api/config - 获取配置
func (h *APIHandler) HandleConfigGet(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.GetConfigAutoUpdate()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	result := map[string]interface{}{}
	comments := map[string]map[string]string{}
	for _, section := range cfg.Sections() {
		secName := section.Name()
		if secName == "DEFAULT" {
			continue
		}
		secMap := map[string]string{}
		secComments := map[string]string{}
		for _, key := range section.Keys() {
			if secName == "server" && key.Name() == "password" {
				continue
			}
			secMap[key.Name()] = key.Value()
			if key.Comment != "" {
				secComments[key.Name()] = strings.TrimSpace(key.Comment)
			}
		}
		result[secName] = secMap
		if len(secComments) > 0 {
			comments[secName] = secComments
		}
	}

	if len(comments) > 0 {
		result["_comments"] = comments
	}

	writeJSON(w, http.StatusOK, result)
}

// PUT /api/config - 更新配置
func (h *APIHandler) HandleConfigPut(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	var updates map[string]map[string]string
	if err := json.Unmarshal(body, &updates); err != nil {
		writeError(w, http.StatusBadRequest, "无效的请求体: "+err.Error())
		return
	}

	// 不允许通过 API 修改 server password
	if serverUpdates, ok := updates["server"]; ok {
		if _, exists := serverUpdates["password"]; exists {
			writeError(w, http.StatusBadRequest, "不允许通过 API 修改 server.password，请通过命令行参数或手动编辑 config.ini 修改")
			return
		}
	}

	// 不允许通过 API 修改 config version，静默忽略
	if configUpdates, ok := updates["config"]; ok {
		delete(configUpdates, "version")
	}

	cfg, err := config.GetConfigAutoUpdate()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// 应用更新
	for secName, secUpdates := range updates {
		if !cfg.HasSection(secName) {
			cfg.NewSection(secName)
		}
		for key, value := range secUpdates {
			cfg.Section(secName).Key(key).SetValue(value)
		}
	}

	// 保存
	if err := cfg.SaveTo("config.ini"); err != nil {
		writeError(w, http.StatusInternalServerError, "保存配置文件失败: "+err.Error())
		return
	}

	// 重新加载配置
	if err := ReloadConfig(); err != nil {
		writeError(w, http.StatusInternalServerError, "重新加载配置失败: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "配置已保存并生效"})
}

// GET /ws - WebSocket 连接
func (h *APIHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	HandleWebSocket(w, r, h.password, h.taskManager.hub, h.sessionManager)
}
