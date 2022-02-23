package main

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type PathInfo struct {
    Name    string
    Size    int64
    IsDir   bool
    Path   string
}

func calcSize(info os.FileInfo) int64 {
    // if directory calculate size of all files in it 
    if info.IsDir() {
        var size int64
        filepath.Walk(info.Name(), func(_ string, info os.FileInfo, err error) error {
            if err != nil {
                return err
            }
            if !info.IsDir() {
                size += info.Size()
            }
            return nil
        })
        return size
    }
    return info.Size()
}

func getPathInfoList(paths []string) ([]PathInfo, error) {
    var files []PathInfo
    for _, path := range paths {
        info, err := os.Stat(path)
        if err != nil {
            return nil, err
        }
        files = append(files, PathInfo{
            Name:    info.Name(),
            Size:    calcSize(info),
            IsDir:   info.IsDir(),

        })

    }
    return files, nil
}

func tarGzGpg(args *PackerArgs) error {
    paths := make([]string, len(args.files))
    for i, file := range args.files {
        paths[i] = file.Path
    }
    
	parts := []string{
		fmt.Sprintf("tar -cf - %s", strings.Join(paths, " ")),
		fmt.Sprintf("pv -s %v", args.FileSize()),
		"gzip",
		fmt.Sprintf("gpg --passphrase %s --batch --quiet --yes -c > '%s'", args.paraphrase, args.output),
	}
	cmd := exec.Command("sh", "-c", strings.Join(parts, " | "))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

type PackerArgs struct {
	files      []PathInfo
	paraphrase string
	output     string
}

func (args *PackerArgs) FileSize() int64 {
    var size int64
    for _, file := range args.files {
        size += file.Size
    }
    return size
}

func NewPackerArgs() (*PackerArgs, error) {
	var args PackerArgs
	// set paraphrase
	args.paraphrase = os.Getenv("PACKER_PASS")
	if args.paraphrase == "" {
		fmt.Println("PACKER_PASS is not set")
		err := survey.AskOne(&survey.Password{Message: "Paraphrase:"}, &args.paraphrase)
		if err != nil {
			return nil, err
		}
		if args.paraphrase == "" {
			return nil, fmt.Errorf("Paraphrase is empty")
		}
	} else {
		fmt.Println("Load paraphrase from env variable PACKER_PASS")
	}

	// set output file path
	if len(os.Args) >= 3 {
		args.output = os.Args[2]
	} else {
		args.output = os.Getenv("PACKER_OUT")
	}
	if args.output == "" {
		fmt.Println("No valid output file specified")
		err := survey.AskOne(&survey.Input{Message: "Output file:"}, &args.output)
		if err != nil {
			return nil, err
		}
		if args.output == "" {
			return nil, fmt.Errorf("output file is empty")
		}
		if !strings.HasSuffix(args.output, ".tar.gz.gpg") {
			args.output += ".tar.gz.gpg"
			fmt.Printf("output file is set to %s\n", args.output)
		}
	} else {
		if len(os.Args) >= 3 {
			fmt.Printf("Load output file from arguments\n")
		} else {
			fmt.Printf("Load output_file='%s' from env PACKER_OUT\n", args.output)
		}
	}

	path := os.Args[1]
    files, err := getPathInfoList([]string{path})
	if err != nil {
		return nil, err
	}


	var items []string
	for i, file := range files {
		// skip hidden files
		if file.Name[0] == '.' {
			continue
		}
		items = append(items, fmt.Sprintf("(%v) [%v MB] %s", i+1, file.Size/1024/1024, file.Name))
	}

	var selected []string
	err = survey.AskOne(&survey.MultiSelect{
		Message:  "Select files:",
		Options:  items,
		PageSize: 20,
	}, &selected)
	if err != nil {
		return nil, err
	}
	indexRegex := regexp.MustCompile(`^\((\d+)\)`)
	for _, item := range selected {
		position, _ := strconv.Atoi(indexRegex.FindStringSubmatch(item)[1]) 
        args.files = append(args.files, files[position-1])
	}
	if len(args.files) == 0 {
		return nil, fmt.Errorf("no files selected")
	}
	return &args, nil
}

func main() {
	// exit if os args are not enough
	if len(os.Args) < 2 {
		fmt.Println("Usage: packer <path-to-check> [output-file]")
		return
	}

	args, err := NewPackerArgs()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	err = tarGzGpg(args)
	if err != nil {
		fmt.Println(err)
		return
	}
}
