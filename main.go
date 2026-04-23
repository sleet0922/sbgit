package main

import (
	"fmt"
	"os"
)

func main() {
	EnableANSI()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法获取当前目录: %v\n", err)
		os.Exit(1)
	}

	git := NewGit(cwd)
	if !git.IsRepo() {
		fmt.Fprintln(os.Stderr, "错误: 当前目录不是 Git 仓库(Repository)")
		fmt.Fprintln(os.Stderr, "请在一个 Git 仓库中运行 sbgit")
		os.Exit(1)
	}

	t := NewTerminal()
	if err := t.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "无法初始化终端: %v\n", err)
		os.Exit(1)
	}
	defer t.Restore()

	app := NewApp(t, git)
	if err := app.Run(); err != nil {
		t.Restore()
		fmt.Fprintf(os.Stderr, "运行错误: %v\n", err)
		os.Exit(1)
	}
}
