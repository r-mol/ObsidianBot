package repository

import (
	"fmt"
	"os"
	"path/filepath"
)

type fileSystem struct {
	AbsolutePath string
}

func New(absolutePath string) *fileSystem {
	return &fileSystem{
		AbsolutePath: absolutePath,
	}
}

func (fs *fileSystem) joinWithAbsolutePath(path string) string {
	return filepath.Join(fs.AbsolutePath, path)
}

func (fs *fileSystem) OpenFile(fp string) (*os.File, error) {
	fp = fs.joinWithAbsolutePath(fp)

	file, err := os.Open(fp)
	if err != nil {
		return nil, fmt.Errorf("open file [filepath = %q]: %w", fp, err)
	}

	return file, nil
}

func (fs *fileSystem) CreateFile(fp string) (*os.File, error) {
	fp = fs.joinWithAbsolutePath(fp)

	file, err := os.Create(fp)
	if err != nil {
		return nil, fmt.Errorf("create file [filepath = %q]: %w", fp, err)
	}

	return file, nil
}

func (fs *fileSystem) FileExist(fp string) (bool, error) {
	fp = fs.joinWithAbsolutePath(fp)

	info, err := os.Stat(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, fmt.Errorf("stat file [filepath = %q]: %w", fp, err)
	}

	return !info.IsDir(), nil
}

func (fs *fileSystem) ReadFromFile(fp string) (string, error) {
	fp = fs.joinWithAbsolutePath(fp)

	data, err := os.ReadFile(fp)
	if err != nil {
		return "", fmt.Errorf("read file [filepath = %q]: %w", fp, err)
	}

	return string(data), nil
}

func (fs *fileSystem) ReadDir(path string) ([]os.DirEntry, error) {
	path = fs.joinWithAbsolutePath(path)

	entities, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read dir [path = %q]: %w", path, err)
	}

	return entities, nil
}

func (fs *fileSystem) AppendToFile(fp string, data string) error {
	fp = fs.joinWithAbsolutePath(fp)

	file, err := os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file [filepath = %q]: %w", fp, err)
	}
	defer file.Close()

	if _, err := file.WriteString(data); err != nil {
		return fmt.Errorf("write string [filepath = %q]: %w", fp, err)
	}

	return nil
}

func (fs *fileSystem) WriteToFile(fp string, data string) error {
	fp = fs.joinWithAbsolutePath(fp)

	return os.WriteFile(fp, []byte(data), 0644)
}
