package reposync

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	SourceType string
	SourceURL  string
	TargetPath string
	Branch     string
	Path       string
	SingleFile bool
	Proxy      string
	ProxyURL   string
	AuthToken  string
	HttpProxy  string
}

func Run(args []string) {
	fs := flag.NewFlagSet("reposync", flag.ExitOnError)
	var cfg Config
	fs.StringVar(&cfg.SourceType, "source-type", "git", "Source type: git or url")
	fs.StringVar(&cfg.SourceURL, "source-url", "", "Source url")
	fs.StringVar(&cfg.TargetPath, "target-path", "", "Target path")
	fs.StringVar(&cfg.Branch, "branch", "", "Branch")
	fs.StringVar(&cfg.Path, "path", "", "Path for sparse checkout")
	fs.BoolVar(&cfg.SingleFile, "single-file", false, "Single file mode")
	fs.StringVar(&cfg.Proxy, "proxy", "none", "Proxy type")
	fs.StringVar(&cfg.ProxyURL, "proxy-url", "", "Custom proxy url")
	fs.StringVar(&cfg.AuthToken, "auth-token", "", "Auth token")
	fs.StringVar(&cfg.HttpProxy, "http-proxy", "", "Http proxy")

	fs.Parse(args)

	if cfg.SourceURL == "" || cfg.TargetPath == "" {
		fmt.Println("Error: --source-url and --target-path are required")
		os.Exit(1)
	}

	fmt.Printf("Arguments: %v\n", args)

	if cfg.SourceType == "git" {
		syncGit(cfg)
	} else {
		syncURL(cfg)
	}
}

func syncGit(cfg Config) {
	env := os.Environ()

	if isRawFileURL(cfg.SourceURL) {
		fmt.Println("Raw file URL detected, switching to URL mode")
		syncURL(cfg)
		return
	}

	if cfg.HttpProxy != "" {
		env = append(env, "http_proxy="+cfg.HttpProxy, "https_proxy="+cfg.HttpProxy)
	}

	repoURL := buildProxyURL(cfg.SourceURL, cfg.Proxy, cfg.ProxyURL)
	if cfg.AuthToken != "" && strings.HasPrefix(repoURL, "https://") {
		repoURL = strings.Replace(repoURL, "https://", "https://"+cfg.AuthToken+"@", 1)
	}

	dest := cfg.TargetPath

	if cfg.Path != "" && cfg.SingleFile {
		syncGitFile(cfg, repoURL, env)
		return
	}

	gitDir := filepath.Join(dest, ".git")
	if isDir(dest) && !pathExists(gitDir) {
		repoName := getRepoName(cfg.SourceURL)
		dest = filepath.Join(dest, repoName)
		fmt.Printf("Appending repo name to target path: %s\n", dest)
		gitDir = filepath.Join(dest, ".git")
	}

	if pathExists(gitDir) {
		fmt.Println("Executing git pull")
		if cfg.Branch != "" {
			runCmd([]string{"git", "checkout", cfg.Branch}, dest, env)
		}
		runCmd([]string{"git", "pull"}, dest, env)
	} else {
		fmt.Println("Executing git clone")
		parentDir := filepath.Dir(dest)
		if parentDir != "" {
			os.MkdirAll(parentDir, 0755)
		}

		if pathExists(dest) && !isDirEmpty(dest) {
			fmt.Printf("Error: Target dir '%s' is not empty.\n", dest)
			os.Exit(1)
		}

		cloneCmd := []string{"git", "clone", "--depth", "1"}
		if cfg.Branch != "" {
			cloneCmd = append(cloneCmd, "-b", cfg.Branch)
		}

		if cfg.Path != "" {
			cloneCmd = append(cloneCmd, "--filter=blob:none", "--no-checkout", repoURL, dest)
			runCmd(cloneCmd, "", env)
			runCmd([]string{"git", "sparse-checkout", "init", "--cone"}, dest, env)
			runCmd([]string{"git", "sparse-checkout", "set", cfg.Path}, dest, env)
			runCmd([]string{"git", "checkout"}, dest, env)
		} else {
			cloneCmd = append(cloneCmd, repoURL, dest)
			runCmd(cloneCmd, "", env)
		}
	}
	fmt.Println("Sync completed")
}

func syncURL(cfg Config) {
	downloadURL := buildProxyURL(cfg.SourceURL, cfg.Proxy, cfg.ProxyURL)
	dest := cfg.TargetPath

	if isDir(dest) || strings.HasSuffix(dest, string(os.PathSeparator)) || strings.HasSuffix(dest, "/") {
		urlPath := strings.Split(cfg.SourceURL, "?")[0]
		filename := filepath.Base(urlPath)
		if filename == "" {
			filename = "downloaded_file"
		}
		dest = filepath.Join(dest, filename)
		fmt.Printf("Target file: %s\n", dest)
	}

	downloadFile(downloadURL, dest, cfg.AuthToken)
}

