package verbs

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dop251/goja_nodejs/require"
	glazedconfig "github.com/go-go-golems/glazed/pkg/config"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
	"github.com/go-go-golems/loupedeck/examples"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	VerbRepositoriesEnvVar = "LOUPEDECK_VERB_REPOSITORIES"
	VerbRepositoryFlag     = "verbs-repository"
)

type Bootstrap struct {
	Repositories []Repository
}

type Repository struct {
	Name       string
	Source     string
	SourceRef  string
	RootDir    string
	EmbeddedFS fs.FS
	Embedded   bool
	EmbeddedAt string
}

type scannedRepository struct {
	Repository Repository
	Registry   *jsverbs.Registry
}

type discoveredVerb struct {
	Repository scannedRepository
	Verb       *jsverbs.VerbSpec
}

type appConfig struct {
	Verbs verbsConfig `yaml:"verbs"`
}

type verbsConfig struct {
	Repositories []repositorySpec `yaml:"repositories"`
}

type repositorySpec struct {
	Name    string `yaml:"name,omitempty"`
	Path    string `yaml:"path"`
	Enabled *bool  `yaml:"enabled,omitempty"`
}

func DiscoverBootstrap(args []string) (Bootstrap, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Bootstrap{}, fmt.Errorf("resolve cwd: %w", err)
	}
	cliRepos, err := repositoriesFromArgs(args, cwd)
	if err != nil {
		return Bootstrap{}, err
	}
	return discoverBootstrap(cwd, cliRepos)
}

func DiscoverBootstrapFromCommand(cmd *cobra.Command) (Bootstrap, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return Bootstrap{}, fmt.Errorf("resolve cwd: %w", err)
	}
	cliRepos, err := repositoriesFromCommand(cmd, cwd)
	if err != nil {
		return Bootstrap{}, err
	}
	return discoverBootstrap(cwd, cliRepos)
}

func discoverBootstrap(cwd string, cliRepos []Repository) (Bootstrap, error) {
	repositories := []Repository{builtinRepository()}
	seen := map[string]struct{}{repositoryIdentity(builtinRepository()): {}}

	configRepos, err := loadConfigRepositories(context.Background())
	if err != nil {
		return Bootstrap{}, err
	}
	for _, repo := range configRepos {
		if err := appendRepository(&repositories, seen, repo); err != nil {
			return Bootstrap{}, err
		}
	}

	envRepos, err := repositoriesFromEnv(cwd)
	if err != nil {
		return Bootstrap{}, err
	}
	for _, repo := range envRepos {
		if err := appendRepository(&repositories, seen, repo); err != nil {
			return Bootstrap{}, err
		}
	}

	for _, repo := range cliRepos {
		if err := appendRepository(&repositories, seen, repo); err != nil {
			return Bootstrap{}, err
		}
	}

	return Bootstrap{Repositories: repositories}, nil
}

func builtinRepository() Repository {
	return Repository{
		Name:       "builtin",
		Source:     "embedded",
		SourceRef:  "builtin repository",
		EmbeddedFS: examples.ScriptsFS,
		Embedded:   true,
		EmbeddedAt: "js",
	}
}

func appendRepository(repositories *[]Repository, seen map[string]struct{}, repo Repository) error {
	identity := repositoryIdentity(repo)
	if _, ok := seen[identity]; ok {
		return nil
	}
	seen[identity] = struct{}{}
	*repositories = append(*repositories, repo)
	return nil
}

func repositoryIdentity(repo Repository) string {
	if repo.Embedded {
		return "embedded:" + repo.Name + ":" + repo.EmbeddedAt
	}
	return "path:" + filepath.Clean(repo.RootDir)
}

func loadConfigRepositories(ctx context.Context) ([]Repository, error) {
	plan := glazedconfig.NewPlan(
		glazedconfig.WithLayerOrder(glazedconfig.LayerSystem, glazedconfig.LayerUser),
		glazedconfig.WithDedupePaths(),
	).Add(
		glazedconfig.SystemAppConfig("loupedeck").Named("system-app-config"),
		glazedconfig.XDGAppConfig("loupedeck").Named("xdg-app-config"),
		glazedconfig.HomeAppConfig("loupedeck").Named("home-app-config"),
	)

	files, _, err := plan.Resolve(ctx)
	if err != nil {
		return nil, fmt.Errorf("resolve loupedeck app config: %w", err)
	}

	ret := []Repository{}
	for _, file := range files {
		repos, err := loadRepositoriesFromConfigFile(file.Path)
		if err != nil {
			return nil, err
		}
		ret = append(ret, repos...)
	}
	return ret, nil
}

func loadRepositoriesFromConfigFile(path string) ([]Repository, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read app config %s: %w", path, err)
	}
	cfg := &appConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse app config %s: %w", path, err)
	}
	baseDir := filepath.Dir(path)
	ret := []Repository{}
	for _, spec := range cfg.Verbs.Repositories {
		if spec.Enabled != nil && !*spec.Enabled {
			continue
		}
		normalized, err := normalizeFilesystemRepositoryPath(spec.Path, baseDir)
		if err != nil {
			return nil, fmt.Errorf("config repository %q in %s: %w", spec.Path, path, err)
		}
		name := strings.TrimSpace(spec.Name)
		if name == "" {
			name = filepath.Base(normalized)
		}
		ret = append(ret, Repository{
			Name:      name,
			Source:    "config",
			SourceRef: path,
			RootDir:   normalized,
		})
	}
	return ret, nil
}

