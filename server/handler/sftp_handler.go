package handler

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"service-manage/config"
	"service-manage/model"
	sshutil "service-manage/utils/ssh"

	"github.com/gin-gonic/gin"
	"github.com/pkg/sftp"
	"gorm.io/gorm"
)

type SFTPHandler struct {
	DB *gorm.DB
}

func NewSFTPHandler(db *gorm.DB) *SFTPHandler {
	return &SFTPHandler{DB: db}
}

func (h *SFTPHandler) getSFTPClient(c *gin.Context) (*sftp.Client, func(), error) {
	id, _ := strconv.Atoi(c.Param("id"))
	var machine model.Machine
	if err := userScope(c, h.DB).First(&machine, id).Error; err != nil {
		return nil, nil, fmt.Errorf("主机不存在")
	}
	if !machine.SSHEnabled || machine.SSHUser == "" || machine.SSHPassword == "" {
		return nil, nil, fmt.Errorf("该主机未配置SSH连接信息")
	}

	sshPort := machine.SSHPort
	if sshPort == 0 {
		sshPort = config.AppConfig.SSH.DefaultPort
	}

	sshClient, err := sshutil.NewClient(&sshutil.Config{
		Host:     machine.IP,
		Port:     sshPort,
		User:     machine.SSHUser,
		Password: machine.SSHPassword,
		Timeout:  time.Duration(config.AppConfig.SSH.Timeout) * time.Second,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("SSH连接失败: %v", err)
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, nil, fmt.Errorf("SFTP会话创建失败: %v", err)
	}

	cleanup := func() {
		sftpClient.Close()
		sshClient.Close()
	}

	return sftpClient, cleanup, nil
}

func (h *SFTPHandler) List(c *gin.Context) {
	client, cleanup, err := h.getSFTPClient(c)
	if err != nil {
		jsonError(c, err.Error())
		return
	}
	defer cleanup()

	dirPath := c.DefaultQuery("path", "/")
	showHidden := c.DefaultQuery("hidden", "false") == "true"

	entries, err := client.ReadDir(dirPath)
	if err != nil {
		jsonError(c, fmt.Sprintf("读取目录失败: %v", err))
		return
	}

	type FileEntry struct {
		Name    string `json:"name"`
		Size    int64  `json:"size"`
		Mode    string `json:"mode"`
		ModTime string `json:"modTime"`
		IsDir   bool   `json:"isDir"`
		Path    string `json:"path"`
	}

	var files []FileEntry
	for _, entry := range entries {
		name := entry.Name()
		if !showHidden && strings.HasPrefix(name, ".") {
			continue
		}
		fullPath := path.Join(dirPath, name)
		files = append(files, FileEntry{
			Name:    name,
			Size:    entry.Size(),
			Mode:    entry.Mode().String(),
			ModTime: entry.ModTime().Format("2006-01-02 15:04:05"),
			IsDir:   entry.IsDir(),
			Path:    fullPath,
		})
	}

	jsonSuccess(c, gin.H{
		"path":  dirPath,
		"files": files,
	})
}

func (h *SFTPHandler) Download(c *gin.Context) {
	client, cleanup, err := h.getSFTPClient(c)
	if err != nil {
		jsonError(c, err.Error())
		return
	}
	defer cleanup()

	filePath := c.Query("path")
	if filePath == "" {
		jsonError(c, "文件路径不能为空")
		return
	}

	stat, err := client.Stat(filePath)
	if err != nil {
		jsonError(c, fmt.Sprintf("文件不存在: %v", err))
		return
	}

	if stat.IsDir() {
		jsonError(c, "不能下载目录")
		return
	}

	remoteFile, err := client.Open(filePath)
	if err != nil {
		jsonError(c, fmt.Sprintf("打开文件失败: %v", err))
		return
	}
	defer remoteFile.Close()

	fileName := path.Base(filePath)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", strconv.FormatInt(stat.Size(), 10))

	io.Copy(c.Writer, remoteFile)
}

func (h *SFTPHandler) Upload(c *gin.Context) {
	client, cleanup, err := h.getSFTPClient(c)
	if err != nil {
		jsonError(c, err.Error())
		return
	}
	defer cleanup()

	dirPath := c.DefaultQuery("path", "/")

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		jsonError(c, fmt.Sprintf("获取上传文件失败: %v", err))
		return
	}
	defer file.Close()

	remotePath := path.Join(dirPath, header.Filename)
	dst, err := client.Create(remotePath)
	if err != nil {
		jsonError(c, fmt.Sprintf("创建远程文件失败: %v", err))
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		jsonError(c, fmt.Sprintf("写入文件失败: %v", err))
		return
	}

	uid, uname := getLogUserInfo(c)
	logOperation(h.DB, uid, uname, "upload", "sftp", 0, fmt.Sprintf("%s (%d bytes)", remotePath, written))

	jsonSuccess(c, gin.H{
		"path": remotePath,
		"size": written,
	})
}

