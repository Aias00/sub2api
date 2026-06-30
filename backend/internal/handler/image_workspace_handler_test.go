package handler

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestResolveImageWorkspaceStoragePathRestrictsToRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "image-workspace")
	t.Setenv("IMAGE_WORKSPACE_STORAGE_ROOT", root)

	allowed, ok := resolveImageWorkspaceStoragePath(filepath.Join(root, "user-1", "task-1", "image.png"))
	require.True(t, ok)
	require.Equal(t, filepath.Join(root, "user-1", "task-1", "image.png"), allowed)

	_, ok = resolveImageWorkspaceStoragePath(filepath.Join(root, "..", "secret.png"))
	require.False(t, ok)

	_, ok = resolveImageWorkspaceStoragePath("relative/image.png")
	require.False(t, ok)
}

func TestImageWorkspaceRemoteArtifactRequiresAllowlistedHost(t *testing.T) {
	require.False(t, isImageWorkspaceRemoteArtifactAllowed("https://evil.example.com/image.png"))

	t.Setenv("IMAGE_WORKSPACE_OBJECT_STORAGE_PUBLIC_BASE_URL", "https://static.example.com/image-workspace")
	require.True(t, isImageWorkspaceRemoteArtifactAllowed("https://static.example.com/image-workspace/user-1/image.png"))
	require.False(t, isImageWorkspaceRemoteArtifactAllowed("https://static.example.net/image-workspace/user-1/image.png"))
	require.False(t, isImageWorkspaceRemoteArtifactAllowed("javascript:alert(1)"))
}

func TestImageWorkspaceRemoteArtifactRedirectsForPublicStorage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("IMAGE_WORKSPACE_OBJECT_STORAGE_PUBLIC_BASE_URL", "https://static.example.com/image-workspace")

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/download", nil)

	ok := redirectImageWorkspaceRemoteArtifact(ctx, &service.ImageWorkspaceArtifact{
		ID:              7,
		TaskID:          3,
		StorageProvider: "r2",
		ImageURL:        "https://static.example.com/image-workspace/user-1/task-3/image.png",
	}, "image-task-3-7.png", time.Now())

	require.True(t, ok)
	require.Equal(t, http.StatusFound, recorder.Code)
	require.Equal(t, "https://static.example.com/image-workspace/user-1/task-3/image.png", recorder.Header().Get("Location"))
}

func TestImageWorkspaceUpstreamArtifactDoesNotRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Setenv("IMAGE_WORKSPACE_ARTIFACT_REMOTE_HOST_ALLOWLIST", "static.example.com")

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/download", nil)

	ok := redirectImageWorkspaceRemoteArtifact(ctx, &service.ImageWorkspaceArtifact{
		ID:              8,
		TaskID:          4,
		StorageProvider: "upstream",
		ImageURL:        "https://static.example.com/image.png",
	}, "image-task-4-8.png", time.Now())

	require.False(t, ok)
	require.Equal(t, http.StatusOK, recorder.Code)
}
