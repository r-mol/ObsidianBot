package usecases

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"
	_ "time/tzdata"
)

type Tag string

const (
	TagInbox        Tag = "inbox"
	TagShoppingList Tag = "shopping"
	TagAction       Tag = "action"
)

const (
	FilenameShoppingList = "Shopping List.md"
	FilenameWishList     = "Wish List.md"
	DirTimestamps        = "Timestamps"
)

type Repository interface {
	CreateFile(fp string) (*os.File, error)
	FileExist(fp string) (bool, error)
	ReadFromFile(fp string) (string, error)
	AppendToFile(fp string, data string) error
	WriteToFile(fp string, data string) error
}

type obsidian struct {
	Repo   Repository
	UserID int64
}

func NewObsidian(repo Repository, userID int64) *obsidian {
	return &obsidian{
		Repo:   repo,
		UserID: userID,
	}
}

func (us *obsidian) ParseMessage(ctx context.Context, msg string) (string, error) {
	tag, text, err := extractTagAndText(msg)
	if err != nil {
		if isSingleLine(msg) {
			return us.CreateNewNoteToInbox(ctx, msg)
		} else {
			return "", fmt.Errorf("extract tag: %w", err)
		}
	}

	var newMsg string
	switch Tag(tag) {
	case TagInbox:
		newMsg, err = us.CreateNewNoteToInbox(ctx, text)
	case TagShoppingList:
		newMsg, err = us.AddItemsToShoppingList(ctx, text)
	case TagAction:
		newMsg, err = us.AddAction(ctx, text)
	default:
		return "", fmt.Errorf("unknown tag [tag = %q]: %w", tag, err)
	}
	if err != nil {
		return "", fmt.Errorf("execute usecase for [tag = %q]: %w", tag, err)
	}

	return newMsg, nil
}

func (us *obsidian) CreateNewNoteToInbox(ctx context.Context, msg string) (string, error) {
	templateContent, err := us.Repo.ReadFromFile(FilePathInboxTemplate)
	if err != nil {
		return "", fmt.Errorf("read from file: %w", err)
	}

	templateContent = transformPlaceholders(templateContent)

	msg = strings.Title(msg)
	data := Inbox{
		Title: msg,
	}

	tmpl, err := template.New("inbox").Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("parse \"inbox\" template: %w", err)
	}

	outputFilePath := fmt.Sprintf("%s.md", msg)

	exist, err := us.Repo.FileExist(outputFilePath)
	if err != nil {
		return "", fmt.Errorf("check file exist: %w", err)
	}

	if exist {
		return "Note with such name already exist.", nil
	}

	outputFile, err := us.Repo.CreateFile(outputFilePath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)

	}
	defer outputFile.Close()

	err = tmpl.Execute(outputFile, data)
	if err != nil {
		return "", fmt.Errorf("execute template writing to file [filepath = %q]: %w", outputFilePath, err)
	}

	return fmt.Sprintf("Successfully create note %q with inbox tag.", msg), nil
}

func (us *obsidian) AddAction(ctx context.Context, msg string) (string, error) {
	moscowLocation, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return "", fmt.Errorf("load Moscow location: %w", err)
	}

	currentTime := time.Now().In(moscowLocation)

	content := fmt.Sprintf("\n%s - %s", currentTime.Format("15:04"), msg)

	filename := fmt.Sprintf("%s.md", currentTime.Format("2006-01-02"))
	fp := filepath.Join(DirTimestamps, filename)

	err = us.Repo.AppendToFile(fp, content)
	if err != nil {
		return "", fmt.Errorf("append to file: %w", err)
	}

	return fmt.Sprintf("Successfully add action to file. %s", fp), nil
}

func (us *obsidian) GetWishList(ctx context.Context, msg string) (string, error) {
	data, err := us.Repo.ReadFromFile(FilenameWishList)
	if err != nil {
		return "", fmt.Errorf("read from file: %w", err)
	}

	if data == "" {
		return "Wish list is empty!", nil
	}

	return data, nil
}

