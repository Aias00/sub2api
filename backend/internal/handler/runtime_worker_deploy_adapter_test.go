package handler

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

type runtimeWorkerCommandRunnerStub struct {
	calls []runtimeWorkerCommandCall
}

type runtimeWorkerCommandCall struct {
	dir  string
	name string
	args []string
}

func (s *runtimeWorkerCommandRunnerStub) Run(_ context.Context, dir string, name string, args ...string) ([]byte, error) {
	s.calls = append(s.calls, runtimeWorkerCommandCall{
		dir:  dir,
		name: name,
		args: append([]string{}, args...),
	})
	return []byte("ok"), nil
}

func TestRuntimeWorkerDeployAdapterValidateImage(t *testing.T) {
	adapter := runtimeWorkerDeployAdapter{
		allowedPrefixes: []string{"registry.cn-qingdao.aliyuncs.com/cola/images:"},
	}

	image, err := adapter.validateImage(" registry.cn-qingdao.aliyuncs.com/cola/images:content-worker-sha-a2eef2d ")
	require.NoError(t, err)
	require.Equal(t, "registry.cn-qingdao.aliyuncs.com/cola/images:content-worker-sha-a2eef2d", image)

	_, err = adapter.validateImage("registry.cn-qingdao.aliyuncs.com/cola/images:latest;docker ps")
	require.EqualError(t, err, "image contains unsupported characters")

	_, err = adapter.validateImage("ghcr.io/aias00/cloudbase-content-worker:latest")
	require.EqualError(t, err, "image is not allowed by WORKER_MANAGER_IMAGE_ALLOWLIST")
}

func TestUpdateRuntimeWorkerEnvFileRewritesOnlyTargetKey(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	require.NoError(t, os.WriteFile(envPath, []byte("CONTENT_WORKER_IMAGE=old\n# CONTENT_WORKER_IMAGE=ignored\nWECHAT_WORKER_IMAGE=wechat\n"), 0o640))

	backupPath, err := updateRuntimeWorkerEnvFile(envPath, "CONTENT_WORKER_IMAGE", "registry.cn-qingdao.aliyuncs.com/cola/images:content-worker-sha-new")
	require.NoError(t, err)
	require.FileExists(t, backupPath)

	content, err := os.ReadFile(envPath)
	require.NoError(t, err)
	require.Equal(t, "CONTENT_WORKER_IMAGE=registry.cn-qingdao.aliyuncs.com/cola/images:content-worker-sha-new\n# CONTENT_WORKER_IMAGE=ignored\nWECHAT_WORKER_IMAGE=wechat\n", string(content))
}

func TestRuntimeWorkerDeployAdapterDeployRunsFixedComposeCommands(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "deploy"), 0o755))
	envPath := filepath.Join(dir, "deploy", ".env")
	require.NoError(t, os.WriteFile(envPath, []byte("CONTENT_WORKER_IMAGE=old\n"), 0o600))
	runner := &runtimeWorkerCommandRunnerStub{}
	adapter := runtimeWorkerDeployAdapter{
		enabled:         true,
		deployDir:       dir,
		envFile:         "deploy/.env",
		composeFiles:    []string{"deploy/docker-compose.yml", "deploy/docker-compose.content-worker.yml"},
		allowedPrefixes: []string{"registry.cn-qingdao.aliyuncs.com/cola/images:"},
		commandRunner:   runner,
	}
	target := runtimeWorkerTargetInfo{
		ID:              workerNodeContent,
		ImageEnvKey:     "CONTENT_WORKER_IMAGE",
		ComposeService:  "content-worker",
		ComposeProfiles: []string{"content-worker"},
	}

	result, err := adapter.deploy(context.Background(), target, runtimeWorkerDeployRequest{
		Image:   "registry.cn-qingdao.aliyuncs.com/cola/images:content-worker-sha-new",
		Pull:    true,
		Restart: true,
	})
	require.NoError(t, err)
	require.Equal(t, "CONTENT_WORKER_IMAGE", result.EnvKey)
	require.Equal(t, "content-worker", result.Service)
	require.Len(t, runner.calls, 2)
	require.Equal(t, "docker", runner.calls[0].name)
	require.Equal(t, []string{
		"compose", "--env-file", "deploy/.env",
		"-f", "deploy/docker-compose.yml",
		"-f", "deploy/docker-compose.content-worker.yml",
		"--profile", "content-worker",
		"pull", "content-worker",
	}, runner.calls[0].args)
	require.Equal(t, []string{
		"compose", "--env-file", "deploy/.env",
		"-f", "deploy/docker-compose.yml",
		"-f", "deploy/docker-compose.content-worker.yml",
		"--profile", "content-worker",
		"up", "-d", "--no-deps", "content-worker",
	}, runner.calls[1].args)
}
