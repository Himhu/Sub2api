package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type settingHandlerRepoStub struct {
	values      map[string]string
	lastUpdates map[string]string
}

func (s *settingHandlerRepoStub) Get(ctx context.Context, key string) (*service.Setting, error) {
	panic("unexpected Get call")
}

func (s *settingHandlerRepoStub) GetValue(ctx context.Context, key string) (string, error) {
	if s.values != nil {
		if value, ok := s.values[key]; ok {
			return value, nil
		}
	}
	return "", nil
}

func (s *settingHandlerRepoStub) Set(ctx context.Context, key, value string) error {
	panic("unexpected Set call")
}

func (s *settingHandlerRepoStub) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		if value, ok := s.values[key]; ok {
			out[key] = value
		}
	}
	return out, nil
}

func (s *settingHandlerRepoStub) SetMultiple(ctx context.Context, settings map[string]string) error {
	s.lastUpdates = make(map[string]string, len(settings))
	for key, value := range settings {
		s.lastUpdates[key] = value
		if s.values == nil {
			s.values = map[string]string{}
		}
		s.values[key] = value
	}
	return nil
}

func (s *settingHandlerRepoStub) GetAll(ctx context.Context) (map[string]string, error) {
	out := make(map[string]string, len(s.values))
	for key, value := range s.values {
		out[key] = value
	}
	return out, nil
}

func (s *settingHandlerRepoStub) Delete(ctx context.Context, key string) error {
	panic("unexpected Delete call")
}

func TestSettingHandler_UpdateSettings_PersistsAdvancedScheduler(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := &settingHandlerRepoStub{values: map[string]string{}}
	svc := service.NewSettingService(repo, &config.Config{Default: config.DefaultConfig{UserConcurrency: 5}})
	handler := NewSettingHandler(svc, nil, nil)

	body := map[string]any{
		"openai_advanced_scheduler_enabled":                       true,
		"openai_oauth_scheduling_rate_multiplier":                 0.05,
		"openai_advanced_scheduler_subscription_priority_enabled": true,
	}
	rawBody, err := json.Marshal(body)
	require.NoError(t, err)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/v1/admin/settings", bytes.NewReader(rawBody))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateSettings(c)

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "true", repo.values["openai_advanced_scheduler_enabled"])
	require.Equal(t, "0.05", repo.values[service.SettingKeyOpenAIOAuthSchedulingRateMultiplier])
	require.Equal(t, "true", repo.values[service.SettingKeyOpenAIAdvancedSchedulerSubscriptionPriorityEnabled])

	var resp response.Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	data, ok := resp.Data.(map[string]any)
	require.True(t, ok)
	require.Equal(t, true, data["openai_advanced_scheduler_enabled"])
	require.Equal(t, 0.05, data["openai_oauth_scheduling_rate_multiplier"])
	require.Equal(t, true, data["openai_advanced_scheduler_subscription_priority_enabled"])
}
