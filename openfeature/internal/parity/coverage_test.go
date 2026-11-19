package parity

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

type gateScenario struct {
	ID           string   `json:"id"`
	Requirements []string `json:"requirements"`
	Deviation    string   `json:"deviation"`
}

func gateScenarios(t *testing.T) map[string]gateScenario {
	t.Helper()
	files, err := filepath.Glob(filepath.Join("scenarios", "*.json"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no scenarios: %v", err)
	}
	out := map[string]gateScenario{}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var sc gateScenario
		if err := json.Unmarshal(data, &sc); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		out[strings.TrimSuffix(filepath.Base(f), ".json")] = sc
	}
	return out
}

func gateSurface(t *testing.T) map[string]surfaceEntry {
	t.Helper()
	data, err := os.ReadFile("surface.json")
	if err != nil {
		t.Fatal(err)
	}
	var s map[string]surfaceEntry
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatal(err)
	}
	return s
}

func openfeatureTestFiles(t *testing.T) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(filepath.Join("..", ".."), func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == "node_modules" {
			return filepath.SkipDir
		}
		if !d.IsDir() && strings.HasSuffix(p, "_test.go") {
			files = append(files, p)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func parseGo(t *testing.T, path string) *ast.File {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("%s: %v", path, err)
	}
	return f
}

func TestEveryStaticContextRequirementCovered(t *testing.T) {
	covered := map[string]bool{}
	for _, sc := range gateScenarios(t) {
		for _, r := range sc.Requirements {
			covered[r] = true
		}
	}
	for _, e := range gateSurface(t) {
		if e.Absent != "" {
			covered[e.Absent] = true
		}
	}
	for _, r := range strings.Fields("1.3.2.1 1.4.2.1 1.7.2.1 3.2.2.1 3.2.2.2 3.2.2.3 3.2.2.4 3.2.4.1 3.2.4.2 3.3.2.1 4.3.3.1 5.3.4.1 5.3.4.2 5.3.4.3 6.1.2.1") {
		if !covered[r] {
			t.Errorf("requirement %s is covered by no scenario and no absent surface entry", r)
		}
	}
}

func TestEveryGoldenHasAScenarioAndViceVersa(t *testing.T) {
	scenarios := gateScenarios(t)
	goldens, err := filepath.Glob(filepath.Join("golden", "*.trace.json"))
	if err != nil {
		t.Fatal(err)
	}
	have := map[string]bool{}
	for _, g := range goldens {
		id := strings.TrimSuffix(filepath.Base(g), ".trace.json")
		have[id] = true
		if _, ok := scenarios[id]; !ok {
			t.Errorf("golden %s has no scenario", id)
		}
	}
	for id := range scenarios {
		if !have[id] {
			t.Errorf("scenario %s has no golden", id)
		}
	}
}

func testFuncs(t *testing.T) map[string]bool {
	t.Helper()
	funcs := map[string]bool{}
	files, err := filepath.Glob(filepath.Join("..", "..", "*_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range files {
		for _, d := range parseGo(t, p).Decls {
			if fn, ok := d.(*ast.FuncDecl); ok && fn.Recv == nil {
				funcs[fn.Name.Name] = true
			}
		}
	}
	return funcs
}

func TestOnlyAllowedDeviations(t *testing.T) {
	allowed := regexp.MustCompile(`^D[1-8]$`)
	funcs := testFuncs(t)
	for id, sc := range gateScenarios(t) {
		if sc.Deviation == "" {
			continue
		}
		if !allowed.MatchString(sc.Deviation) {
			t.Errorf("scenario %s: deviation %q is not in D1-D8", id, sc.Deviation)
		}
		if sc.Deviation == "D2" || sc.Deviation == "D3" {
			name := "TestDeviation_" + strings.ReplaceAll(id, "-", "_")
			if !funcs[name] {
				t.Errorf("scenario %s (%s): Go-only assertion %s not found in openfeature/*_test.go", id, sc.Deviation, name)
			}
		}
	}
	for member, e := range gateSurface(t) {
		if e.Deviation != "" && !allowed.MatchString(e.Deviation) {
			t.Errorf("surface %s: deviation %q is not in D1-D8", member, e.Deviation)
		}
	}
}

func TestNoSkips(t *testing.T) {
	files := openfeatureTestFiles(t)
	if len(files) == 0 {
		t.Fatal("no test files found")
	}
	for _, p := range files {
		ast.Inspect(parseGo(t, p), func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if ok && (sel.Sel.Name == "Skip" || sel.Sel.Name == "SkipNow" || sel.Sel.Name == "Skipf") {
				t.Errorf("%s: %s is forbidden", p, sel.Sel.Name)
			}
			return true
		})
	}
}

func TestWasmExclusionsAreExactlyDeviationsAndMem(t *testing.T) {
	var want []string
	for id, sc := range gateScenarios(t) {
		if sc.Deviation != "" || strings.HasPrefix(id, "mem-") {
			want = append(want, id)
		}
	}
	var got []string
	for _, d := range parseGo(t, "jsparity_wasm_test.go").Decls {
		gd, ok := d.(*ast.GenDecl)
		if !ok {
			continue
		}
		for _, s := range gd.Specs {
			vs, ok := s.(*ast.ValueSpec)
			if !ok || len(vs.Names) != 1 || vs.Names[0].Name != "jsExcluded" {
				continue
			}
			for _, el := range vs.Values[0].(*ast.CompositeLit).Elts {
				k, err := strconv.Unquote(el.(*ast.KeyValueExpr).Key.(*ast.BasicLit).Value)
				if err != nil {
					t.Fatal(err)
				}
				got = append(got, k)
			}
		}
	}
	slices.Sort(want)
	slices.Sort(got)
	if !slices.Equal(want, got) {
		t.Errorf("jsExcluded = %v, want exactly the deviation and mem-* scenarios %v", got, want)
	}
}

func TestGoldensPinned(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("js", "package-lock.json"))
	if err != nil {
		t.Fatal(err)
	}
	var lock struct {
		Packages map[string]struct {
			Version string `json:"version"`
		} `json:"packages"`
	}
	if err := json.Unmarshal(data, &lock); err != nil {
		t.Fatal(err)
	}
	for pkg, want := range map[string]string{
		"node_modules/@openfeature/web-sdk": "1.10.0",
		"node_modules/@openfeature/core":    "1.12.0",
	} {
		if got := lock.Packages[pkg].Version; got != want {
			t.Errorf("%s resolves to %q, want %q", pkg, got, want)
		}
	}
}
