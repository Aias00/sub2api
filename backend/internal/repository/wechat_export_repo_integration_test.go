//go:build integration

package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// WeChatExportRepoIntegrationSuite tests wechat export repository with real database
type WeChatExportRepoIntegrationSuite struct {
	suite.Suite
	ctx  context.Context
	repo *wechatExportRepository
}

// SetupTest runs before each test
func (s *WeChatExportRepoIntegrationSuite) SetupTest() {
	s.ctx = context.Background()
	// Use integration test harness
	tx := testEntTx(s.T())
	s.repo = &wechatExportRepository{
		db:  integrationDB,
		sql: tx,
	}
}

// TearDownTest runs after each test
func (s *WeChatExportRepoIntegrationSuite) TearDownTest() {
	// Cleanup is handled by transaction rollback
}

func TestWeChatExportRepoIntegrationSuite(t *testing.T) {
	suite.Run(t, new(WeChatExportRepoIntegrationSuite))
}

// TestClaimNextTask_Concurrent tests concurrent task claiming with FOR UPDATE SKIP LOCKED
func (s *WeChatExportRepoIntegrationSuite) TestClaimNextTask_Concurrent() {
	// Create test user
	user, err := s.repo.db.ExecContext(s.ctx, `
		INSERT INTO users (email, password_hash, role, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id
	`, "test@example.com", "hash", "user", 10.0)
	require.NoError(s.T(), err)

	// Create multiple queued tasks
	for i := 0; i < 10; i++ {
		task := &service.WeChatExportTask{
			UserID:               42, // Mock user ID
			Status:               service.WeChatExportTaskStatusQueued,
			SelectedArticleCount: 3,
			PayloadJSON:          `{"article_ids":[1,2,3],"formats":["html"]}`,
			ResultManifestJSON:   "{}",
			RetentionDays:        7,
			CostEstimate:         0,
		}
		err := s.repo.CreateTask(s.ctx, task)
		require.NoError(s.T(), err)
	}

	// Concurrent claiming: 10 workers try to claim tasks simultaneously
	const workers = 10
	var wg sync.WaitGroup
	claimedTasks := make([]*service.WeChatExportTask, workers)
	leaseTokens := make([]string, workers)
	errs := make([]error, workers)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			task, _, leaseToken, err := s.repo.ClaimNextTask(s.ctx, 60)
			claimedTasks[idx] = task
			leaseTokens[idx] = leaseToken
			errs[idx] = err
		}(i)
	}
	wg.Wait()

	// Verify: all workers succeeded
	for i, err := range errs {
		require.NoError(s.T(), err, "Worker %d failed", i)
	}

	// Verify: each task is unique (no duplicate claiming)
	taskIDs := make(map[int64]bool)
	for i, task := range claimedTasks {
		if task != nil {
			require.False(s.T(), taskIDs[task.ID], "Task %d was claimed by multiple workers", task.ID)
			taskIDs[task.ID] = true
			require.NotEmpty(s.T(), leaseTokens[i], "Worker %d should have a lease token", i)
		}
	}

	// Verify: exactly 10 tasks were claimed
	require.Equal(s.T(), 10, len(taskIDs), "All 10 tasks should be claimed")
}

// TestCreateTask_ConcurrentBalanceDeduction tests concurrent balance deduction
func (s *WeChatExportRepoIntegrationSuite) TestCreateTask_ConcurrentBalanceDeduction() {
	// Create test user with limited balance
	userID := int64(42)
	balance := 5.0 // Enough for 5 tasks at 1.0 each

	// Concurrent task creation: 10 tasks trying to consume 1.0 each
	const tasks = 10
	var wg sync.WaitGroup
	errs := make([]error, tasks)

	for i := 0; i < tasks; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			task := &service.WeChatExportTask{
				UserID:               userID,
				Status:               service.WeChatExportTaskStatusQueued,
				SelectedArticleCount: 1,
				PayloadJSON:          `{"article_ids":[1],"formats":["html"]}`,
				ResultManifestJSON:   "{}",
				RetentionDays:        7,
				CostEstimate:         1.0, // Cost 1.0 per task
			}
			errs[idx] = s.repo.CreateTask(s.ctx, task)
		}(i)
	}
	wg.Wait()

	// Verify: some tasks succeed, some fail due to insufficient balance
	successCount := 0
	for i, err := range errs {
		if err == nil {
			successCount++
		} else {
			require.Equal(s.T(), service.ErrWeChatInsufficientBalance, err, "Worker %d should get insufficient balance error", i)
		}
	}

	// Verify: exactly 5 tasks succeeded (balance / cost_per_task)
	require.Equal(s.T(), 5, successCount, "Should create exactly 5 tasks with 5.0 balance")
}