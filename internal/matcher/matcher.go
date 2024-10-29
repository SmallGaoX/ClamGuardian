package matcher

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sync"

	"ClamGuardian/internal/metrics"
)

// MatchRule 定义匹配规则的结构
type MatchRule struct {
	Pattern string `mapstructure:"pattern"`
	Level   string `mapstructure:"level"`
}

// Rule 内部使用的规则结构
type Rule struct {
	Pattern *regexp.Regexp
	Level   string
}

// Matcher 正则匹配器
type Matcher struct {
	rules      []Rule
	bufferSize int
	matchCount int64
	mu         sync.RWMutex
}

// NewMatcher 创建新的匹配器
func NewMatcher(rules []MatchRule, bufferSize int) (*Matcher, error) {
	var compiledRules []Rule
	for _, r := range rules {
		pattern, err := regexp.Compile(r.Pattern)
		if err != nil {
			return nil, fmt.Errorf("编译正则表达式失败 %s: %v", r.Pattern, err)
		}
		compiledRules = append(compiledRules, Rule{
			Pattern: pattern,
			Level:   r.Level,
		})
	}

	return &Matcher{
		rules:      compiledRules,
		bufferSize: bufferSize,
	}, nil
}

// ProcessFile 处理文件内容
func (m *Matcher) ProcessFile(filename string, offset int64) (int64, error) {
	file, err := os.Open(filename)
	if err != nil {
		return offset, fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	if _, err := file.Seek(offset, 0); err != nil {
		return offset, fmt.Errorf("设置文件偏移量失败: %v", err)
	}

	scanner := bufio.NewScanner(file)
	buf := make([]byte, m.bufferSize)
	scanner.Buffer(buf, m.bufferSize)

	var newOffset int64
	for scanner.Scan() {
		line := scanner.Text()
		m.matchLine(line)
		newOffset, _ = file.Seek(0, 1)
	}

	if err := scanner.Err(); err != nil {
		return offset, fmt.Errorf("扫描文件失败: %v", err)
	}

	return newOffset, nil
}

// GetMatchCount 获取总匹配次数
func (m *Matcher) GetMatchCount() int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.matchCount
}

// matchLine 匹配单行内容
func (m *Matcher) matchLine(line string) {
	for _, rule := range m.rules {
		if rule.Pattern.MatchString(line) {
			m.mu.Lock()
			m.matchCount++
			m.mu.Unlock()

			metrics.RuleMatches.WithLabelValues(rule.Level).Inc()
			// logger.Logger.Info("匹配到告警",
			// 	zap.String("level", rule.Level),
			// 	zap.String("content", line))
		}
	}
}
