package proto

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var CreateCmd = &cobra.Command{
	Use:   "create",
	Short: "create the proto code",
	Long:  "create the proto code",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			fmt.Println("Enter proto files or directory")
			return
		}

		plugins := []string{"protoc-gen-go", "protoc-gen-go-grpc", "protoc-gen-gin", "protoc-gen-openapiv2", "protoc-gen-validate"}
		err := find(plugins...)
		if err != nil {
			cmd := exec.Command("hack", "install")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err = cmd.Run(); err != nil {
				return
			}
		}

		err = walk(strings.TrimSpace(args[0]))
		if err != nil {
			fmt.Println(err)
		}
	},
}

func find(name ...string) error {
	var err error
	for _, n := range name {
		_, err = exec.LookPath(n)
		if err != nil {
			break
		}
	}
	return err
}

func walk(dir string) error {
	if len(dir) == 0 {
		return errors.New("dir invalid")
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) != ".proto" || strings.HasPrefix(path, "third_party") {
			return nil
		}
		return create(path, dir)
	})
}

func create(path, dir string) error {
	fmt.Println("dir:", dir)
	cmd := []string{
		"-I=.",
		"-I=" + "third_party",
		"--go_out=" + dir,
		"--go_opt=paths=source_relative",
		"--go-grpc_out=" + dir,
		"--go-grpc_opt=paths=source_relative",
		"--gin_out=" + dir,
		"--gin_opt=paths=source_relative",
		"--openapiv2_out=" + dir,
		"--openapiv2_opt=logtostderr=true",
		"--openapiv2_opt=json_names_for_fields=false",
	}
	protoBytes, err := os.ReadFile(path)
	if err == nil && len(protoBytes) > 0 {
		ok, _ := regexp.Match(`\n[^/]*(import)\s+"validate/validate.proto"`, protoBytes)
		if ok {
			cmd = append(cmd, "--validate_out="+dir, "--validate_opt=paths=source_relative,lang=go")
		}
	}
	cmd = append(cmd, path)
	fd := exec.Command("protoc", cmd...)
	fmt.Println("cmd:", fd.String())
	fd.Stdout = os.Stdout
	fd.Stderr = os.Stderr
	return fd.Run()
}