func syncGitFile(cfg Config, repoURL string, env []string) {
	sourceURL := cfg.SourceURL
	filePath := cfg.Path
	dest := cfg.TargetPath

	if isDir(dest) || strings.HasSuffix(dest, string(os.PathSeparator)) || strings.HasSuffix(dest, "/") {
		filename := filepath.Base(filePath)
		dest = filepath.Join(dest, filename)
		fmt.Printf("Auto corrected path to: %s\n", dest)
	}

	branch := cfg.Branch
	if branch == "" {
		branch = getRemoteDefaultBranch(repoURL, env)
	}

	cleanURL := strings.TrimSuffix(cfg.SourceURL, ".git")
	rawURL := ""

	if strings.Contains(sourceURL, "github.com") {
		base := strings.Replace(strings.TrimSuffix(cfg.SourceURL, ".git"), "github.com", "raw.githubusercontent.com", 1)
		rawURL = fmt.Sprintf("%s/%s/%s", base, branch, filePath)
	} else if strings.Contains(sourceURL, "gitlab.com") {
		rawURL = fmt.Sprintf("%s/-/raw/%s/%s", cleanURL, branch, filePath)
	} else if strings.Contains(sourceURL, "gitee.com") {
		rawURL = fmt.Sprintf("%s/raw/%s/%s", cleanURL, branch, filePath)
	} else {
		rawURL = fmt.Sprintf("%s/raw/%s/%s", cleanURL, branch, filePath)
	}

	rawURL = buildProxyURL(rawURL, cfg.Proxy, cfg.ProxyURL)
	downloadFile(rawURL, dest, cfg.AuthToken)
}

func getRemoteDefaultBranch(repoURL string, env []string) string {
	fmt.Printf("Detecting remote default branch: %s\n", repoURL)
	cmd := exec.Command("git", "ls-remote", "--symref", repoURL, "HEAD")
	cmd.Env = env
	out, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			parts := strings.Fields(line)
			if len(parts) >= 2 && parts[0] == "ref:" && strings.Contains(parts[1], "refs/heads/") {
				branch := strings.TrimPrefix(parts[1], "refs/heads/")
				fmt.Printf("Detected branch: %s\n", branch)
				return branch
			}
		}
	}
	fmt.Println("Failed to detect, using 'main'")
	return "main"
}

func buildProxyURL(url string, proxyType string, proxyURL string) string {
	if proxyType == "" || proxyType == "none" {
		return url
	}
	base := ""
	if proxyType == "ghproxy" {
		base = "https://gh-proxy.com/"
	} else if proxyType == "mirror" {
		base = "https://mirror.ghproxy.com/"
	} else if proxyType == "custom" && proxyURL != "" {
		base = strings.TrimSuffix(proxyURL, "/") + "/"
	}

	if base != "" && strings.HasPrefix(url, "http") {
		return base + url
	}
	return url
}

func downloadFile(url, dest, authToken string) {
	fmt.Printf("Downloading: %s\n", url)
	fmt.Printf("To: %s\n", dest)

	parentDir := filepath.Dir(dest)
	if parentDir != "" {
		os.MkdirAll(parentDir, 0755)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Download prep failed: %v\n", err)
		os.Exit(1)
	}
	if authToken != "" {
		req.Header.Set("Authorization", "token "+authToken)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; reposync)")

	client := &http.Client{Timeout: 300 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Download request failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Printf("Download failed, HTTP code: %d\n", resp.StatusCode)
		os.Exit(1)
	}

	out, err := os.Create(dest)
	if err != nil {
		fmt.Printf("Failed to create file: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	n, err := io.Copy(out, resp.Body)
	if err != nil {
		fmt.Printf("Failed to write data: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("File size: %d bytes\n", n)
	fmt.Println("Download completed")
}

func isRawFileURL(url string) bool {
	rawPatterns := []string{
		"raw.githubusercontent.com",
		"/raw/",
		"/-/raw/",
		"/blob/",
	}
	for _, p := range rawPatterns {
		if strings.Contains(url, p) {
			return true
		}
	}
	return false
}

func getRepoName(url string) string {
	u := strings.TrimSuffix(url, "/")
	u = strings.TrimSuffix(u, ".git")
	return filepath.Base(u)
}

func runCmd(args []string, dir string, env []string) {
	fmt.Printf(">> %s\n", strings.Join(args, " "))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Command failed: %v\n", err)
		os.Exit(1)
	}
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func isDirEmpty(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	_, err = f.Readdirnames(1)
	return err == io.EOF
}
