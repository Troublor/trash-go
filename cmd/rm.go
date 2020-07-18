package cmd

import (
	"fmt"
	"github.com/Troublor/go-trash/errs"
	"github.com/Troublor/go-trash/storage"
	"github.com/Troublor/go-trash/system"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path"
	"path/filepath"
)

var recursive bool
var directory bool
var permanent bool

var rmCmd = &cobra.Command{
	Use:   "rm [-d]|[-r][-p]",
	Short: "Remove the files or directories by putting them in trash bin",
	Long:  `Remove the files or directories by putting them in trash bin`,
	Run: func(cmd *cobra.Command, args []string) {
		for _, itemPath := range args {
			id, err := Remove(itemPath, directory, recursive, permanent)
			if err != nil {
				fmt.Println("Remove Error: " + err.Error())
			} else {
				fmt.Println("remove " + itemPath + " complete, trash id = " + id)
			}
		}
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&recursive, "recursive", "r", false,
		"recursively remove files in directory")
	rmCmd.Flags().BoolVarP(&directory, "directory", "d", false,
		"Remove directory")
	rmCmd.Flags().BoolVarP(&permanent, "permanent", "p", false,
		"permanently remove (without putting into trash bin)")
}

func Remove(itemPath string, isDirectory bool, recursive bool, permanent bool) (string, error) {
	trashDir := storage.GetTrashBinPath()
	var err error
	if !filepath.IsAbs(itemPath) {
		itemPath, err = filepath.Abs(itemPath)
		if err != nil {
			panic(err)
		}
	}
	fileInfo, err := os.Stat(itemPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", errs.ItemNotExistError
		} else {
			panic(err)
		}
	}

	if !isDirectory {
		if fileInfo.IsDir() {
			return "", errs.IsDirectoryError
		}
		if permanent {
			err = os.Remove(itemPath)
			return "", err
		} else {
			// add information in database
			id := storage.DbInsertTrashItem(itemPath, trashDir, fileInfo.Name(), storage.TYPE_FILE, system.GetUser())
			// move the item into trash directory
			err := system.SafeRename(itemPath, path.Join(trashDir, id))
			if err != nil {
				panic(err)
			}
			return id, nil
		}

	} else {
		if !fileInfo.IsDir() {
			return "", errs.IsFileError
		}
		isEmpty, err := DirectoryIsEmpty(itemPath)
		if err != nil {
			panic(err)
		}
		if !isEmpty && !recursive {
			return "", errs.DirectoryNotEmptyError
		}
		if permanent {
			err = os.RemoveAll(itemPath)
			return "", err
		} else {
			// add information in database
			id := storage.DbInsertTrashItem(itemPath, trashDir, fileInfo.Name(), storage.TYPE_DIRECTORY, system.GetUser())
			// move the item into trash directory
			err = system.SafeRename(itemPath, path.Join(trashDir, id))
			if err != nil {
				panic(err)
			}
			return id, nil
		}
	}
}

func DirectoryIsEmpty(dirPath string) (bool, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		return false, err
	}
	defer func() {
		_ = f.Close()
	}()
	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}
