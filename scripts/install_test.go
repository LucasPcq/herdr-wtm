package scripts

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
)

const version = "9.9.9"

var archiveName = fmt.Sprintf("herdr-wtm_%s_%s_%s.tar.gz", version, runtime.GOOS, runtime.GOARCH)

// pluginRoot lays out a fake plugin checkout: the script under scripts/ and a manifest.
func pluginRoot(t *testing.T, name string) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(filepath.Join(root, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	script, err := os.ReadFile("install.sh")
	if err != nil {
		t.Fatal(err)
	}
	must(t, os.WriteFile(filepath.Join(root, "scripts", "install.sh"), script, 0o755))
	must(t, os.WriteFile(filepath.Join(root, "herdr-plugin.toml"), []byte("id = \"lucaspcq.wtm\"\nversion = \""+version+"\"\nmin_herdr_version = \"0.9.0\"\n"), 0o644))
	return root
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// archive returns a tar.gz holding an executable herdr-wtm that prints "prebuilt".
func archive(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	body := []byte("#!/bin/sh\necho prebuilt\n")
	must(t, tw.WriteHeader(&tar.Header{Name: "herdr-wtm", Mode: 0o755, Size: int64(len(body))}))
	_, err := tw.Write(body)
	must(t, err)
	must(t, tw.Close())
	must(t, gz.Close())
	return buf.Bytes()
}

// releaseServer serves v<version>/<archive> and checksums.txt; checksums
// overrides the checksums file content when non-empty. hits counts requests.
func releaseServer(t *testing.T, checksums string) (*httptest.Server, *atomic.Int32) {
	t.Helper()
	data := archive(t)
	if checksums == "" {
		checksums = fmt.Sprintf("%x  %s\n", sha256.Sum256(data), archiveName)
	}
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		switch r.URL.Path {
		case "/v" + version + "/" + archiveName:
			_, _ = w.Write(data)
		case "/v" + version + "/checksums.txt":
			_, _ = w.Write([]byte(checksums))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

// fakeGo puts a `go` on PATH that "builds" a binary printing "source".
func fakeGo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	script := "#!/bin/sh\nmkdir -p bin && printf '#!/bin/sh\\necho source\\n' > bin/herdr-wtm && chmod +x bin/herdr-wtm\n"
	must(t, os.WriteFile(filepath.Join(dir, "go"), []byte(script), 0o755))
	return dir
}

// install runs the script with a minimal PATH (system tools, plus extraPath first).
func install(t *testing.T, root, baseURL, extraPath string, env ...string) (string, error) {
	t.Helper()
	path := "/usr/bin:/bin:/usr/sbin:/sbin"
	if extraPath != "" {
		path = extraPath + ":" + path
	}
	cmd := exec.Command("sh", filepath.Join(root, "scripts", "install.sh"))
	cmd.Dir = t.TempDir()
	cmd.Env = append([]string{"PATH=" + path, "HOME=" + t.TempDir(), "HERDR_WTM_RELEASE_BASE_URL=" + baseURL}, env...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// runBinary executes the installed binary and returns what it prints.
func runBinary(t *testing.T, root string) string {
	t.Helper()
	out, err := exec.Command(filepath.Join(root, "bin", "herdr-wtm")).Output()
	if err != nil {
		t.Fatalf("installed binary does not run: %v", err)
	}
	return strings.TrimSpace(string(out))
}

func TestInstallDownloadsPrebuiltBinary(t *testing.T) {
	srv, _ := releaseServer(t, "")
	root := pluginRoot(t, "plugin")
	if out, err := install(t, root, srv.URL, ""); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got := runBinary(t, root); got != "prebuilt" {
		t.Fatalf("binary prints %q", got)
	}
}

func TestInstallFallsBackToGoWhenReleaseMissing(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	root := pluginRoot(t, "plugin")
	out, err := install(t, root, srv.URL, fakeGo(t))
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got := runBinary(t, root); got != "source" {
		t.Fatalf("binary prints %q", got)
	}
	if !strings.Contains(out, "building from source") {
		t.Fatalf("output %q", out)
	}
}

func TestInstallFailsWithoutReleaseAndGo(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	t.Cleanup(srv.Close)
	root := pluginRoot(t, "plugin")
	out, err := install(t, root, srv.URL, "")
	if err == nil {
		t.Fatalf("want failure, got success:\n%s", out)
	}
	if !strings.Contains(out, "Go is not installed") {
		t.Fatalf("output %q", out)
	}
}

func TestInstallChecksumMismatchFallsBack(t *testing.T) {
	srv, _ := releaseServer(t, strings.Repeat("0", 64)+"  "+archiveName+"\n")
	root := pluginRoot(t, "plugin")
	out, err := install(t, root, srv.URL, fakeGo(t))
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got := runBinary(t, root); got != "source" {
		t.Fatalf("binary prints %q (an unverified archive was installed)", got)
	}
}

func TestInstallMissingChecksumLineFallsBack(t *testing.T) {
	srv, _ := releaseServer(t, strings.Repeat("a", 64)+"  some-other-file.tar.gz\n")
	root := pluginRoot(t, "plugin")
	out, err := install(t, root, srv.URL, fakeGo(t))
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got := runBinary(t, root); got != "source" {
		t.Fatalf("binary prints %q (an unverified archive was installed)", got)
	}
}

func TestInstallForcedSourceBuildSkipsNetwork(t *testing.T) {
	srv, hits := releaseServer(t, "")
	root := pluginRoot(t, "plugin")
	out, err := install(t, root, srv.URL, fakeGo(t), "HERDR_WTM_BUILD_FROM_SOURCE=1")
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got := runBinary(t, root); got != "source" {
		t.Fatalf("binary prints %q", got)
	}
	if hits.Load() != 0 {
		t.Fatalf("%d network requests in forced source mode", hits.Load())
	}
}

func TestInstallReplacesExistingBinary(t *testing.T) {
	srv, _ := releaseServer(t, "")
	root := pluginRoot(t, "plugin")
	must(t, os.MkdirAll(filepath.Join(root, "bin"), 0o755))
	must(t, os.WriteFile(filepath.Join(root, "bin", "herdr-wtm"), []byte("#!/bin/sh\necho old\n"), 0o755))
	if out, err := install(t, root, srv.URL, ""); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got := runBinary(t, root); got != "prebuilt" {
		t.Fatalf("binary prints %q", got)
	}
}

func TestInstallWorksWithSpaceInRoot(t *testing.T) {
	srv, _ := releaseServer(t, "")
	root := pluginRoot(t, "my plugin")
	if out, err := install(t, root, srv.URL, ""); err != nil {
		t.Fatalf("%v\n%s", err, out)
	}
	if got := runBinary(t, root); got != "prebuilt" {
		t.Fatalf("binary prints %q", got)
	}
}