//func GetReadingList(ctx context.Context, msg string) (string, error) {
//
//}
//
//func GetWatchingList(ctx context.Context, msg string) (string, error) {
//
//}
// ---------------------------------------- Shopping ----------------------------------------

func (us *obsidian) GetShoppingList(ctx context.Context, msg string) (string, error) {
	data, err := us.Repo.ReadFromFile(FilenameShoppingList)
	if err != nil {
		return "", fmt.Errorf("read from file: %w", err)
	}

	items, err := extractItems(data)
	if err != nil {
		return "", fmt.Errorf("extract items to slice: %w", err)
	}

	if len(items) == 0 {
		return "Shopping list is empty!", nil
	}

	var content string
	for i, item := range items {
		item = strings.TrimPrefix(item, "-")
		item = strings.TrimSpace(item)

		content += fmt.Sprintf("%d. %s\n", i+1, item)
	}

	return content, nil
}

func (us *obsidian) AddItemsToShoppingList(ctx context.Context, msg string) (string, error) {
	data, err := us.Repo.ReadFromFile(FilenameShoppingList)
	if err != nil {
		return "", fmt.Errorf("read from file: %w", err)
	}

	oldItems, err := extractItems(data)
	if err != nil {
		return "", fmt.Errorf("extract old items to slice: %w", err)
	}

	newItems, err := extractItems(msg)
	if err != nil {
		return "", fmt.Errorf("extract new items to slice: %w", err)
	}

	items := append(oldItems, newItems...)

	updatedContent := strings.Join(items, "\n")

	err = us.Repo.WriteToFile(FilenameShoppingList, updatedContent)
	if err != nil {
		return "", fmt.Errorf("write to file: %w", err)
	}

	return "Successfully add items to shopping list. You can check it by /shopping_list", nil
}

func (us *obsidian) ClearShoppingList(ctx context.Context, msg string) (string, error) {
	err := us.Repo.WriteToFile(FilenameShoppingList, "")
	if err != nil {
		return "", fmt.Errorf("write to file: %w", err)
	}

	return "Successfully clear shopping list.", nil
}

func (us *obsidian) RemoveItemsFromShoppingList(ctx context.Context, msg string) (string, error) {
	msg = strings.TrimSpace(strings.TrimPrefix(msg, "/remove_item"))
	if msg == "" {
		return "", fmt.Errorf("should be provided minimum one id")
	}

	args := strings.Split(msg, ",")
	if len(args) < 1 {
		return "", fmt.Errorf("should be provided minimum one id")
	}

	var ids = make([]int, len(args))
	for i, arg := range args {
		arg = strings.TrimSpace(arg)

		id, err := strconv.Atoi(arg)
		if err != nil {
			return "", fmt.Errorf("convert string to int [id = %q]: %w", arg, err)
		}

		ids[i] = id - 1
	}

	data, err := us.Repo.ReadFromFile(FilenameShoppingList)
	if err != nil {
		return "", fmt.Errorf("read from file: %w", err)
	}

	items, err := extractItems(data)
	if err != nil {
		return "", fmt.Errorf("extract items to slice: %w", err)
	}

	for _, id := range ids {
		if id < 0 || id >= len(items) {
			return "", fmt.Errorf("invalid line index: %d", id+1)
		}
	}

	var removedItems = make([]string, len(ids))
	for i, id := range ids {
		removedItems[i] = items[id]
	}

	for i, id := range ids {
		id -= i
		items = append(items[:id], items[id+1:]...)
	}

	updatedContent := strings.Join(items, "\n")

	err = us.Repo.WriteToFile(FilenameShoppingList, updatedContent)
	if err != nil {
		return "", fmt.Errorf("write to file: %w", err)
	}

	return fmt.Sprintf("Successfully delete items from shopping list. Items:\n\n%s", strings.Join(removedItems, "\n")), nil
}

func (us *obsidian) RememberAboutInbox(ctx context.Context) string {
	return "👉🏼 **Please** _sort inbox_ 👈🏼"
}
