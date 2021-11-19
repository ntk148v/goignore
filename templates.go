// Copyright 2021 Kien Nguyen-Tuan <kiennt2609@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/termie/go-shutil"
)

const (
	gitignoreUrl = "https://github.com/github/gitignore"
	dataDir      = ".cache/goignore"
)

var dataPath = path.Join(os.Getenv("HOME"), dataDir)

// commandExists checks if the given command exists
func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// unzip - get from https://stackoverflow.com/questions/20357223/easy-way-to-unzip-file-with-golang
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	return nil
}

// downloadTemplates gets the templates from Github repository
// 1. If git is installed, use git clone.
// 2. If not, just download then unzip it.
func downloadTemplates() error {
	// Check if git is already installed
	if commandExists("git") {
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
		return cmd.Run()
	}

	archivePath := path.Join(os.TempDir(), "gitignore.zip")
	// Create the file
	out, err := os.Create(archivePath)
	if err != nil {
		return err
	}

	resp, err := http.Get(fmt.Sprintf("%s/archive/master.zip", gitignoreUrl))
	if err != nil {
		return err
	}
	defer func() {
		resp.Body.Close()
		out.Close()
		os.Remove(archivePath)
	}()

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	// Unzip to a temporary directory
	if err = unzip(archivePath, os.TempDir()); err != nil {
		return err
	}

	err = shutil.CopyTree(path.Join(os.TempDir(), "gitignore-master"), dataPath, nil)
	if err != nil {
		return err
	}

	return nil
}

// initTemplates run at start up
func initTemplates() error {
	if _, err := os.Stat(dataPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		fmt.Println(statusMessageStyle("Initializing Gitignore template source..."))
		if err := downloadTemplates(); err != nil {
			// if error, clean the path
			os.Remove(dataPath)
			return err
		}
		fmt.Println(statusMessageStyle("Gitignore templates are downloaded"))
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
