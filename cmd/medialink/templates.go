package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"math"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/dustin/go-humanize"
	"golang.org/x/exp/utf8string"
)

//go:embed *.tpl
var templateFS embed.FS

const templateFileGlobFilter = "*.tpl"

var (
	compactNumUnits = []string{"", "k", "M"}

	tplFuncMap = template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"color": func(num int) string {
			return string(runeIrcColor) + strconv.Itoa(num)
		},
		"bcolor": func(fgNum, bgNum int) string {
			return string(runeIrcColor) + strconv.Itoa(fgNum) + "," + strconv.Itoa(bgNum)
		},
		"bold": func() string {
			return string(runeIrcBold)
		},
		"italic": func() string {
			return string(runeIrcItalic)
		},
		"reset": func() string {
			return string(runeIrcReset) + string(runeIrcColor)
		},
		"reverse": func() string {
			return string(runeIrcReverse)
		},
		"underline": func() string {
			return string(runeIrcUnderline)
		},
		"urlencode": func(s string) string {
			return url.QueryEscape(s)
		},
		"yesno": func(yes string, no string, value bool) string {
			if value {
				return yes
			}

			return no
		},
		"excerpt": func(maxLength uint16, text string) string {
			utf8str := utf8string.NewString(text)
			if utf8str.RuneCount() > int(maxLength) {
				return utf8str.Slice(0, int(maxLength-1)) + "\u2026"
			}
			return text
		},
		"comma": func(num uint64) string {
			return humanize.Comma(int64(num))
		},
		"compactnum": func(num uint64) string {
			// 1 => 0
			// 1000 => 1
			// 1000000 => 2
			log10 := math.Floor(math.Log10(float64(num)) / 3)

			// Cut to available units
			cut := int(math.Min(float64(len(compactNumUnits)-1), log10))

			numf := float64(num)
			numf /= math.Pow10(cut * 3)

			// Rounding
			numf = math.Floor((numf*10)+.5) / 10
			if numf >= 1000 {
				numf /= 1000
				if cut < len(compactNumUnits)-1 {
					cut++
				}
			}

			unit := compactNumUnits[cut]
			f := "%.1f%s"
			if numf-math.Floor(numf) < 0.05 {
				f = "%.0f%s"
			}

			return fmt.Sprintf(f, numf, unit)
		},
		"ago": func(t time.Time) string {
			return humanize.Time(t)
		},
		"size": func(s uint64) string {
			return humanize.Bytes(s)
		},
	}

	rxInsignificantWhitespace = regexp.MustCompile(`\s+`)
)

type templateCollection struct {
	template *template.Template
}

func loadTemplates() (*templateCollection, error) {
	t := template.New("").
		Funcs(tplFuncMap)

	// create a list of templates we look for
	templateFileEntries, err := fs.Glob(templateFS, templateFileGlobFilter)
	if err != nil {
		return nil, err
	}
	loadedTemplates := map[string]bool{}
	for _, templateFileEntry := range templateFileEntries {
		loadedTemplates[templateFileEntry] = false
	}

	// load customized templates if any exist
	templateFileNames, err := filepath.Glob(templateFileGlobFilter)
	if err != nil {
		return nil, err
	}
	if len(templateFileNames) > 0 {
		actualFileNamesToLoad := []string{}
		for _, templateFileName := range templateFileNames {
			if _, ok := loadedTemplates[templateFileName]; ok {
				loadedTemplates[templateFileName] = true
				actualFileNamesToLoad = append(actualFileNamesToLoad, templateFileName)
			}
		}
		t, err = t.ParseFiles(actualFileNamesToLoad...)
		if err != nil {
			return nil, err
		}
	}

	// load default templates for those where none were supplied by the user
	defaultTemplatesToLoad := []string{}
	for templateFileName, userTemplateLoaded := range loadedTemplates {
		// skip templates supplied by user
		if userTemplateLoaded {
			continue
		}
		defaultTemplatesToLoad = append(defaultTemplatesToLoad, templateFileName)
	}
	t, err = t.ParseFS(templateFS, defaultTemplatesToLoad...)
	if err != nil {
		return nil, err
	}

	return &templateCollection{t}, nil
}

func (c *templateCollection) tplString(name string, data interface{}) (string, error) {
	w := new(bytes.Buffer)
	if err := c.template.ExecuteTemplate(w, name, data); err != nil {
		return "", err
	}
	s := w.String()
	s = rxInsignificantWhitespace.ReplaceAllString(s, " ")
	s = strings.Trim(s, " ")
	log.Printf("tplString(%v): %s", name, s)
	return s, nil
}
