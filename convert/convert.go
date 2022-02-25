package convert

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Config struct {
    paraphrase string
    output     string 
    files   []string
    fileSize int64
}

func TarGzGpg(args *Config) error {
    paths := make([]string, len(args.files))
	parts := []string{
		fmt.Sprintf("tar -cf - %s", strings.Join(paths, " ")),
		fmt.Sprintf("pv -s %v", args.fileSize),
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
