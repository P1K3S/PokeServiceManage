package frp

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"service-manage/config"
	"service-manage/logger"
	sshutil "service-manage/utils/ssh"
)

type FrpServerConfig struct {
	ServerPort int
	AuthToken  string
}

type DiscoverParams struct {
	MachineIP       string
	SSHPort         int
	SSHUser         string
	SSHPassword     string
	ContainerName   string
}

var (
	cachedConfig *FrpServerConfig
	cacheMu      sync.RWMutex
	cacheTime    time.Time
	cacheTTL     = 5 * time.Minute
)

func discoverFromContainer(p *DiscoverParams) (*FrpServerConfig, error) {
	sshPort := p.SSHPort
	if sshPort == 0 {
		sshPort = config.AppConfig.SSH.DefaultPort
	}

	sshCfg := &sshutil.Config{
		Host:     p.MachineIP,
		Port:     sshPort,
		User:     p.SSHUser,
		Password: p.SSHPassword,
	}

	inspectCmd := fmt.Sprintf(
		"docker inspect %s --format='{{range .Mounts}}{{.Type}}:{{.Source}}:{{.Destination}}{{println}}{{end}}'",
		p.ContainerName,
	)
	output, err := sshutil.RunCommand(sshCfg, inspectCmd)
	if err != nil {
		return nil, fmt.Errorf("docker inspect 失败: %w", err)
	}

	configPath := findFrpsConfigPath(output)
	if configPath == "" {
		return nil, fmt.Errorf("未找到 frps 配置文件挂载，docker inspect 输出: %s", strings.TrimSpace(output))
	}

	content, err := sshutil.RunCommand(sshCfg, "cat "+configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件 %s 失败: %w", configPath, err)
	}

	result := parseFrpsConfig(content)
	if result.ServerPort <= 0 {
		return nil, fmt.Errorf("解析 frps 配置失败: bindPort 未找到，配置内容前200字符: %s", truncate(content, 200))
	}

	return result, nil
}

func findFrpsConfigPath(inspectOutput string) string {
	lines := strings.Split(inspectOutput, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		parts := strings.Split(line, ":")
		if len(parts) < 3 {
			continue
		}
		sourcePath := strings.Join(parts[1:len(parts)-1], ":")
		destPath := parts[len(parts)-1]
		if isFrpsConfig(destPath) || isFrpsConfig(sourcePath) {
			return sourcePath
		}
	}
	return ""
}

func isFrpsConfig(path string) bool {
	lower := strings.ToLower(path)
	return strings.Contains(lower, "frps") &&
		(strings.HasSuffix(lower, ".toml") || strings.HasSuffix(lower, ".ini") || strings.HasSuffix(lower, ".yaml") || strings.HasSuffix(lower, ".yml"))
}

func parseFrpsConfig(content string) *FrpServerConfig {
	result := &FrpServerConfig{}

	portRe := regexp.MustCompile(`(?i)bindPort\s*=\s*(\d+)`)
	if m := portRe.FindStringSubmatch(content); len(m) > 1 {
		result.ServerPort, _ = strconv.Atoi(m[1])
	}

	tokenRe := regexp.MustCompile(`(?i)auth\.token\s*=\s*"([^"]*)"`)
	if m := tokenRe.FindStringSubmatch(content); len(m) > 1 {
		result.AuthToken = m[1]
	}

	return result
}

func fallbackFromConfig() *FrpServerConfig {
	cfg := config.AppConfig.Frp
	if cfg.ServerPort > 0 && cfg.AuthToken != "" {
		return &FrpServerConfig{
			ServerPort: cfg.ServerPort,
			AuthToken:  cfg.AuthToken,
		}
	}
	return nil
}

func GetServerConfig(p *DiscoverParams) *FrpServerConfig {
	cacheKey := fmt.Sprintf("%s:%d:%s", p.MachineIP, p.SSHPort, p.ContainerName)
	cacheMu.RLock()
	if cachedConfig != nil && time.Since(cacheTime) < cacheTTL && cacheKey == currentCacheKey {
		result := cachedConfig
		cacheMu.RUnlock()
		return result
	}
	cacheMu.RUnlock()

	cacheMu.Lock()
	defer cacheMu.Unlock()

	if cachedConfig != nil && time.Since(cacheTime) < cacheTTL && cacheKey == currentCacheKey {
		return cachedConfig
	}

	discovered, err := discoverFromContainer(p)
	if err != nil {
		logger.Log.Sugar().Warnf("FRP配置自动发现失败(容器%s@%s): %v，尝试配置文件兜底", p.ContainerName, p.MachineIP, err)
		if fb := fallbackFromConfig(); fb != nil {
			cachedConfig = fb
			cacheTime = time.Now()
			currentCacheKey = cacheKey
			return fb
		}
		if cachedConfig != nil {
			return cachedConfig
		}
		return &FrpServerConfig{}
	}

	cachedConfig = discovered
	cacheTime = time.Now()
	currentCacheKey = cacheKey
	logger.Log.Sugar().Infof("FRP配置自动发现成功(容器%s@%s): bindPort=%d", p.ContainerName, p.MachineIP, discovered.ServerPort)
	return discovered
}

func RefreshConfig(p *DiscoverParams) {
	cacheMu.Lock()
	cachedConfig = nil
	cacheTime = time.Time{}
	currentCacheKey = ""
	cacheMu.Unlock()

	GetServerConfig(p)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func ParseDockerInspectMounts(output string) []map[string]string {
	var results []map[string]string
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) >= 3 {
			results = append(results, map[string]string{
				"type":   parts[0],
				"source": strings.Join(parts[1:len(parts)-1], ":"),
				"dest":   parts[len(parts)-1],
			})
		}
	}
	return results
}

func ParseContainerJSON(data string) (string, []map[string]string, error) {
	var container struct {
		Name   string `json:"Name"`
		Mounts []struct {
			Type   string `json:"Type"`
			Source string `json:"Source"`
			Dest   string `json:"Destination"`
		} `json:"Mounts"`
	}
	if err := json.Unmarshal([]byte(data), &container); err != nil {
		return "", nil, err
	}
	name := strings.TrimPrefix(container.Name, "/")
	var mounts []map[string]string
	for _, m := range container.Mounts {
		mounts = append(mounts, map[string]string{
			"type":   m.Type,
			"source": m.Source,
			"dest":   m.Dest,
		})
	}
	return name, mounts, nil
}

var currentCacheKey string
