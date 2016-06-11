package wikipedia

import (
	"strings"

	"gopkg.in/neurosnap/sentences.v1/english"
)

func prepareSummary(title string, summary string) string {
	// Sentence tokenizer - English will work fine in most cases for now
	tokenizer, err := english.NewSentenceTokenizer(nil)
	if err != nil {
		panic(err)
	}

	// Get enough sentences to at least fill 64 characters.
	sentences := tokenizer.Tokenize(summary)
	text := ""
	for i := 0; len(text) < 64 && i < len(sentences); i++ {
		if len(text) > 0 {
			text += " "
		}
		text += sentences[i].Text
	}

	// Highlight the main word in the summary
	if s := strings.Index(strings.ToUpper(text), strings.ToUpper(title)); s < 0 {
		summary = "\x02" + title + "\x02: " + text
	} else {
		summary = text[0:s] + "\x02" + text[s:s+len(title)] + "\x02" + text[s+len(title):]
	}

	// Append ellipsis
	return summary + " \u2026"
}
