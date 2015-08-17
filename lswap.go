package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type content []string

var contents content
var from, to string

// implementing flag.Value.String()
func (f *content) String() string {
	return fmt.Sprint(*f)
}

// implementing flag.Value.Set()
func (f *content) Set(value string) error {
	for _, c := range strings.Split(value, ",") {
		if len(c) == 0 {
			return errors.New("cannot provide empty string for content")
		}
		*f = append(*f, c)
	}
	return nil
}

func init() {
	flag.StringVar(&from, "from", "", "folder where code currently is")
	flag.StringVar(&to, "to", "", "folder where code will be moved")
	flag.Var(&contents, "contents", "comma separated list of contents under `from` to move")
}

func main() {
	flag.Parse()
	runIt()
}

func runIt() {

	// from, to, and contents are specified
	err := checkFlagsPresent(from, to, contents)
	if err != nil {
		log.Fatal(err)
	}

	// from resolves to a valid path
	fromPath, err := resolveAbsPath(from)
	if err != nil {
		log.Fatal(err)
	}
	// to resolves to a valid path
	toPath, err := resolveAbsPath(to)
	if err != nil {
		log.Fatal(err)
	}

	// get rid of additional redirections in path
	fromPath = filepath.Clean(fromPath)
	toPath = filepath.Clean(toPath)

	// ensure from and to point to diffrent folders
	if fromPath == toPath {
		reportSame(fromPath, toPath)
	}

	// neither of from and to are symlinks
	fromFI := checkSymlinkGetFI(fromPath)
	toFI := checkSymlinkGetFI(toPath)

	// double check equality with fileinfo comparison
	if os.SameFile(fromFI, toFI) {
		reportSame(fromPath, toPath)
	}

	// from and to must be folders
	if !fromFI.IsDir() || !toFI.IsDir() {
		log.Fatal("from and to both need to be directories")
	}

	// check all contents found under from, if not exit
	// if any of contents is found as symlink under from, exit
	err = hasContentsAsNonSymLinks(fromPath, contents)
	if err != nil {
		log.Fatal(err)
	}

	// if any of contents is found under `to`, exit
	// Note: any content found as symlink under `to` is ok as long as -
	// for contents under `to` that are symlinks, they can only be symlinks to
	// real content under `from`
	err = contentsAbsentOrOnlyAsSymLinksFrom(toPath, contents, fromPath)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("source and destination looks good, starting work ..")

	// remove contents symlinks from `to`
	err = removeSymLinks(toPath, contents)
	if err != nil {
		log.Fatal(err)
	}

	// move contents `from` -> `to`
	err = move(fromPath, toPath, contents)
	if err != nil {
		log.Fatal(err)
	}

	// create symlinks for to/contents under from/
	err = link(toPath, fromPath, contents)
	if err != nil {
		log.Fatal(err)
	}

	// done
	log.Println("done")
}

func removeSymLinks(toPath string, contents content) error {
	for _, c := range contents {

		_, err := os.Lstat(filepath.Join(toPath, c))
		if err != nil && os.IsNotExist(err) {
			continue
		}

		if err != nil && !os.IsNotExist(err) {
			return err
		}

		err = os.Remove(filepath.Join(toPath, c))
		if err != nil {
			return err
		}
	}

	return nil
}

func move(fromPath, toPath string, contents content) error {
	for _, c := range contents {

		src := filepath.Join(fromPath, c)
		dest := filepath.Join(toPath, c)

		err := os.Rename(src, dest)
		if err != nil {
			return err
		}
	}

	return nil
}

func link(source, dest string, contents content) error {
	for _, c := range contents {

		original := filepath.Join(source, c)
		link := filepath.Join(dest, c)

		err := os.Symlink(original, link)
		if err != nil {
			return err
		}
	}

	return nil
}

func contentsAbsentOrOnlyAsSymLinksFrom(
	to string,
	contents content,
	from string,
) error {

	for _, c := range contents {
		err := contentAbsentOrIsSymLink(to, c, from)
		if err != nil {
			return err
		}
	}

	return nil
}

func contentAbsentOrIsSymLink(
	to string,
	content string,
	from string,
) error {

	toPath := filepath.Join(to, content)

	// file not found in `to`
	fiTo, err := os.Lstat(toPath)
	if err != nil && os.IsNotExist(err) {
		return nil
	}

	// some other error
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// no error, file found
	if fiTo.Mode()&os.ModeSymlink != os.ModeSymlink {
		return fmt.Errorf("%s is not a symlink", toPath)
	}

	sourcePath, err := os.Readlink(toPath)
	if err != nil {
		return fmt.Errorf(
			"cannot get source path for `to` (target) %s; details: %s",
			toPath,
			err,
		)
	}

	fiToSource, err := os.Lstat(sourcePath)
	if err != nil {
		return fmt.Errorf(
			"cannot lstat source folder of `to` symlink %s; details: %s",
			toPath,
			err,
		)
	}

	fromPath := filepath.Join(from, content)
	fiFrom, err := os.Lstat(fromPath)
	if err != nil {
		return fmt.Errorf(
			"cannot get source of target symlink %s in expected path %s, details: %s",
			toPath,
			fromPath,
			err,
		)
	}

	if !os.SameFile(fiToSource, fiFrom) {
		return fmt.Errorf(
			"target symlink %s does not point to expected source %s",
			toPath,
			fromPath,
		)
	}

	return nil
}

func hasContentsAsNonSymLinks(dir string, contents content) error {
	for _, c := range contents {
		err := contentExistsAsNonSymLink(dir, c)
		if err != nil {
			return err
		}
	}

	return nil
}

func contentExistsAsNonSymLink(folder, content string) error {
	fi, err := os.Lstat(filepath.Join(folder, content))

	if err != nil {
		return err
	}

	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		return fmt.Errorf("%s is not a symlink under %s", content, folder)
	}

	return nil
}

func checkSymlinkGetFI(path string) os.FileInfo {
	fi, err := os.Lstat(path)
	if err != nil {
		log.Fatalf("Cannot read file information for: %s", path)
	}
	if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
		log.Fatalf("Provided path should not be a symlink: %s", path)
	}

	return fi
}

func reportSame(fromPath, toPath string) {
	log.Fatalf("from: %s and to: %s are same file locations", fromPath, toPath)
}

func checkFlagsPresent(from, to string, contents content) error {
	if from == "" || to == "" || contents == nil {
		return errors.New("Need to provide from, to and contents")
	}
	return nil
}

// resolveAbsPath return a string that is an absolute path for the given
// relative path `relPath`. A blank string with an error is returned if the
// resolution fails.
func resolveAbsPath(relPath string) (string, error) {
	absConfigPath, err := filepath.Abs(filepath.FromSlash(relPath))
	if err != nil {
		return "", fmt.Errorf("cannot resolve absolute path for %s, error: %s",
			relPath, err)
	}

	return absConfigPath, nil
}
