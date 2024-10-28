package position

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"ClamGuardian/internal/logger"
	"go.uber.org/zap"
)

// Manager 位置管理器
type Manager struct {
	positions   map[string]int64
	storePath   string
	mu          sync.RWMutex
	updateTimer *time.Timer
}

// NewManager 创建新的位置管理器
func NewManager(storePath string, updateInterval int) (*Manager, error) {
	m := &Manager{
		positions: make(map[string]int64),
		storePath: storePath,
	}

	if err := m.load(); err != nil {
		return nil, err
	}

	m.updateTimer = time.NewTimer(time.Duration(updateInterval) * time.Second)
	go m.periodicUpdate()

	return m, nil
}

// GetPosition 获取文件位置
func (m *Manager) GetPosition(filename string) int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.positions[filename]
}

// UpdatePosition 更新文件位置
func (m *Manager) UpdatePosition(filename string, position int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.positions[filename] = position
}

// load 从磁盘加载位置信息
func (m *Manager) load() error {
	data, err := os.ReadFile(m.storePath)
	if os.IsNotExist(err) {
		logger.Logger.Info("位置文件不存在，将创建新文件",
			zap.String("path", m.storePath))
		return nil
	}
	if err != nil {
		logger.Logger.Error("读取位置文件失败",
			zap.String("path", m.storePath),
			zap.Error(err))
		return fmt.Errorf("读取位置文件失败: %v", err)
	}

	if err := json.Unmarshal(data, &m.positions); err != nil {
		logger.Logger.Error("解析位置文件失败", zap.Error(err))
		return err
	}

	logger.Logger.Info("成功加载位置信息",
		zap.Int("文件数", len(m.positions)))
	return nil
}

// save 保存位置信息到磁盘
func (m *Manager) save() error {
	m.mu.RLock()
	data, err := json.Marshal(m.positions)
	m.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("序列化位置信息失败: %v", err)
	}

	return os.WriteFile(m.storePath, data, 0644)
}

// periodicUpdate 定期更新位置信息
func (m *Manager) periodicUpdate() {
	for range m.updateTimer.C {
		if err := m.save(); err != nil {
			logger.Logger.Error("保存位置信息失败", zap.Error(err))
		} else {
			logger.Logger.Debug("位置信息已保存")
		}
		m.updateTimer.Reset(time.Duration(5) * time.Second)
	}
}

// RemovePosition 移除文件位置记录
func (m *Manager) RemovePosition(filename string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.positions, filename)
}
