package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/spaceweasel/promptui"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type FileInfo struct {
	Name  string
	Size  int64
	IsDir bool
	Path  string
}

func (fi *FileInfo) SizeMB() string {
	return fmt.Sprintf("%.2f", float64(fi.Size)/(1024*1024))
}

func (fi *FileInfo) GetRawPath() string {
	pathRaw := fi.Path
	pathRaw = strings.Replace(pathRaw, "\"", "\\\"", -1)
	return fmt.Sprintf("\"%s\"", pathRaw)
}

type PackerArgs struct {
	Paraphrase    string
	Output        string
	Directory     string
	Files         []FileInfo
	SelectedFiles []FileInfo
}

func (p *PackerArgs) GetSelectedFileSize() int64 {
	var totalSize int64
	for _, file := range p.SelectedFiles {
		totalSize += file.Size
	}
	return totalSize
}

func (p *PackerArgs) GetSelectedFilesRawPaths() []string {
	var paths []string
	for _, file := range p.SelectedFiles {
		paths = append(paths, file.GetRawPath())
	}
	return paths
}

func (p *PackerArgs) LoadFromEnv() {
	p.Paraphrase = os.Getenv("PACKER_PASS")
	p.Output = os.Getenv("PACKER_OUT")
}

func (p *PackerArgs) LoadFromCmdArgs() {
	flag.StringVar(&p.Paraphrase, "pass", p.Paraphrase, "paraphrase")
	flag.StringVar(&p.Output, "out", p.Output, "output file name")
	flag.StringVar(&p.Directory, "dir", p.Directory, "directory to pack")
	flag.Parse()

	args := flag.Args()
	if len(args) > 0 {
		p.Directory = args[0]
	} else {
		p.Directory = "."
	}
}

func calculateDirectorySize(path string) int64 {
	var size int64
	err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if p == path {
				return nil
			}
			size += calculateDirectorySize(p)
		} else {
			size += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0
	}
	return size
}

func (p *PackerArgs) ListDownDirectory() error {
	files, err := ioutil.ReadDir(p.Directory)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() {
			p.Files = append(p.Files, FileInfo{
				Name:  file.Name(),
				Size:  calculateDirectorySize(filepath.Join(p.Directory, file.Name())),
				IsDir: true,
				Path:  filepath.Join(p.Directory, file.Name()),
			})
		} else {
			p.Files = append(p.Files, FileInfo{
				Name:  file.Name(),
				Size:  file.Size(),
				IsDir: false,
				Path:  filepath.Join(p.Directory, file.Name()),
			})
		}
	}
	return nil
}

func (p *PackerArgs) PromptForSelectFiles() error {
	// use prompt-ui to select files
	items := make([]string, len(p.Files))
	for i, file := range p.Files {
		items[i] = fmt.Sprintf("%s (%s)", file.Name, file.SizeMB())
	}
	prompt := promptui.MultiSelect{
		Label: "Select files",
		Items: items,
	}
	indexes, err := prompt.Run()
	if err != nil {
		return err
	}

	for _, index := range indexes {
		p.SelectedFiles = append(p.SelectedFiles, p.Files[index])
	}
	return nil
}

func (p *PackerArgs) Validate() error {
	if p.Paraphrase == "" {
		return errors.New("paraphrase is empty")
	}
	if p.Output == "" {
		return errors.New("paraphrase is empty")
	}
	if p.Directory == "" {
		return errors.New("paraphrase is empty")
	}
	if len(p.Files) == 0 {
		return errors.New("files is empty")
	}
	return nil
}

func New() (*PackerArgs, error) {
	var err error
	var args PackerArgs
	args.LoadFromEnv()
	args.LoadFromCmdArgs()
	err = args.ListDownDirectory()
	if err != nil {
		return nil, err
	}
	err = args.PromptForSelectFiles()
	if err != nil {
		return nil, err
	}
	err = args.Validate()
	if err != nil {
		return nil, err
	}
	return &args, nil
}
