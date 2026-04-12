package main

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

type JobTypeConfig struct {
	MaxRetries int32
}

type Config struct {
	mu                sync.RWMutex
	jobTypeConfig     map[string]JobTypeConfig
	defaultMaxRetries int32
}

var globalConfig *Config

func init() {
	globalConfig = &Config{
		jobTypeConfig:     make(map[string]JobTypeConfig),
		defaultMaxRetries: 3,
	}
	loadJobTypeConfigs()
}

func loadJobTypeConfigs() {
	envVal := os.Getenv("JOB_TYPE_MAX_RETRIES")
	if envVal == "" {
		return
	}

	pairs := strings.Split(envVal, ",")
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) != 2 {
			continue
		}
		jobType := strings.TrimSpace(parts[0])
		maxRetries, err := strconv.Atoi(strings.TrimSpace(parts[1]))
		if err != nil || maxRetries < 0 {
			continue
		}
		globalConfig.mu.Lock()
		globalConfig.jobTypeConfig[jobType] = JobTypeConfig{
			MaxRetries: int32(maxRetries),
		}
		globalConfig.mu.Unlock()
	}
}

func GetMaxRetriesForJobType(jobType string) int32 {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()

	if cfg, ok := globalConfig.jobTypeConfig[jobType]; ok {
		return cfg.MaxRetries
	}
	return globalConfig.defaultMaxRetries
}

func SetDefaultMaxRetries(maxRetries int32) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.defaultMaxRetries = maxRetries
}
