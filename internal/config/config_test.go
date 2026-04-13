package config

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestGetMaxRetriesForJobType(t *testing.T) {
	SetDefaultMaxRetries(3)

	retries := GetMaxRetriesForJobType("email")
	if retries != 3 {
		t.Errorf("expected default 3, got %d", retries)
	}
}

func TestConfigConcurrentAccess(t *testing.T) {
	testConfig := &Config{
		jobTypeConfig:     make(map[string]JobTypeConfig),
		defaultMaxRetries: 3,
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = testConfig.getMaxRetries("email")
		}()
	}
	wg.Wait()
}

func TestLoadJobTypeConfigsFromEnv(t *testing.T) {
	os.Setenv("JOB_TYPE_MAX_RETRIES", "email:5,sms:3,push_notification:2")
	defer os.Unsetenv("JOB_TYPE_MAX_RETRIES")

	testConfig := &Config{
		jobTypeConfig:     make(map[string]JobTypeConfig),
		defaultMaxRetries: 3,
	}
	testConfig.loadFromEnv()

	if testConfig.jobTypeConfig["email"].MaxRetries != 5 {
		t.Errorf("expected email max_retries 5, got %d", testConfig.jobTypeConfig["email"].MaxRetries)
	}
	if testConfig.jobTypeConfig["sms"].MaxRetries != 3 {
		t.Errorf("expected sms max_retries 3, got %d", testConfig.jobTypeConfig["sms"].MaxRetries)
	}
	if testConfig.jobTypeConfig["push_notification"].MaxRetries != 2 {
		t.Errorf("expected push_notification max_retries 2, got %d", testConfig.jobTypeConfig["push_notification"].MaxRetries)
	}
}

func (c *Config) getMaxRetries(jobType string) int32 {
	if cfg, ok := c.jobTypeConfig[jobType]; ok {
		return cfg.MaxRetries
	}
	return c.defaultMaxRetries
}

func (c *Config) loadFromEnv() {
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
		c.mu.Lock()
		c.jobTypeConfig[jobType] = JobTypeConfig{
			MaxRetries: int32(maxRetries),
		}
		c.mu.Unlock()
	}
}