func (h *SFTPHandler) Mkdir(c *gin.Context) {
	client, cleanup, err := h.getSFTPClient(c)
	if err != nil {
		jsonError(c, err.Error())
		return
	}
	defer cleanup()

	var req struct {
		Path string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Path == "" {
		jsonError(c, "路径不能为空")
		return
	}

	if err := client.Mkdir(req.Path); err != nil {
		jsonError(c, fmt.Sprintf("创建目录失败: %v", err))
		return
	}

	jsonSuccess(c, gin.H{"path": req.Path})
}

func (h *SFTPHandler) Remove(c *gin.Context) {
	client, cleanup, err := h.getSFTPClient(c)
	if err != nil {
		jsonError(c, err.Error())
		return
	}
	defer cleanup()

	var req struct {
		Path  string `json:"path"`
		IsDir bool   `json:"isDir"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Path == "" {
		jsonError(c, "路径不能为空")
		return
	}

	var removeErr error
	if req.IsDir {
		removeErr = client.RemoveDirectory(req.Path)
	} else {
		removeErr = client.Remove(req.Path)
	}

	if removeErr != nil {
		jsonError(c, fmt.Sprintf("删除失败: %v", removeErr))
		return
	}

	jsonSuccess(c, gin.H{"path": req.Path})
}

func (h *SFTPHandler) Rename(c *gin.Context) {
	client, cleanup, err := h.getSFTPClient(c)
	if err != nil {
		jsonError(c, err.Error())
		return
	}
	defer cleanup()

	var req struct {
		OldPath string `json:"oldPath"`
		NewPath string `json:"newPath"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.OldPath == "" || req.NewPath == "" {
		jsonError(c, "路径不能为空")
		return
	}

	if err := client.Rename(req.OldPath, req.NewPath); err != nil {
		jsonError(c, fmt.Sprintf("重命名失败: %v", err))
		return
	}

	jsonSuccess(c, gin.H{"oldPath": req.OldPath, "newPath": req.NewPath})
}

func (h *SFTPHandler) ReadFile(c *gin.Context) {
	client, cleanup, err := h.getSFTPClient(c)
	if err != nil {
		jsonError(c, err.Error())
		return
	}
	defer cleanup()

	filePath := c.Query("path")
	if filePath == "" {
		jsonError(c, "文件路径不能为空")
		return
	}

	stat, err := client.Stat(filePath)
	if err != nil {
		jsonError(c, fmt.Sprintf("文件不存在: %v", err))
		return
	}

	maxSize := int64(512 * 1024)
	if stat.Size() > maxSize {
		jsonError(c, fmt.Sprintf("文件过大（%d字节），最大支持 %d 字节", stat.Size(), maxSize))
		return
	}

	remoteFile, err := client.Open(filePath)
	if err != nil {
		jsonError(c, fmt.Sprintf("打开文件失败: %v", err))
		return
	}
	defer remoteFile.Close()

	content, err := io.ReadAll(remoteFile)
	if err != nil {
		jsonError(c, fmt.Sprintf("读取文件失败: %v", err))
		return
	}

	jsonSuccess(c, gin.H{
		"path":    filePath,
		"content": string(content),
		"size":    len(content),
	})
}

func (h *SFTPHandler) WriteFile(c *gin.Context) {
	client, cleanup, err := h.getSFTPClient(c)
	if err != nil {
		jsonError(c, err.Error())
		return
	}
	defer cleanup()

	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Path == "" {
		jsonError(c, "参数错误")
		return
	}

	dst, err := client.Create(req.Path)
	if err != nil {
		jsonError(c, fmt.Sprintf("创建文件失败: %v", err))
		return
	}
	defer dst.Close()

	written, err := dst.Write([]byte(req.Content))
	if err != nil {
		jsonError(c, fmt.Sprintf("写入文件失败: %v", err))
		return
	}

	jsonSuccess(c, gin.H{
		"path": req.Path,
		"size": written,
	})
}

func (h *SFTPHandler) Stat(c *gin.Context) {
	client, cleanup, err := h.getSFTPClient(c)
	if err != nil {
		jsonError(c, err.Error())
		return
	}
	defer cleanup()

	filePath := c.Query("path")
	if filePath == "" {
		jsonError(c, "路径不能为空")
		return
	}

	stat, err := client.Stat(filePath)
	if err != nil {
		jsonError(c, fmt.Sprintf("获取文件信息失败: %v", err))
		return
	}

	jsonSuccess(c, gin.H{
		"name":    stat.Name(),
		"size":    stat.Size(),
		"mode":    stat.Mode().String(),
		"modTime": stat.ModTime().Format("2006-01-02 15:04:05"),
		"isDir":   stat.IsDir(),
	})
}

func (h *SFTPHandler) DownloadDir(c *gin.Context) {
	client, cleanup, err := h.getSFTPClient(c)
	if err != nil {
		jsonError(c, err.Error())
		return
	}
	defer cleanup()

	dirPath := c.Query("path")
	if dirPath == "" {
		jsonError(c, "目录路径不能为空")
		return
	}

	stat, err := client.Stat(dirPath)
	if err != nil || !stat.IsDir() {
		jsonError(c, "目录不存在")
		return
	}

	tmpDir, err := os.MkdirTemp("", "sftp-download-*")
	if err != nil {
		jsonError(c, "创建临时目录失败")
		return
	}
	defer os.RemoveAll(tmpDir)

	localDir := filepath.Join(tmpDir, path.Base(dirPath))
	if err := h.downloadDirRecursive(client, dirPath, localDir); err != nil {
		jsonError(c, fmt.Sprintf("下载目录失败: %v", err))
		return
	}

	zipPath := tmpDir + ".zip"
	if err := h.zipDir(localDir, zipPath); err != nil {
		jsonError(c, fmt.Sprintf("压缩失败: %v", err))
		return
	}
	defer os.Remove(zipPath)

	fileName := path.Base(dirPath) + ".zip"
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	c.Header("Content-Type", "application/zip")
	http.ServeFile(c.Writer, c.Request, zipPath)
}

func (h *SFTPHandler) downloadDirRecursive(client *sftp.Client, remoteDir, localDir string) error {
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return err
	}

	entries, err := client.ReadDir(remoteDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		remotePath := path.Join(remoteDir, entry.Name())
		localPath := filepath.Join(localDir, entry.Name())

		if entry.IsDir() {
			if err := h.downloadDirRecursive(client, remotePath, localPath); err != nil {
				return err
			}
		} else {
			remoteFile, err := client.Open(remotePath)
			if err != nil {
				continue
			}
			localFile, err := os.Create(localPath)
			if err != nil {
				remoteFile.Close()
				continue
			}
			io.Copy(localFile, remoteFile)
			localFile.Close()
			remoteFile.Close()
		}
	}
	return nil
}

func (h *SFTPHandler) zipDir(srcDir, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zw := zip.NewWriter(zipFile)
	defer zw.Close()

	return filepath.Walk(srcDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(filepath.Dir(srcDir), filePath)
		if err != nil {
			return err
		}

		if info.IsDir() {
			_, err = zw.Create(relPath + "/")
			return err
		}

		w, err := zw.Create(relPath)
		if err != nil {
			return err
		}

		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(w, f)
		return err
	})
}
