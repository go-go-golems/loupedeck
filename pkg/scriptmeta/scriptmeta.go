package scriptmeta

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dop251/goja_nodejs/require"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/runner"
	"github.com/go-go-golems/glazed/pkg/cmds/values"
	"github.com/go-go-golems/go-go-goja/engine"
	"github.com/go-go-golems/go-go-goja/pkg/jsdoc/batch"
	jsdocexport "github.com/go-go-golems/go-go-goja/pkg/jsdoc/export"
	jsdocmodel "github.com/go-go-golems/go-go-goja/pkg/jsdoc/model"
	"github.com/go-go-golems/go-go-goja/pkg/jsverbs"
)

type Target struct {
	Path      string
	RootDir   string
	EntryFile string
	IsDir     bool
}

func ResolveTarget(path string) (*Target, error) {
	if strings.TrimSpace(path) == "" {
		return nil, fmt.Errorf("target path is empty")
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve absolute path: %w", err)
	}
	info, err := os.Stat(absPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("stat %s: %w", absPath, err)
		}
		absPath, err = resolveScriptShorthand(absPath)
		if err != nil {
			return nil, err
		}
		info, err = os.Stat(absPath)
		if err != nil {
			return nil, fmt.Errorf("stat %s: %w", absPath, err)
		}
	}
	if info.IsDir() {
		return &Target{Path: absPath, RootDir: absPath, IsDir: true}, nil
	}
	return &Target{Path: absPath, RootDir: filepath.Dir(absPath), EntryFile: absPath}, nil
}

func resolveScriptShorthand(absPath string) (string, error) {
	if filepath.Ext(absPath) != "" {
		return "", fmt.Errorf("stat %s: %w", absPath, os.ErrNotExist)
	}
	matches := []string{}
	for _, pattern := range []string{absPath + "*.js", absPath + "*.cjs"} {
		found, err := filepath.Glob(pattern)
		if err != nil {
			return "", fmt.Errorf("glob %s: %w", pattern, err)
		}
		for _, match := range found {
			info, err := os.Stat(match)
			if err == nil && !info.IsDir() {
				matches = append(matches, match)
			}
		}
	}
	sort.Strings(matches)
	switch len(matches) {
	case 0:
		return "", fmt.Errorf("stat %s: %w", absPath, os.ErrNotExist)
	case 1:
		return matches[0], nil
	default:
		return "", fmt.Errorf("script path %q is ambiguous: %s", absPath, strings.Join(matches, ", "))
	}
}

func ScanVerbRegistry(path string) (*Target, *jsverbs.Registry, error) {
	target, err := ResolveTarget(path)
	if err != nil {
		return nil, nil, err
	}
	opts := jsverbs.DefaultScanOptions()
	opts.IncludePublicFunctions = false
	registry, err := jsverbs.ScanDir(target.RootDir, opts)
	if err != nil {
		return nil, nil, err
	}
	return target, registry, nil
}

func EntryVerbs(target *Target, registry *jsverbs.Registry) []*jsverbs.VerbSpec {
	if registry == nil {
		return nil
	}
	ret := []*jsverbs.VerbSpec{}
	for _, verb := range registry.Verbs() {
		if target != nil && target.EntryFile != "" {
			if verb.File == nil || verb.File.AbsPath != target.EntryFile {
				continue
			}
		}
		ret = append(ret, verb)
	}
	return ret
}

func FindVerb(target *Target, registry *jsverbs.Registry, selector string) (*jsverbs.VerbSpec, error) {
	selector = strings.TrimSpace(selector)
	if registry == nil {
		return nil, fmt.Errorf("registry is nil")
	}
	entryVerbs := EntryVerbs(target, registry)
	if selector != "" {
		if target == nil || target.EntryFile == "" {
			if verb, ok := registry.Verb(selector); ok {
				return verb, nil
			}
		}
		matches := []*jsverbs.VerbSpec{}
		for _, verb := range entryVerbs {
			if verb.FullPath() == selector || verb.Name == selector || verb.FunctionName == selector || strings.HasSuffix(verb.FullPath(), " "+selector) {
				matches = append(matches, verb)
			}
		}
		if len(matches) == 1 {
			return matches[0], nil
		}
		if len(matches) > 1 {
			paths := make([]string, 0, len(matches))
			for _, verb := range matches {
				paths = append(paths, verb.FullPath())
			}
			sort.Strings(paths)
			return nil, fmt.Errorf("verb selector %q is ambiguous: %s", selector, strings.Join(paths, ", "))
		}
		return nil, fmt.Errorf("verb %q not found", selector)
	}
	if len(entryVerbs) == 1 {
		return entryVerbs[0], nil
	}
	if len(entryVerbs) == 0 {
		return nil, fmt.Errorf("no explicit jsverbs found for %s", target.Path)
	}
	paths := make([]string, 0, len(entryVerbs))
	for _, verb := range entryVerbs {
		paths = append(paths, verb.FullPath())
	}
	sort.Strings(paths)
	return nil, fmt.Errorf("multiple verbs found; specify one of: %s", strings.Join(paths, ", "))
}

