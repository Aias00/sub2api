package handler

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestResolveWeChatExportStoragePathRestrictsToRoot(t *testing.T) {
	root := filepath.Join(t.TempDir(), "wechat-export")
	t.Setenv("WECHAT_EXPORT_STORAGE_ROOT", root)

	allowed, ok := resolveWeChatExportStoragePath(filepath.Join(root, "task-1", "export.json"))
	require.True(t, ok)
	require.Equal(t, filepath.Join(root, "task-1", "export.json"), allowed)

	_, ok = resolveWeChatExportStoragePath(filepath.Join(root, "..", "secret.txt"))
	require.False(t, ok)

	_, ok = resolveWeChatExportStoragePath("relative/export.json")
	require.False(t, ok)
}

func TestUniqueWeChatArtifactZipNameSanitizesAndDeduplicates(t *testing.T) {
	seen := map[string]int{}

	require.Equal(t, "export.json", uniqueWeChatArtifactZipName("../export.json", "json", seen))
	require.Equal(t, "export-2.json", uniqueWeChatArtifactZipName("export.json", "json", seen))
	require.Equal(t, "wechat-export.html", uniqueWeChatArtifactZipName("", "html", seen))
	require.NotContains(t, uniqueWeChatArtifactZipName("..\\secret.md", "markdown", seen), "\\")
}

func TestWriteWeChatArtifactZipEntryWritesDownloadableFile(t *testing.T) {
	root := t.TempDir()
	artifactPath := filepath.Join(root, "task-1", "demo.html")
	require.NoError(t, os.MkdirAll(filepath.Dir(artifactPath), 0o755))
	require.NoError(t, os.WriteFile(artifactPath, []byte("<h1>export</h1>"), 0o600))

	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)
	require.NoError(t, writeWeChatArtifactZipEntry(zipWriter, wechatArtifactZipEntry{
		Artifact: service.WeChatExportArtifact{Format: "html", FileName: "../demo.html"},
		Path:     artifactPath,
		Name:     uniqueWeChatArtifactZipName("../demo.html", "html", map[string]int{}),
	}))
	require.NoError(t, zipWriter.Close())

	reader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	require.NoError(t, err)
	require.Len(t, reader.File, 1)
	require.Equal(t, "demo.html", reader.File[0].Name)

	entry, err := reader.File[0].Open()
	require.NoError(t, err)
	defer func() { _ = entry.Close() }()
	content, err := os.ReadFile(artifactPath)
	require.NoError(t, err)
	zippedContent, err := io.ReadAll(entry)
	require.NoError(t, err)
	require.Equal(t, content, zippedContent)
}

func TestWeChatArtifactRemoteZipEntryRequiresAllowlistedHost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("remote artifact"))
	}))
	defer server.Close()

	require.False(t, isWeChatArtifactRemoteZipAllowed(server.URL+"/artifact.json"))

	parsedHost := strings.TrimPrefix(server.URL, "http://")
	t.Setenv("WECHAT_EXPORT_ZIP_REMOTE_HOST_ALLOWLIST", parsedHost)
	require.True(t, isWeChatArtifactRemoteZipAllowed(server.URL+"/artifact.json"))

	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)
	require.NoError(t, writeWeChatArtifactZipEntry(zipWriter, wechatArtifactZipEntry{
		Artifact:  service.WeChatExportArtifact{Format: "json", FileName: "remote.json"},
		RemoteURL: server.URL + "/artifact.json",
		Name:      "remote.json",
	}))
	require.NoError(t, zipWriter.Close())

	reader, err := zip.NewReader(bytes.NewReader(buffer.Bytes()), int64(buffer.Len()))
	require.NoError(t, err)
	require.Len(t, reader.File, 1)
	entry, err := reader.File[0].Open()
	require.NoError(t, err)
	defer func() { _ = entry.Close() }()
	content, err := io.ReadAll(entry)
	require.NoError(t, err)
	require.Equal(t, "remote artifact", string(content))
}

func TestWeChatArtifactRemoteDownloadRequiresAllowlistedHost(t *testing.T) {
	require.False(t, isWeChatArtifactRemoteDownloadAllowed("https://evil.example.com/artifact.html"))

	t.Setenv("WECHAT_EXPORT_ARTIFACT_PUBLIC_BASE_URL", "https://files.example.com/export")
	require.True(t, isWeChatArtifactRemoteDownloadAllowed("https://files.example.com/artifact.html"))
	require.False(t, isWeChatArtifactRemoteDownloadAllowed("https://files.example.net/artifact.html"))
	require.False(t, isWeChatArtifactRemoteDownloadAllowed("javascript:alert(1)"))
}