func repositoriesFromEnv(cwd string) ([]Repository, error) {
	value := strings.TrimSpace(os.Getenv(VerbRepositoriesEnvVar))
	if value == "" {
		return nil, nil
	}
	ret := []Repository{}
	for _, raw := range filepath.SplitList(value) {
		normalized, err := normalizeFilesystemRepositoryPath(raw, cwd)
		if err != nil {
			return nil, fmt.Errorf("environment repository %q: %w", raw, err)
		}
		ret = append(ret, Repository{
			Name:      filepath.Base(normalized),
			Source:    "env",
			SourceRef: VerbRepositoriesEnvVar,
			RootDir:   normalized,
		})
	}
	return ret, nil
}

func repositoriesFromArgs(args []string, cwd string) ([]Repository, error) {
	paths := []string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--"+VerbRepositoryFlag:
			if i+1 < len(args) {
				paths = append(paths, args[i+1])
				i++
			}
		case strings.HasPrefix(arg, "--"+VerbRepositoryFlag+"="):
			paths = append(paths, strings.TrimPrefix(arg, "--"+VerbRepositoryFlag+"="))
		}
	}
	return repositoriesFromCLIPaths(paths, cwd)
}

func repositoriesFromCommand(cmd *cobra.Command, cwd string) ([]Repository, error) {
	if cmd == nil {
		return nil, nil
	}
	paths := []string{}
	if cmd.Flags().Lookup(VerbRepositoryFlag) != nil {
		flagPaths, err := cmd.Flags().GetStringArray(VerbRepositoryFlag)
		if err != nil {
			return nil, err
		}
		paths = append(paths, flagPaths...)
	}
	if len(paths) == 0 && cmd.InheritedFlags().Lookup(VerbRepositoryFlag) != nil {
		flagPaths, err := cmd.InheritedFlags().GetStringArray(VerbRepositoryFlag)
		if err != nil {
			return nil, err
		}
		paths = append(paths, flagPaths...)
	}
	return repositoriesFromCLIPaths(paths, cwd)
}

func repositoriesFromCLIPaths(paths []string, cwd string) ([]Repository, error) {
	ret := []Repository{}
	for _, raw := range paths {
		normalized, err := normalizeFilesystemRepositoryPath(raw, cwd)
		if err != nil {
			return nil, fmt.Errorf("CLI repository %q: %w", raw, err)
		}
		ret = append(ret, Repository{
			Name:      filepath.Base(normalized),
			Source:    "cli",
			SourceRef: "--" + VerbRepositoryFlag,
			RootDir:   normalized,
		})
	}
	return ret, nil
}

func normalizeFilesystemRepositoryPath(path string, baseDir string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("repository path is empty")
	}
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}
	path = filepath.Clean(path)
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}
	return path, nil
}

func scanRepositories(bootstrap Bootstrap) ([]scannedRepository, error) {
	opts := jsverbs.DefaultScanOptions()
	opts.IncludePublicFunctions = false

	ret := make([]scannedRepository, 0, len(bootstrap.Repositories))
	for _, repo := range bootstrap.Repositories {
		var (
			registry *jsverbs.Registry
			err      error
		)
		if repo.Embedded {
			registry, err = jsverbs.ScanFS(repo.EmbeddedFS, repo.EmbeddedAt, opts)
		} else {
			registry, err = jsverbs.ScanDir(repo.RootDir, opts)
		}
		if err != nil {
			return nil, fmt.Errorf("scan repository %s: %w", repo.Name, err)
		}
		ret = append(ret, scannedRepository{Repository: repo, Registry: registry})
	}
	return ret, nil
}

func collectDiscoveredVerbs(repositories []scannedRepository) ([]discoveredVerb, error) {
	seen := map[string]discoveredVerb{}
	ret := []discoveredVerb{}
	for _, repo := range repositories {
		for _, verb := range repo.Registry.Verbs() {
			key := verb.FullPath()
			candidate := discoveredVerb{Repository: repo, Verb: verb}
			if prev, ok := seen[key]; ok {
				return nil, fmt.Errorf("duplicate jsverb path %q from %s and %s", key, discoveredVerbSource(prev), discoveredVerbSource(candidate))
			}
			seen[key] = candidate
			ret = append(ret, candidate)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Verb.FullPath() < ret[j].Verb.FullPath()
	})
	return ret, nil
}

func discoveredVerbSource(verb discoveredVerb) string {
	if verb.Verb == nil || verb.Verb.File == nil {
		return verb.Repository.Repository.Name
	}
	if verb.Verb.File.AbsPath != "" {
		return fmt.Sprintf("%s (%s)", verb.Repository.Repository.Name, verb.Verb.File.AbsPath)
	}
	return fmt.Sprintf("%s (%s)", verb.Repository.Repository.Name, verb.Verb.File.RelPath)
}

func (r scannedRepository) runtimeOptions() []engine.Option {
	opts := []engine.Option{engine.WithRequireOptions(require.WithLoader(r.Registry.RequireLoader()))}
	if r.Repository.Embedded {
		return opts
	}
	folders := []string{r.Repository.RootDir, filepath.Join(r.Repository.RootDir, "node_modules")}
	parent := filepath.Dir(r.Repository.RootDir)
	if parent != r.Repository.RootDir {
		folders = append(folders, parent, filepath.Join(parent, "node_modules"))
	}
	return append(opts, engine.WithRequireOptions(require.WithGlobalFolders(folders...)))
}