func EngineOptionsForTarget(target *Target, registry *jsverbs.Registry) ([]engine.Option, error) {
	if target == nil {
		return nil, fmt.Errorf("target is nil")
	}
	opts := []engine.Option{}
	if target.EntryFile != "" {
		opts = append(opts, engine.WithModuleRootsFromScript(target.EntryFile, engine.DefaultModuleRootsOptions()))
	} else {
		folders := []string{target.RootDir, filepath.Join(target.RootDir, "node_modules")}
		parent := filepath.Dir(target.RootDir)
		if parent != target.RootDir {
			folders = append(folders, parent, filepath.Join(parent, "node_modules"))
		}
		opts = append(opts, engine.WithRequireOptions(require.WithGlobalFolders(folders...)))
	}
	if registry != nil {
		opts = append(opts, engine.WithRequireOptions(require.WithLoader(registry.RequireLoader())))
	}
	return opts, nil
}

type descriptionOnlyCommand struct {
	*cmds.CommandDescription
}

func ParseVerbValues(desc *cmds.CommandDescription, configFiles []string, valuesJSON string) (*values.Values, error) {
	if desc == nil {
		return nil, fmt.Errorf("command description is nil")
	}
	parseOpts := []runner.ParseOption{}
	if len(configFiles) > 0 {
		parseOpts = append(parseOpts, runner.WithConfigFiles(configFiles...))
	}
	if strings.TrimSpace(valuesJSON) != "" {
		sections, err := parseValuesJSON(valuesJSON)
		if err != nil {
			return nil, err
		}
		parseOpts = append(parseOpts, runner.WithValuesForSections(sections))
	}
	return runner.ParseCommandValues(&descriptionOnlyCommand{CommandDescription: desc}, parseOpts...)
}

func parseValuesJSON(raw string) (map[string]map[string]interface{}, error) {
	var sectioned map[string]map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &sectioned); err == nil {
		return sectioned, nil
	}
	var flat map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &flat); err != nil {
		return nil, fmt.Errorf("parse verb values json: %w", err)
	}
	return map[string]map[string]interface{}{"default": flat}, nil
}

func BuildDocStore(ctx context.Context, path string) (*Target, *jsdocmodel.DocStore, error) {
	target, err := ResolveTarget(path)
	if err != nil {
		return nil, nil, err
	}
	inputs := []batch.InputFile{}
	if target.EntryFile != "" {
		inputs = append(inputs, batch.InputFile{Path: target.EntryFile})
	} else {
		err = filepath.WalkDir(target.RootDir, func(filePath string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				if d.Name() == "node_modules" || strings.HasPrefix(d.Name(), ".") {
					if filePath == target.RootDir {
						return nil
					}
					return filepath.SkipDir
				}
				return nil
			}
			ext := strings.ToLower(filepath.Ext(filePath))
			if ext != ".js" && ext != ".cjs" {
				return nil
			}
			inputs = append(inputs, batch.InputFile{Path: filePath})
			return nil
		})
		if err != nil {
			return nil, nil, err
		}
	}
	if err != nil {
		return nil, nil, err
	}
	result, err := batch.BuildStore(ctx, inputs, batch.BatchOptions{ContinueOnError: false})
	if err != nil {
		return nil, nil, err
	}
	return target, result.Store, nil
}

func ExportDocStore(ctx context.Context, store *jsdocmodel.DocStore, format string) ([]byte, error) {
	opts := jsdocexport.Options{Indent: "  "}
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "json":
		opts.Format = jsdocexport.FormatJSON
	case "markdown", "md":
		opts.Format = jsdocexport.FormatMarkdown
	default:
		return nil, fmt.Errorf("unsupported doc format %q", format)
	}
	var b strings.Builder
	if err := jsdocexport.Export(ctx, store, &b, opts); err != nil {
		return nil, err
	}
	return []byte(b.String()), nil
}
