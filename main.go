package main

import (
	"fmt"
	"os"
	"os/exec"
	"packer/config"
	"strings"
)

func tarGzGpg(args *config.PackerArgs) error {
	parts := []string{
		fmt.Sprintf("tar -cf - %s", strings.Join(args.GetSelectedFilesRawPaths(), " ")),
		fmt.Sprintf("pv -s %v", args.GetSelectedFileSize()),
		"gzip",
		fmt.Sprintf("gpg --passphrase %s --batch --quiet --yes -c > '%s'", args.Paraphrase, args.Output),
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

func main() {
	args, err := config.New()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = tarGzGpg(args)
	if err != nil {
		return
	}
}
