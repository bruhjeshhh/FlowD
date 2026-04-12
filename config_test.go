package main

import (
	"os"
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
	globalConfig = &Config{
		jobTypeConfig:     make(map[string]JobTypeConfig),
		defaultMaxRetries: 3,
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			GetMaxRetriesForJobType("email")
		}()
	}
	wg.Wait()
}

func TestLoadJobTypeConfigsFromEnv(t *testing.T) {
	os.Setenv("JOB_TYPE_MAX_RETRIES", "email:5,sms:3,push_notification:2")
	defer os.Unsetenv("JOB_TYPE_MAX_RETRIES")

	globalConfig = &Config{
		jobTypeConfig:     make(map[string]JobTypeConfig),
		defaultMaxRetries: 3,
	}
	loadJobTypeConfigs()

	if globalConfig.jobTypeConfig["email"].MaxRetries != 5 {
		t.Errorf("expected email max_retries 5, got %d", globalConfig.jobTypeConfig["email"].MaxRetries)
	}
	if globalConfig.jobTypeConfig["sms"].MaxRetries != 3 {
		t.Errorf("expected sms max_retries 3, got %d", globalConfig.jobTypeConfig["sms"].MaxRetries)
	}
	if globalConfig.jobTypeConfig["push_notification"].MaxRetries != 2 {
		t.Errorf("expected push_notification max_retries 2, got %d", globalConfig.jobTypeConfig["push_notification"].MaxRetries)
	}
}
