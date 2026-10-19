//go:build e2e

package e2e

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
)

const (
	testUsername = "e2e-admin"
	testPassword = "e2e-password"

	devThemeCSS = ":scope { color: rebeccapurple; }\n"
)

var (
	baseURL string
	browser playwright.Browser
	// devBaseURL is the instance in theme dev mode, whose theme in development
	// is the project of writeDevThemeProject.
	devBaseURL string
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

	url, logPath, stop, err := startUmono(binPath, tmp, "work", nil, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer stop()
	baseURL = url

	project := filepath.Join(tmp, "devtheme")
	if err := writeDevThemeProject(project); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	devURL, devLogPath, devStop, err := startUmono(binPath, tmp, "dev", []string{"UMONO_THEME_DEV=" + project}, func(workDir string) error {
		return linkDevTheme(workDir, project)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	defer devStop()
	devBaseURL = devURL

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
		printLog(logPath)
		printLog(devLogPath)
	}
	return code
}

// startUmono starts binPath in tmp/<name>/ with a .env of the test
// credentials, on a free port, with the extra environment env, and waits
// until it listens. Its log is tmp/<name>.log. prepare, when not nil, runs on
// the work directory before the start. stop stops the process.
func startUmono(binPath, tmp, name string, env []string, prepare func(workDir string) error) (url, logPath string, stop func(), err error) {
	workDir := filepath.Join(tmp, name)
	if err = os.Mkdir(workDir, 0o755); err != nil {
		return "", "", nil, err
	}

	envContent := fmt.Sprintf("APP_ENV=\nUSERNAME=%s\nPASSWORD=%s\nUMONO_SECRET=%s\n",
		testUsername, testPassword, "2b7e151628aed2a6abf7158809cf4f3c762e7160f38b4da56a784d9045190cfe")
	if err = os.WriteFile(filepath.Join(workDir, ".env"), []byte(envContent), 0o644); err != nil {
		return "", "", nil, err
	}

	if prepare != nil {
		if err = prepare(workDir); err != nil {
			return "", "", nil, err
		}
	}

	port, err := freePort()
	if err != nil {
		return "", "", nil, err
	}

	logPath = filepath.Join(tmp, name+".log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return "", "", nil, err
	}
	defer logFile.Close()

	cmd := exec.Command(binPath)
	cmd.Dir = workDir
	cmd.Env = append(append(os.Environ(), env...), "PORT="+strconv.Itoa(port))
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err = cmd.Start(); err != nil {
		return "", "", nil, err
	}

	stop = func() {
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
	}

	url = fmt.Sprintf("http://127.0.0.1:%d", port)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	deadline := time.Now().Add(30 * time.Second)
	for {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			conn.Close()
			break
		}
		if time.Now().After(deadline) {
			stop()
			return "", logPath, stop, fmt.Errorf("umono did not start")
		}
		time.Sleep(100 * time.Millisecond)
	}

	return url, logPath, stop, nil
}

func printLog(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Fprintf(os.Stderr, "--- %s ---\n%s", filepath.Base(path), data)
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// writeDevThemeProject writes a theme project at project as umono theme dev
// leaves it: the contract of internal/theme's test package in the contract
// cache, its attributes.toml and an assets/theme.css.
func writeDevThemeProject(project string) error {
	contractSrc, err := os.ReadFile("../internal/theme/testdata/package/contract.toml")
	if err != nil {
		return err
	}
	attributesSrc, err := os.ReadFile("../internal/theme/testdata/package/attributes.toml")
	if err != nil {
		return err
	}

	contractPath := filepath.Join(project, ".dev", "contracts", "example", "blog@v1.0.toml")
	if err := os.MkdirAll(filepath.Dir(contractPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(contractPath, contractSrc, 0o644); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(project, "attributes.toml"), attributesSrc, 0o644); err != nil {
		return err
	}

	assetsDir := filepath.Join(project, "assets")
	if err := os.MkdirAll(assetsDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "theme.css"), []byte(devThemeCSS), 0o644); err != nil {
		return err
	}

	return nil
}

// linkDevTheme writes the dev directory themes/dev/dev/0.0.0/ under workDir
// as umono theme dev does: built.toml with the id dev/dev and the version
// 0.0.0, and symlinks into project.
func linkDevTheme(workDir, project string) error {
	dir := filepath.Join(workDir, "themes", "dev", "dev", "0.0.0")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data, err := os.ReadFile("../internal/theme/testdata/package/built.toml")
	if err != nil {
		return err
	}
	b := strings.Replace(string(data), "acme/minimal", "dev/dev", 1)
	b = strings.Replace(b, "0.1.0", "0.0.0", 1)
	if err := os.WriteFile(filepath.Join(dir, "built.toml"), []byte(b), 0o644); err != nil {
		return err
	}

	if err := os.Symlink(filepath.Join(project, ".dev", "contracts", "example", "blog@v1.0.toml"), filepath.Join(dir, "contract.toml")); err != nil {
		return err
	}
	if err := os.Symlink(filepath.Join(project, "attributes.toml"), filepath.Join(dir, "attributes.toml")); err != nil {
		return err
	}
	if err := os.Symlink(filepath.Join(project, "assets"), filepath.Join(dir, "assets")); err != nil {
		return err
	}

	return nil
}
