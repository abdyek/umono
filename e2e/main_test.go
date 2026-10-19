//go:build e2e

package e2e

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
)

const (
	testUsername = "e2e-admin"
	testPassword = "e2e-password"
)

var (
	baseURL string
	browser playwright.Browser
)

func TestMain(m *testing.M) {
	os.Exit(run(m))
}

func run(m *testing.M) int {
	tmp, err := os.MkdirTemp("", "umono-e2e-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer os.RemoveAll(tmp)

	binPath := filepath.Join(tmp, "umono")
	build := exec.Command("go", "build", "-o", binPath, "./cmd/umono")
	build.Dir = ".."
	build.Stdout = os.Stderr
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	workDir := filepath.Join(tmp, "work")
	if err := os.Mkdir(workDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	envContent := fmt.Sprintf("APP_ENV=\nUSERNAME=%s\nPASSWORD=%s\nUMONO_SECRET=%s\n",
		testUsername, testPassword, "2b7e151628aed2a6abf7158809cf4f3c762e7160f38b4da56a784d9045190cfe")
	if err := os.WriteFile(filepath.Join(workDir, ".env"), []byte(envContent), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	port, err := freePort()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	logFile, err := os.Create(filepath.Join(tmp, "umono.log"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer logFile.Close()

	cmd := exec.Command(binPath)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), "PORT="+strconv.Itoa(port))
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer func() {
		if cmd.Process == nil {
			return
		}
		_ = cmd.Process.Signal(os.Interrupt)
		done := make(chan error, 1)
		go func() { done <- cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = cmd.Process.Kill()
			<-done
		}
	}()

	baseURL = fmt.Sprintf("http://127.0.0.1:%d", port)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	deadline := time.Now().Add(30 * time.Second)
	for {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			conn.Close()
			break
		}
		if time.Now().After(deadline) {
			fmt.Fprintln(os.Stderr, "umono did not start")
			return 1
		}
		time.Sleep(100 * time.Millisecond)
	}

	pw, err := playwright.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer pw.Stop()

	browser, err = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(true)})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer browser.Close()

	code := m.Run()
	if code != 0 {
		// The app's log helps to see why a test failed.
		printLog(logFile.Name())
	}
	return code
}

func printLog(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Fprintf(os.Stderr, "--- umono.log ---\n%s", data)
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
