//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	binaryName = "nuclino-mcp-server"
	mainFile   = "./cmd/server/main.go"
	binDir     = "bin"
)

// Build builds the binary
func Build() error {
	mg.Deps(Clean)
	fmt.Println("Building", binaryName+"...")
	
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}
	
	return sh.Run("go", "build", "-o", filepath.Join(binDir, binaryName), mainFile)
}

// Test runs all tests
func Test() error {
	fmt.Println("Running tests...")
	return sh.Run("go", "test", "-v", "./...")
}

// TestCoverage runs tests with coverage report
func TestCoverage() error {
	fmt.Println("Running tests with coverage...")
	if err := sh.Run("go", "test", "-coverprofile=coverage.out", "./..."); err != nil {
		return err
	}
	if err := sh.Run("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html"); err != nil {
		return err
	}
	fmt.Println("Coverage report generated: coverage.html")
	return nil
}

// Benchmark runs benchmark tests
func Benchmark() error {
	fmt.Println("Running benchmarks...")
	return sh.Run("go", "test", "-bench=.", "-benchmem", "./...")
}

// Lint runs the linter
func Lint() error {
	fmt.Println("Running linter...")
	// Check if golangci-lint is available
	if err := sh.Run("golangci-lint", "--version"); err != nil {
		fmt.Println("golangci-lint not found, installing...")
		if err := installGolangciLint(); err != nil {
			return err
		}
	}
	return sh.Run("golangci-lint", "run")
}

// Fmt formats the code
func Fmt() error {
	fmt.Println("Formatting code...")
	return sh.Run("go", "fmt", "./...")
}

// Install downloads and organizes dependencies
func Install() error {
	fmt.Println("Installing dependencies...")
	if err := sh.Run("go", "mod", "download"); err != nil {
		return err
	}
	return sh.Run("go", "mod", "tidy")
}

// Clean removes build artifacts
func Clean() error {
	fmt.Println("Cleaning...")
	os.RemoveAll(binDir)
	os.Remove("coverage.out")
	os.Remove("coverage.html")
	return nil
}

// Run builds and runs the server
func Run() error {
	mg.Deps(Build)
	fmt.Println("Starting", binaryName+"...")
	return sh.Run(filepath.Join(".", binDir, binaryName))
}

// Dev builds and runs the server in debug mode
func Dev() error {
	mg.Deps(Build)
	fmt.Println("Starting", binaryName, "in development mode...")
	return sh.Run(filepath.Join(".", binDir, binaryName), "--debug")
}

// Docker builds Docker image
func Docker() error {
	fmt.Println("Building Docker image...")
	return sh.Run("docker", "build", "-t", binaryName+":latest", ".")
}

// InstallTools installs development tools
func InstallTools() error {
	fmt.Println("Installing development tools...")
	return sh.Run("go", "install", "github.com/golang/mock/mockgen@latest")
}

// GenerateMocks generates test mocks
func GenerateMocks() error {
	fmt.Println("Generating mocks...")
	// Check if mockgen is available
	if err := sh.Run("mockgen", "-version"); err != nil {
		return fmt.Errorf("mockgen not found, please run 'mage installtools' first")
	}
	return sh.Run("mockgen", "-source=internal/nuclino/client.go", "-destination=tests/mocks/client_mock.go")
}

// Security runs security scan
func Security() error {
	fmt.Println("Running security scan...")
	// Check if gosec is available
	if err := sh.Run("gosec", "-version"); err != nil {
		fmt.Println("gosec not found, installing...")
		if err := installGosec(); err != nil {
			return err
		}
	}
	return sh.Run("gosec", "./...")
}

// CheckGo checks Go version
func CheckGo() error {
	fmt.Println("Checking Go version...")
	return sh.Run("go", "version")
}

// BuildAll builds for all platforms
func BuildAll() error {
	mg.Deps(Clean)
	fmt.Println("Building for all platforms...")
	
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return err
	}

	platforms := []struct {
		goos, goarch, ext string
	}{
		{"linux", "amd64", ""},
		{"darwin", "amd64", ""},
		{"darwin", "arm64", ""},
		{"windows", "amd64", ".exe"},
	}

	for _, platform := range platforms {
		binary := fmt.Sprintf("%s-%s-%s%s", binaryName, platform.goos, platform.goarch, platform.ext)
		fmt.Printf("Building %s...\n", binary)
		
		env := map[string]string{
			"GOOS":   platform.goos,
			"GOARCH": platform.goarch,
		}
		
		if err := sh.RunWith(env, "go", "build", "-o", filepath.Join(binDir, binary), mainFile); err != nil {
			return err
		}
	}
	
	return nil
}

// CI runs all CI tasks
func CI() {
	mg.Deps(Install, Fmt, Lint, Test, Build)
}

// All runs common development tasks
func All() {
	mg.Deps(Install, Fmt, Lint, TestCoverage, Build)
}

// Helper functions

func installGolangciLint() error {
	goos := runtime.GOOS
	var installCmd []string
	
	switch goos {
	case "linux", "darwin":
		installCmd = []string{"sh", "-c", "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2"}
	case "windows":
		return fmt.Errorf("please install golangci-lint manually on Windows")
	default:
		return fmt.Errorf("unsupported OS: %s", goos)
	}
	
	return sh.Run(installCmd[0], installCmd[1:]...)
}

func installGosec() error {
	goos := runtime.GOOS
	var installCmd []string
	
	switch goos {
	case "linux", "darwin":
		installCmd = []string{"sh", "-c", "curl -sfL https://raw.githubusercontent.com/securecodewarrior/gosec/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.18.2"}
	case "windows":
		return fmt.Errorf("please install gosec manually on Windows")
	default:
		return fmt.Errorf("unsupported OS: %s", goos)
	}
	
	return sh.Run(installCmd[0], installCmd[1:]...)
}