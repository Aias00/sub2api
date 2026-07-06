package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const runtimeWorkerDeployOutputLimit = 32 * 1024

var runtimeWorkerImagePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:/@-]{0,299}$`)

type runtimeWorkerDeployRequest struct {
	Image   string
	Pull    bool
	Restart bool
}

type runtimeWorkerDeployResult struct {
	Image      string
	EnvKey     string
	Service    string
	BackupPath string
	Commands   []string
}

type runtimeWorkerCommandRunner interface {
	Run(ctx context.Context, dir string, name string, args ...string) ([]byte, error)
}

type runtimeWorkerExecCommandRunner struct{}

func (runtimeWorkerExecCommandRunner) Run(ctx context.Context, dir string, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if len(output) > runtimeWorkerDeployOutputLimit {
		output = output[len(output)-runtimeWorkerDeployOutputLimit:]
	}
	return output, err
}

type runtimeWorkerDeployAdapter struct {
	enabled         bool
	reason          string
	deployDir       string
	envFile         string
	composeFiles    []string
	allowedPrefixes []string
	commandRunner   runtimeWorkerCommandRunner
}

func newRuntimeWorkerDeployAdapter() runtimeWorkerDeployAdapter {
	adapter := runtimeWorkerDeployAdapter{
		enabled:         homeBusinessEnvEnabled("WORKER_MANAGER_DEPLOY_ENABLED"),
		deployDir:       envOrDefault("WORKER_MANAGER_DEPLOY_DIR", "."),
		envFile:         envOrDefault("WORKER_MANAGER_DEPLOY_ENV_FILE", "deploy/.env"),
		composeFiles:    splitRuntimeWorkerCSV(envOrDefault("WORKER_MANAGER_COMPOSE_FILES", "deploy/docker-compose.yml,deploy/docker-compose.business-worker.yml,deploy/docker-compose.content-worker.yml,deploy/docker-compose.images.yml")),
		allowedPrefixes: splitRuntimeWorkerCSV(os.Getenv("WORKER_MANAGER_IMAGE_ALLOWLIST")),
		commandRunner:   runtimeWorkerExecCommandRunner{},
	}
	if !adapter.enabled {
		adapter.reason = "Worker deploy actions are disabled. Set WORKER_MANAGER_DEPLOY_ENABLED=true to allow controlled image updates."
		return adapter
	}
	if len(adapter.composeFiles) == 0 {
		adapter.enabled = false
		adapter.reason = "Worker deploy compose files are not configured."
		return adapter
	}
	if _, err := os.Stat(adapter.absEnvFile()); err != nil {
		adapter.enabled = false
		adapter.reason = "Worker deploy env file is not available at " + adapter.absEnvFile() + "."
		return adapter
	}
	return adapter
}

func (a runtimeWorkerDeployAdapter) disabledReason() string {
	if a.reason != "" {
		return a.reason
	}
	if a.enabled {
		return ""
	}
	return "Worker deploy actions are disabled."
}

func (a runtimeWorkerDeployAdapter) enrich(worker adminWorkerRuntimeStatusDTO) adminWorkerRuntimeStatusDTO {
	target, ok := runtimeWorkerTarget(worker.NodeID)
	if !ok {
		target, ok = runtimeWorkerTarget(worker.ID)
	}
	if !ok || target.ImageEnvKey == "" {
		worker.DeployReason = "Worker is not mapped to a deployable image."
		return worker
	}
	worker.Image = strings.TrimSpace(os.Getenv(target.ImageEnvKey))
	worker.Deployable = a.enabled
	if !a.enabled {
		worker.DeployReason = a.disabledReason()
		return worker
	}
	return worker
}

func (a runtimeWorkerDeployAdapter) deploy(ctx context.Context, target runtimeWorkerTargetInfo, req runtimeWorkerDeployRequest) (runtimeWorkerDeployResult, error) {
	if !a.enabled {
		return runtimeWorkerDeployResult{}, errors.New(a.disabledReason())
	}
	if target.ImageEnvKey == "" || target.ComposeService == "" {
		return runtimeWorkerDeployResult{}, errors.New("worker target is not deployable")
	}
	image, err := a.validateImage(req.Image)
	if err != nil {
		return runtimeWorkerDeployResult{}, err
	}
	backupPath, err := updateRuntimeWorkerEnvFile(a.absEnvFile(), target.ImageEnvKey, image)
	if err != nil {
		return runtimeWorkerDeployResult{}, err
	}
	result := runtimeWorkerDeployResult{
		Image:      image,
		EnvKey:     target.ImageEnvKey,
		Service:    target.ComposeService,
		BackupPath: backupPath,
	}
	if req.Pull {
		args := a.composeArgs(target.ComposeProfiles, "pull", target.ComposeService)
		output, err := a.commandRunner.Run(ctx, a.deployDir, "docker", args...)
		result.Commands = append(result.Commands, runtimeWorkerCommandSummary("docker", args, output))
		if err != nil {
			return result, fmt.Errorf("docker compose pull failed: %w: %s", err, strings.TrimSpace(string(output)))
		}
	}
	if req.Restart {
		args := a.composeArgs(target.ComposeProfiles, "up", "-d", "--no-deps", target.ComposeService)
		output, err := a.commandRunner.Run(ctx, a.deployDir, "docker", args...)
		result.Commands = append(result.Commands, runtimeWorkerCommandSummary("docker", args, output))
		if err != nil {
			return result, fmt.Errorf("docker compose up failed: %w: %s", err, strings.TrimSpace(string(output)))
		}
	}
	return result, nil
}

func (a runtimeWorkerDeployAdapter) validateImage(raw string) (string, error) {
	image := strings.TrimSpace(raw)
	if image == "" {
		return "", errors.New("image is required")
	}
	if !runtimeWorkerImagePattern.MatchString(image) {
		return "", errors.New("image contains unsupported characters")
	}
	if len(a.allowedPrefixes) == 0 {
		return image, nil
	}
	for _, prefix := range a.allowedPrefixes {
		if strings.HasPrefix(image, prefix) {
			return image, nil
		}
	}
	return "", errors.New("image is not allowed by WORKER_MANAGER_IMAGE_ALLOWLIST")
}

func (a runtimeWorkerDeployAdapter) composeArgs(profiles []string, command ...string) []string {
	args := []string{"compose", "--env-file", a.envFile}
	for _, file := range a.composeFiles {
		args = append(args, "-f", file)
	}
	for _, profile := range profiles {
		if strings.TrimSpace(profile) != "" {
			args = append(args, "--profile", strings.TrimSpace(profile))
		}
	}
	args = append(args, command...)
	return args
}

func (a runtimeWorkerDeployAdapter) absEnvFile() string {
	if filepath.IsAbs(a.envFile) {
		return a.envFile
	}
	return filepath.Join(a.deployDir, a.envFile)
}

func updateRuntimeWorkerEnvFile(path string, key string, value string) (string, error) {
	if strings.TrimSpace(key) == "" {
		return "", errors.New("env key is required")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	backupPath := fmt.Sprintf("%s.bak-%s", path, time.Now().Format("20060102-150405"))
	if err := os.WriteFile(backupPath, content, 0o600); err != nil {
		return "", err
	}
	updated := rewriteRuntimeWorkerEnv(content, key, value)
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	tmpPath := fmt.Sprintf("%s.tmp-%d", path, time.Now().UnixNano())
	if err := os.WriteFile(tmpPath, updated, info.Mode().Perm()); err != nil {
		return "", err
	}
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return "", err
	}
	return backupPath, nil
}

func rewriteRuntimeWorkerEnv(content []byte, key string, value string) []byte {
	lines := bytes.Split(content, []byte("\n"))
	prefix := []byte(key + "=")
	replacement := []byte(key + "=" + value)
	replaced := false
	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)
		if bytes.HasPrefix(trimmed, []byte("#")) || !bytes.HasPrefix(line, prefix) {
			continue
		}
		lines[i] = replacement
		replaced = true
	}
	if !replaced {
		if len(content) > 0 && !bytes.HasSuffix(content, []byte("\n")) {
			lines = append(lines, []byte{})
		}
		lines = append(lines, replacement)
	}
	return bytes.Join(lines, []byte("\n"))
}

func runtimeWorkerCommandSummary(name string, args []string, output []byte) string {
	command := name + " " + strings.Join(args, " ")
	message := strings.TrimSpace(string(output))
	if message == "" {
		return command
	}
	if len(message) > 1000 {
		message = message[len(message)-1000:]
	}
	return command + "\n" + message
}

func splitRuntimeWorkerCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
