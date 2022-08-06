package environment

import (
	"github.com/gojinja/gojinja/src/encoding"
	"github.com/gojinja/gojinja/src/errors"
	"github.com/gojinja/gojinja/src/utils"
	"github.com/gojinja/gojinja/src/utils/mapUtils"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func splitTemplatePath(template string) (pieces []string, err error) {
	for _, piece := range strings.Split(template, "/") {
		if strings.Contains(piece, string(os.PathSeparator)) || piece == ".." {
			return nil, errors.TemplateNotFound(template, "")
		} else if piece != "." {
			pieces = append(pieces, piece)
		}
	}
	return
}

type Loader struct {
	LoaderEmbed
}

type LoaderEmbed interface {
	HasSourceAccess() bool
	GetSource(env *Environment, template string) (string, *string, UpToDate, error)
	ListTemplates() ([]string, error)
}

func (l *Loader) Load(env *Environment, name string, globals map[string]any) (ITemplate, error) {
	source, filename, upToDate, err := l.GetSource(env, name)
	if err != nil {
		return nil, err
	}
	return FromString(env, source, filename, globals, upToDate)
}

type fsLoader struct {
	searchPath  []string
	encoding    string
	followLinks bool
}

func (f fsLoader) HasSourceAccess() bool {
	return true
}

func (f fsLoader) GetSource(_ *Environment, template string) (string, *string, UpToDate, error) {
	pieces, err := splitTemplatePath(template)
	if err != nil {
		return "", nil, nil, err
	}

	var filename string
	var info os.FileInfo
	for _, searchPath := range f.searchPath {
		for _, piece := range pieces {
			searchPath = path.Join(searchPath, piece)
		}
		info, err = os.Stat(searchPath)
		if err != nil {
			return "", nil, nil, err
		}
		if info.Mode().IsRegular() {
			filename = path.Clean(searchPath)
			break
		}
	}

	if filename == "" {
		return "", nil, nil, errors.TemplateNotFound(template, "")
	}

	mtime := info.ModTime()

	contents, err := os.ReadFile(filename)
	if err != nil {
		return "", nil, nil, err
	}
	decoded, err := encoding.Decode(contents, f.encoding)
	if err != nil {
		return "", nil, nil, err
	}

	upToDate := func() bool {
		info, err = os.Stat(filename)
		if err != nil {
			return false
		}
		return info.ModTime() == mtime
	}

	return decoded, &filename, upToDate, nil
}

func (f fsLoader) ListTemplates() ([]string, error) {
	found := make(map[string]struct{})

	for _, searchPath := range f.searchPath {
		walkFn := func(p string, d fs.DirEntry, err error) error {
			template := p[len(searchPath):]
			template = strings.Trim(template, string(os.PathSeparator))
			template = strings.Replace(template, string(os.PathSeparator), "/", -1)
			template = strings.TrimPrefix(template, "./")
			found[template] = struct{}{}
			return nil
		}

		err := filepath.WalkDir(searchPath, walkFn)
		if err != nil {
			return nil, err
		}
	}
	return mapUtils.SortedKeys(found), nil
}

func NewFileSystemLoader[S utils.StrOrSlice](searchPath S, encoding string, followLinks bool) *Loader {
	return &Loader{
		fsLoader{
			searchPath:  utils.ToStrSlice(searchPath),
			encoding:    encoding,
			followLinks: followLinks,
		},
	}
}
