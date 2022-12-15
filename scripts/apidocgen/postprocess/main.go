package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"golang.org/x/xerrors"
)

const (
	apiSubdir = "api"

	apiIndexFile    = "index.md"
	apiIndexContent = `Get started with Coderd API:

<children>
  This page is rendered on https://coder.com/docs/coder-oss/api. Refer to the other documents in the ` + "`" + `api/` + "`" + ` directory.
</children>
`
)

var (
	docsDirectory  string
	inMdFileSingle string

	sectionSeparator     = []byte("<!-- APIDOCGEN: BEGIN SECTION -->\n")
	nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9 ]+`)
)

func main() {
	log.Println("Postprocess API docs")

	flag.StringVar(&docsDirectory, "docs-directory", "../../docs", "Path to Coder docs directory")
	flag.StringVar(&inMdFileSingle, "in-md-file-single", "", "Path to single Markdown file, output from widdershins.js")
	flag.Parse()

	if inMdFileSingle == "" {
		flag.Usage()
		log.Fatal("missing value for in-md-file-single")
	}

	sections, err := loadMarkdownSections()
	if err != nil {
		log.Fatal("can't load markdown sections: ", err)
	}

	err = prepareDocsDirectory()
	if err != nil {
		log.Fatal("can't prepare docs directory: ", err)
	}

	err = writeDocs(sections)
	if err != nil {
		log.Fatal("can't write docs directory: ", err)
	}

	fmt.Println("Done")
}

func loadMarkdownSections() ([][]byte, error) {
	log.Printf("Read the md-file-single: %s", inMdFileSingle)
	mdFile, err := os.ReadFile(inMdFileSingle)
	if err != nil {
		return nil, xerrors.Errorf("can't read the md-file-single: %w", err)
	}
	log.Printf("Read %dB", len(mdFile))

	sections := bytes.Split(mdFile, sectionSeparator)
	if len(sections) < 2 {
		return nil, xerrors.Errorf("At least 1 section is expected: %w", err)
	}
	sections = sections[1:] // Skip the first element which is the empty byte array
	log.Printf("Loaded %d sections", len(sections))
	return sections, nil
}

func prepareDocsDirectory() error {
	log.Println("Prepare docs directory")

	apiPath := path.Join(docsDirectory, apiSubdir)

	err := os.RemoveAll(apiPath)
	if err != nil {
		return xerrors.Errorf(`os.RemoveAll failed for "%s": %w`, apiPath, err)
	}

	err = os.MkdirAll(apiPath, 0755)
	if err != nil {
		return xerrors.Errorf(`os.MkdirAll failed for "%s": %w`, apiPath, err)
	}
	return nil
}

func writeDocs(sections [][]byte) error {
	log.Println("Write docs to destination")

	apiDir := path.Join(docsDirectory, apiSubdir)
	err := os.WriteFile(path.Join(apiDir, apiIndexFile), []byte(apiIndexContent), 0644) // #nosec
	if err != nil {
		return xerrors.Errorf(`can't write the index file: %w`, err)
	}

	for _, section := range sections {
		sectionName, err := extractSectionName(section)
		if err != nil {
			return xerrors.Errorf("can't extract section name: %w", err)
		}
		log.Printf("Write section: %s", sectionName)

		docPath := path.Join(apiDir, sectionName)
		err = os.WriteFile(docPath, section, 0644) // #nosec
		if err != nil {
			return xerrors.Errorf(`can't write doc file "%s": %w`, docPath, err)
		}
	}
	return nil
}

func extractSectionName(section []byte) (string, error) {
	scanner := bufio.NewScanner(bytes.NewBuffer(section))
	if !scanner.Scan() {
		return "", xerrors.Errorf("section header was expected")
	}

	header := scanner.Text()[2:] // Skip #<space>
	return nonAlphanumericRegex.ReplaceAllLiteralString(strings.ToLower(strings.TrimSpace(header)), "-") + ".md", nil
}
