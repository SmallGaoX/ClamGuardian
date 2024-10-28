package monitor

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	"ClamGuardian/internal/matcher"
	"ClamGuardian/internal/position"
	"github.com/fsnotify/fsnotify"
)

// Monitor 文件监控器
type Monitor struct {
	watcher    *fsnotify.Watcher
	paths      []string
	patterns   []string
	matcher    *matcher.Matcher
	posManager *position.Manager
	bufferSize int
	mu         sync.Mutex
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
				return
			}
			m.handleEvent(event)

		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("监控错误: %v\n", err)

		case <-ctx.Done():
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
	// 获取当前文件位置
	currentPos := m.posManager.GetPosition(filename)

	// 处理文件内容
	newPos, err := m.matcher.ProcessFile(filename, currentPos)
	if err != nil {
		fmt.Printf("处理文件失败 %s: %v\n", filename, err)
		return
	}

	// 更新文件位置
	m.posManager.UpdatePosition(filename, newPos)
}

// handleFileCreate 处理文件创建事件
func (m *Monitor) handleFileCreate(filename string) {
	// 对于新创建的文件，从头开始读取
	newPos, err := m.matcher.ProcessFile(filename, 0)
	if err != nil {
		fmt.Printf("处理新文件失败 %s: %v\n", filename, err)
		return
	}

	// 更新文件位置
	m.posManager.UpdatePosition(filename, newPos)

	// 添加到监控列表
	if err := m.watcher.Add(filename); err != nil {
		fmt.Printf("添加文件到监控失败 %s: %v\n", filename, err)
	}
}

// handleFileRemove 处理文件删除事件
func (m *Monitor) handleFileRemove(filename string) {
	// 从监控列表中移除
	m.watcher.Remove(filename)

	// 从位置管理器中移除记录
	m.posManager.RemovePosition(filename)
}

// Stop 停止监控
func (m *Monitor) Stop() error {
	return m.watcher.Close()
}
