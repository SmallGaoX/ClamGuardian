package monitor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"ClamGuardian/internal/logger"
	"ClamGuardian/internal/matcher"
	"ClamGuardian/internal/metrics"
	"ClamGuardian/internal/position"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

// Monitor 文件监控器
type Monitor struct {
	watcher    *fsnotify.Watcher
	paths      []string
	patterns   []string
	matcher    *matcher.Matcher
	posManager *position.Manager
	bufferSize int
	mu         sync.RWMutex
	fileCount  int
}

// NewMonitor 创建新的监控器
func NewMonitor(paths []string, patterns []string, m *matcher.Matcher, pm *position.Manager, bufferSize int) (*Monitor, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建监控器失败: %v", err)
	}

	return &Monitor{
		watcher:    w,
		paths:      paths,
		patterns:   patterns,
		matcher:    m,
		posManager: pm,
		bufferSize: bufferSize,
	}, nil
}

// Start 开始监控
func (m *Monitor) Start(ctx context.Context) error {
	// 添加所有目录到监控
	for _, path := range m.paths {
		if err := m.watcher.Add(path); err != nil {
			return fmt.Errorf("添加监控路径失败 %s: %v", path, err)
		}
	}

	go m.watch(ctx)
	return nil
}

// watch 监控文件变化
func (m *Monitor) watch(ctx context.Context) {
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				logger.Logger.Info("监控事件通道已关闭")
				return
			}
			m.handleEvent(event)

		case err, ok := <-m.watcher.Errors:
			if !ok {
				logger.Logger.Info("监控错误通道已关闭")
				return
			}
			logger.Logger.Error("监控错误", zap.Error(err))

		case <-ctx.Done():
			logger.Logger.Info("监控服务停止")
			return
		}
	}
}

// handleEvent 处理文件事件
func (m *Monitor) handleEvent(event fsnotify.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 检查文件是否匹配模式
	matched := false
	for _, pattern := range m.patterns {
		if match, _ := filepath.Match(pattern, filepath.Base(event.Name)); match {
			matched = true
			break
		}
	}

	if !matched {
		return
	}

	switch {
	case event.Op&fsnotify.Write == fsnotify.Write:
		m.handleFileWrite(event.Name)
	case event.Op&fsnotify.Create == fsnotify.Create:
		m.handleFileCreate(event.Name)
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		m.handleFileRemove(event.Name)
	}
}

// handleFileWrite 处理文件写入事件
func (m *Monitor) handleFileWrite(filename string) {
	currentPos := m.posManager.GetPosition(filename)
	fileInfo, err := os.Stat(filename)
	if err != nil {
		logger.Logger.Error("获取文件信息失败", zap.Error(err))
		return
	}

	newPos, err := m.matcher.ProcessFile(filename, currentPos)
	if err != nil {
		logger.Logger.Error("处理文件失败", zap.Error(err))
		return
	}

	// 更新状态管理器
	stateManager := metrics.GetStateManager()
	stateManager.UpdateFileStatus(filename, &metrics.FileStatus{
		Filename:     filename,
		Position:     newPos,
		Size:         fileInfo.Size(),
		Progress:     float64(newPos) / float64(fileInfo.Size()),
		LastModified: fileInfo.ModTime(),
	})

	m.posManager.UpdatePosition(filename, newPos)
}

// handleFileCreate 处理文件创建事件
func (m *Monitor) handleFileCreate(filename string) {
	m.mu.Lock()
	m.fileCount++
	m.mu.Unlock()

	logger.Logger.Info("检测到新文件",
		zap.String("filename", filename))

	newPos, err := m.matcher.ProcessFile(filename, 0)
	if err != nil {
		logger.Logger.Error("处理新文件失败",
			zap.String("filename", filename),
			zap.Error(err))
		return
	}

	m.posManager.UpdatePosition(filename, newPos)
	if err := m.watcher.Add(filename); err != nil {
		logger.Logger.Error("添加文件到监控失败",
			zap.String("filename", filename),
			zap.Error(err))
	}
	metrics.ProcessedFiles.Inc()
}

// handleFileRemove 处理文件删除事件
func (m *Monitor) handleFileRemove(filename string) {
	m.mu.Lock()
	m.fileCount--
	m.mu.Unlock()

	logger.Logger.Info("文件被删除",
		zap.String("filename", filename))
	m.watcher.Remove(filename)
	m.posManager.RemovePosition(filename)
}

// Stop 停止监控
func (m *Monitor) Stop() error {
	return m.watcher.Close()
}

// GetFileCount 获取当前监控的文件数
func (m *Monitor) GetFileCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.fileCount
}
