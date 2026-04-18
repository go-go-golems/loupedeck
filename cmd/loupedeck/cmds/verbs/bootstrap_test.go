package verbs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRepositoriesFromArgsParsesRepeatedFlag(t *testing.T) {
	cwd := t.TempDir()
	repoA := filepath.Join(cwd, "repo-a")
	repoB := filepath.Join(cwd, "repo-b")
	if err := os.MkdirAll(repoA, 0o755); err != nil {
		t.Fatalf("mkdir repoA: %v", err)
	}
	if err := os.MkdirAll(repoB, 0o755); err != nil {
		t.Fatalf("mkdir repoB: %v", err)
	}

	repos, err := repositoriesFromArgs([]string{"--verbs-repository", repoA, "verbs", "documented", "configure", "--verbs-repository=" + repoB}, cwd)
	if err != nil {
		t.Fatalf("repositoriesFromArgs: %v", err)
	}
	if len(repos) != 2 {
		t.Fatalf("expected 2 repositories, got %#v", repos)
	}
}

func TestRepositoriesFromEnvUsesPathList(t *testing.T) {
	cwd := t.TempDir()
	repoA := filepath.Join(cwd, "repo-a")
	repoB := filepath.Join(cwd, "repo-b")
	if err := os.MkdirAll(repoA, 0o755); err != nil {
		t.Fatalf("mkdir repoA: %v", err)
	}
	if err := os.MkdirAll(repoB, 0o755); err != nil {
		t.Fatalf("mkdir repoB: %v", err)
	}
	t.Setenv(VerbRepositoriesEnvVar, strings.Join([]string{repoA, repoB, repoA}, string(os.PathListSeparator)))

	repos, err := repositoriesFromEnv(cwd)
	if err != nil {
		t.Fatalf("repositoriesFromEnv: %v", err)
	}
	if len(repos) != 3 {
		t.Fatalf("expected raw env repositories before dedupe, got %#v", repos)
	}

	bootstrap, err := DiscoverBootstrap(nil)
	if err != nil {
		t.Fatalf("DiscoverBootstrap: %v", err)
	}
	count := 0
	for _, repo := range bootstrap.Repositories {
		if !repo.Embedded && (repo.RootDir == repoA || repo.RootDir == repoB) {
			count++
		}
	}
	if count != 2 {
		t.Fatalf("expected deduped env repositories in bootstrap, got %#v", bootstrap.Repositories)
	}
}

func TestLoadConfigRepositoriesFromXDGConfig(t *testing.T) {
	home := t.TempDir()
	xdg := filepath.Join(home, ".config")
	repoPath := filepath.Join(home, "repo")
	if err := os.MkdirAll(repoPath, 0o755); err != nil {
		t.Fatalf("mkdir repoPath: %v", err)
	}
	configDir := filepath.Join(xdg, "loupedeck")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir configDir: %v", err)
	}
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("verbs:\n  repositories:\n    - name: team\n      path: ../../repo\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	oldHome := os.Getenv("HOME")
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Setenv("HOME", home)
	t.Setenv("XDG_CONFIG_HOME", xdg)
	_ = oldHome
	_ = oldXDG

	repos, err := loadConfigRepositories(context.Background())
	if err != nil {
		t.Fatalf("loadConfigRepositories: %v", err)
	}
	if len(repos) != 1 {
		t.Fatalf("expected 1 repository, got %#v", repos)
	}
	if repos[0].Name != "team" || repos[0].RootDir != repoPath {
		t.Fatalf("unexpected repository %#v", repos[0])
	}
}

func TestCollectDiscoveredVerbsRejectsDuplicatePaths(t *testing.T) {
	repoA := t.TempDir()
	repoB := t.TempDir()
	source := []byte(`
__package__({ name: "dup" });
function configure() { return { ok: true }; }
__verb__("configure", { name: "configure", parents: ["dup"] });
`)
	if err := os.WriteFile(filepath.Join(repoA, "entry.js"), source, 0o644); err != nil {
		t.Fatalf("write repoA: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoB, "entry.js"), source, 0o644); err != nil {
		t.Fatalf("write repoB: %v", err)
	}
	repositories, err := scanRepositories(Bootstrap{Repositories: []Repository{
		{Name: "repo-a", Source: "test", RootDir: repoA},
		{Name: "repo-b", Source: "test", RootDir: repoB},
	}})
	if err != nil {
		t.Fatalf("scanRepositories: %v", err)
	}
	_, err = collectDiscoveredVerbs(repositories)
	if err == nil {
		t.Fatal("expected duplicate path error")
	}
	if !strings.Contains(err.Error(), `duplicate jsverb path "dup configure"`) {
		t.Fatalf("unexpected error %v", err)
	}
}
