package main

import (
	"bytes"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/yaml.v2"

	"github.com/serenize/snaker"
)

var (
	flagPackageName = kingpin.Flag("pkg", "Output package name").Default("main").String()
	flagForce       = kingpin.Flag("force", "Force rebuild.").Short('f').Bool()
	flagYamlDefFile = kingpin.Arg("path", "Input YAML definitions file path").ExistingFile()
)

type yamlDef struct {
	Imports []string
	Kinds   map[string]*yamlTypeDef
}

type yamlTypeDef struct {
	Extends []string
	Fields  map[string]string
}

func normalizeFieldId(t string) string {
	switch t {
	case "urn":
		return "URN"
	default:
		return snaker.SnakeToCamel(t)
	}
}

func main() {
	kingpin.Parse()

	// Generate output file name from input file name
	outputFileName := (*flagYamlDefFile)[0:len(*flagYamlDefFile)-len(filepath.Ext(*flagYamlDefFile))] + ".go"

	// Get input file stat
	inputStat, err := os.Stat(*flagYamlDefFile)
	if err != nil {
		panic(err)
	}

	// Does output file exist?
	if !*flagForce {
		if s, err := os.Stat(outputFileName); err == nil {
			// Check output file time against input file time
			if !inputStat.ModTime().After(s.ModTime()) {
				// Output already up to date!
				return
			}
		}
	}

	fmt.Println(filepath.Base(*flagYamlDefFile))

	// Read YAML def
	ymlBytes, err := ioutil.ReadFile(*flagYamlDefFile)
	if err != nil {
		panic(err)
	}

	// Unmarshal YAML def
	var ymlDef yamlDef
	yaml.Unmarshal(ymlBytes, &ymlDef)

	b := new(bytes.Buffer)

	// Write package name
	b.WriteString(fmt.Sprintf("package %s\n", *flagPackageName))

	// Write imports
	if ymlDef.Imports != nil && len(ymlDef.Imports) > 0 {
		b.WriteString("import (\n")
		for _, i := range ymlDef.Imports {
			b.WriteString(fmt.Sprintf("%q", i))
		}
		b.WriteString(")\n")
	}

	// Write types
	registeredTypes := make(map[string]interface{})
	//var ensureType func(string)
	generateType := func(name string, definition *yamlTypeDef) {
		//id := snaker.SnakeToCamel(name)

		/*for _, fieldType := range definition.Fields {
			ensureType(fieldType)
		}*/

		b.WriteString(fmt.Sprintf("\ntype %s struct {\n", name))

		// Extends (as configured)
		for _, t := range definition.Extends {
			b.WriteString(t + "\n")
		}

		// Fields (sorted alphabetically)
		var keys []string
		for k, _ := range definition.Fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, fieldName := range keys {
			fieldType := definition.Fields[fieldName]
			fieldId := normalizeFieldId(fieldName)
			b.WriteString(fmt.Sprintf("%s %s `json:%q`\n", fieldId, fieldType, fieldName))
		}
		b.WriteString("}\n")
	}
	/*ensureType = func(name string) {
		if _, ok := registeredTypes[name]; ok {
			// already generated
			return
		}

		typeDef, ok := ymlDef.Kinds[name]
		if ok {
			registeredTypes[name] = nil
			generateType(name, typeDef)
		}
	}*/

	var keys []string
	for k, _ := range ymlDef.Kinds {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		typeDef := ymlDef.Kinds[name]
		if _, ok := registeredTypes[name]; !ok {
			registeredTypes[name] = nil
			generateType(name, typeDef)
		}
	}

	// Write out formatted source code
	//fmt.Println(string(b.Bytes()))
	fb, err := format.Source(b.Bytes())
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile(outputFileName, fb, 0644)
}
