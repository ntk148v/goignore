package main

import (
	"io"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/charmbracelet/bubbles/list"
)

const (
	gitignoreUrl = "https://github.com/github/gitignore.git"
	dataDir      = ".cache/goignore"
)

var dataPath = path.Join(os.Getenv("HOME"), dataDir)

// initTemplates run at start up
func initTemplates() error {
	if _, err := os.Stat(dataPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		// if data dir is empty, get the newest Gitignore templates.
		// Setup directory
		curUsr, err := user.Current()
		if err != nil {
			return err
		}

		os.MkdirAll(dataPath, 0755)
		uid, _ := strconv.Atoi(curUsr.Uid)
		gid, _ := strconv.Atoi(curUsr.Gid)
		os.Chown(dataPath, uid, gid)

		cmd := exec.Command("git", []string{"clone", "--", gitignoreUrl, dataPath}...)
		cmd.Dir = dataPath
		if err := cmd.Run(); err != nil {
			// if error, clean the path
			os.Remove(dataPath)
			return err
		}
	}

	return nil
}

// updateTemplateList gets a list of templates from data dir
func updateTemplateList() ([]list.Item, error) {
	items := make([]list.Item, 0)
	ignoreRegEx, err := regexp.Compile("^.+(.gitignore)$")
	if err != nil {
		return items, err
	}
	err = filepath.Walk(dataPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && ignoreRegEx.MatchString(info.Name()) {
			if info.IsDir() {
				path = filepath.Join(path, info.Name())
			}
			items = append(items, item{
				title: info.Name(),
				path:  path,
			})
		}
		return nil
	})
	return items, err
}

// copyTemplate updates the .gitignore in current directory
// with the chosen gitignore
func copyTemplate(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}

	// create if file doesn't exist, append if file exists
	destFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0660)
	if err != nil {
		return err
	}

	defer func() {
		srcFile.Close()
		destFile.Close()
	}()

	if _, err = io.Copy(destFile, srcFile); err != nil {
		return err
	}

	err = destFile.Sync()
	return err
}

// pullTemplateUpdates performs git pull
func pullTemplateUpdates() error {
	cmd := exec.Command("git", []string{"pull"}...)
	cmd.Dir = dataPath
	err := cmd.Run()
	return err
}
