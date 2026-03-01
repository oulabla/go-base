package main

import (
	"flag"
	"fmt"
	"go/format"
	"os"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type RawItem struct {
	Usage string `yaml:"usage"`
	Group string `yaml:"group"`
	Type  string `yaml:"type"`
}

type Root struct {
	Config map[string]RawItem `yaml:"config"`
}

type Item struct {
	Key       string
	FieldName string
	Comment   string
	Type      string
}

type Group struct {
	Name  string
	Items []Item
}

type TemplateData struct {
	Groups []Group
}

var allowedTypes = map[string]bool{
	"int":      true,
	"string":   true,
	"duration": true,
	"bool":     true,
}

func main() {
	configPath := flag.String("config", "config/prod.yaml", "")
	templatePath := flag.String("template", "templates/keys.gen.go.tpl", "")
	outputPath := flag.String("output", "internal/config/keys.gen.go", "")
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	must(err)

	var root Root
	must(yaml.Unmarshal(data, &root))

	if len(root.Config) == 0 {
		panic("config section is empty")
	}

	// Проверка дубликатов
	seen := map[string]struct{}{}

	groupMap := map[string][]Item{}

	for key, val := range root.Config {
		if _, ok := seen[key]; ok {
			panic(fmt.Sprintf("duplicate key detected: %s", key))
		}
		seen[key] = struct{}{}

		if val.Group == "" {
			panic(fmt.Sprintf("group is empty for key: %s", key))
		}

		if !allowedTypes[val.Type] {
			panic(fmt.Sprintf("unsupported type '%s' for key: %s", val.Type, key))
		}

		groupMap[val.Group] = append(groupMap[val.Group], Item{
			Key:       key,
			FieldName: toCamel(key),
			Comment:   val.Usage,
			Type:      val.Type,
		})
	}

	var groups []Group

	for name, items := range groupMap {
		sort.Slice(items, func(i, j int) bool {
			return items[i].Key < items[j].Key
		})

		groups = append(groups, Group{
			Name:  name,
			Items: items,
		})
	}

	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Name < groups[j].Name
	})

	tplBytes, err := os.ReadFile(*templatePath)
	must(err)

	tpl := template.Must(template.New("keys").Parse(string(tplBytes)))

	var builder strings.Builder

	must(tpl.Execute(&builder, TemplateData{
		Groups: groups,
	}))

	formatted, err := format.Source([]byte(builder.String()))
	must(err)

	must(os.WriteFile(*outputPath, formatted, 0644))

	fmt.Println("Config generated successfully")
}

func toCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = strings.ToTitle(parts[i])
	}
	return strings.Join(parts, "")
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
