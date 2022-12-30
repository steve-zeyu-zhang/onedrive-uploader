package sdk

import (
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestCreateDeleteDir(t *testing.T) {
	dirName := "/test-" + uuid.New().String()

	// Create
	err := IntegrationClient.CreateDir(dirName)
	checkTestBool(t, err == nil, true)

	// Delete
	err = IntegrationClient.Delete(dirName)
	checkTestBool(t, err == nil, true)
}

func TestDirInfo(t *testing.T) {
	dirName := "/test-" + uuid.New().String()

	// Create
	err := IntegrationClient.CreateDir(dirName)
	checkTestBool(t, err == nil, true)

	// Get info
	item, err := IntegrationClient.Info(dirName)
	checkTestBool(t, err == nil, true)
	checkTestBool(t, item.Type == DriveItemTypeFolder, true)
	checkTestString(t, strings.TrimPrefix(dirName, "/"), item.Name)
	checkTestInt(t, 0, item.Folder.ChildCount)

	// Delete
	err = IntegrationClient.Delete(dirName)
	checkTestBool(t, err == nil, true)
}

func TestUploadDownloadSmall(t *testing.T) {
	dirName := "/test-" + uuid.New().String()
	fileName := uuid.New().String() + ".txt"
	err := createRandomFile("/tmp/"+fileName, 1)
	checkTestBool(t, true, err == nil)
	defer os.Remove("/tmp/" + fileName)
	hash1, _ := getSHA1Hash("/tmp/" + fileName)
	hash256, _ := getSHA256Hash("/tmp/" + fileName)

	// Create
	err = IntegrationClient.CreateDir(dirName)
	checkTestBool(t, err == nil, true)

	// Upload
	err = IntegrationClient.Upload("/tmp/"+fileName, dirName)
	checkTestBool(t, err == nil, true)

	// Get info on folder
	item, err := IntegrationClient.Info(dirName)
	checkTestBool(t, err == nil, true)
	checkTestBool(t, item.Type == DriveItemTypeFolder, true)
	checkTestString(t, strings.TrimPrefix(dirName, "/"), item.Name)
	checkTestInt(t, 1, item.Folder.ChildCount)

	// Get info on file
	item, err = IntegrationClient.Info(dirName + "/" + fileName)
	checkTestBool(t, err == nil, true)
	checkTestBool(t, item.Type == DriveItemTypeFile, true)
	checkTestString(t, fileName, item.Name)
	checkTestString(t, "text/plain; charset=utf-8", item.File.MimeType)
	checkTestInt(t, 1024, int(item.SizeBytes))
	checkTestString(t, strings.ToUpper(hash1), item.File.Hashes.SHA1)
	checkTestString(t, strings.ToUpper(hash256), item.File.Hashes.SHA256)

	// Download
	os.Remove("/tmp/" + fileName)
	IntegrationClient.Download(dirName+"/"+fileName, "/tmp")
	hash256Downloaded, _ := getSHA256Hash("/tmp/" + fileName)
	checkTestString(t, hash256, hash256Downloaded)

	// Delete
	err = IntegrationClient.Delete(dirName)
	checkTestBool(t, err == nil, true)
}

func TestUploadDownloadLarge(t *testing.T) {
	dirName := "/test-" + uuid.New().String()
	fileName := uuid.New().String() + ".dat"
	fileSizeKB := (UploadSessionFileSizeLimit / 1024 * 2) + 10
	err := createRandomFile("/tmp/"+fileName, fileSizeKB)
	checkTestBool(t, true, err == nil)
	defer os.Remove("/tmp/" + fileName)
	hash1, _ := getSHA1Hash("/tmp/" + fileName)
	hash256, _ := getSHA256Hash("/tmp/" + fileName)

	// Create
	err = IntegrationClient.CreateDir(dirName)
	checkTestBool(t, err == nil, true)

	// Upload
	err = IntegrationClient.Upload("/tmp/"+fileName, dirName)
	checkTestBool(t, err == nil, true)

	// Get info on folder
	item, err := IntegrationClient.Info(dirName)
	checkTestBool(t, err == nil, true)
	checkTestBool(t, item.Type == DriveItemTypeFolder, true)
	checkTestString(t, strings.TrimPrefix(dirName, "/"), item.Name)
	checkTestInt(t, 1, item.Folder.ChildCount)

	// Get info on file
	item, err = IntegrationClient.Info(dirName + "/" + fileName)
	checkTestBool(t, err == nil, true)
	checkTestBool(t, item.Type == DriveItemTypeFile, true)
	checkTestString(t, fileName, item.Name)
	checkTestString(t, "application/octet-stream", item.File.MimeType)
	checkTestInt(t, fileSizeKB*1024, int(item.SizeBytes))
	checkTestString(t, strings.ToUpper(hash1), item.File.Hashes.SHA1)
	checkTestString(t, strings.ToUpper(hash256), item.File.Hashes.SHA256)

	// Download
	os.Remove("/tmp/" + fileName)
	IntegrationClient.Download(dirName+"/"+fileName, "/tmp")
	hash256Downloaded, _ := getSHA256Hash("/tmp/" + fileName)
	checkTestString(t, hash256, hash256Downloaded)

	// Delete
	err = IntegrationClient.Delete(dirName)
	checkTestBool(t, err == nil, true)
}

func TestList(t *testing.T) {
	// Create
	dirName := "/test-" + uuid.New().String()
	err := IntegrationClient.CreateDir(dirName)
	checkTestBool(t, err == nil, true)

	// Check empty dir
	items, err := IntegrationClient.List(dirName)
	checkTestBool(t, err == nil, true)
	checkTestInt(t, 0, len(items))

	// Prepare local files
	fileName1 := "a_" + uuid.New().String() + ".txt"
	fileName2 := "b_" + uuid.New().String() + ".png"
	createRandomFile("/tmp/"+fileName1, 1)
	createRandomFile("/tmp/"+fileName2, 2)
	defer os.Remove("/tmp/" + fileName1)
	defer os.Remove("/tmp/" + fileName2)

	// Create sub-folder and upload two files
	IntegrationClient.CreateDir(dirName + "/sub")
	IntegrationClient.Upload("/tmp/"+fileName1, dirName)
	IntegrationClient.Upload("/tmp/"+fileName2, dirName)

	// Check dir again
	items, err = IntegrationClient.List(dirName)
	checkTestBool(t, err == nil, true)
	checkTestInt(t, 3, len(items))
	// 1
	checkTestString(t, fileName1, items[0].Name)
	checkTestBool(t, true, items[0].Type == DriveItemTypeFile)
	checkTestInt(t, 1024, int(items[0].SizeBytes))
	// 2
	checkTestString(t, fileName2, items[1].Name)
	checkTestBool(t, true, items[1].Type == DriveItemTypeFile)
	checkTestInt(t, 2048, int(items[1].SizeBytes))
	// 3
	checkTestString(t, "sub", items[2].Name)
	checkTestBool(t, true, items[2].Type == DriveItemTypeFolder)

	// Delete one file
	IntegrationClient.Delete(dirName + "/" + fileName2)

	// Check dir again
	items, err = IntegrationClient.List(dirName)
	checkTestBool(t, err == nil, true)
	checkTestInt(t, 2, len(items))
	checkTestString(t, fileName1, items[0].Name)
	checkTestString(t, "sub", items[1].Name)

	// Delete
	err = IntegrationClient.Delete(dirName)
	checkTestBool(t, err == nil, true)
}

func createRandomFile(fileName string, sizeKB int) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	data := make([]byte, 1024)
	for i := 1; i <= sizeKB; i++ {
		rand.Read(data)
		if _, err := file.Write(data); err != nil {
			return err
		}
	}
	return nil
}

func getSHA1Hash(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func getSHA256Hash(fileName string) (string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
