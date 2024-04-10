package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"net/url"
	"ntumods/pkg/dto"
	"os"
	"path/filepath"
	"reflect"
)

func IsEmpty(value interface{}) bool {
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			if !IsEmpty(v.Field(i).Interface()) {
				return false
			}
		}
		return true
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		return false
	}
}

func ExportStructToFile(filename string, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling data:", err)
		return
	}

	filePath := filepath.Join("..", "out", filename+".json")

	if _, err := os.Stat(filepath.Dir(filePath)); os.IsNotExist(err) {
		err := os.Mkdir(filepath.Dir(filePath), os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
	}

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}
}

func UploadFileToBlobStorage(blobName string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("[UploadFileToBlobStorage] Error marshaling data:", err)
		return err
	}

	azureStorageAccountAccessKey := os.Getenv("AZURE_STORAGE_ACCOUNT_ACCESS_KEY")

	credential, err := azblob.NewSharedKeyCredential(dto.ACCOUNT_NAME, azureStorageAccountAccessKey)
	if err != nil {
		return fmt.Errorf("[UploadFileToBlobStorage] Failed to create credential: %v", err)
	}

	// Create a service URL that points to your storage account.
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", dto.ACCOUNT_NAME)
	u, err := url.Parse(serviceURL)
	if err != nil {
		return fmt.Errorf("[UploadFileToBlobStorage] Failed to parse service URL: %v", err)
	}

	// Create a pipeline using the storage account's credentials.
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// Create a container URL using the pipeline, service URL, and container name.
	containerURL := azblob.NewContainerURL(*u, p).NewBlockBlobURL(dto.CONTAINER_NAME + "/" + blobName)

	// Upload the JSON data to the blob.
	_, err = azblob.UploadBufferToBlockBlob(context.Background(), jsonData, containerURL, azblob.UploadToBlockBlobOptions{})
	if err != nil {
		return fmt.Errorf("[UploadFileToBlobStorage] Failed to upload JSON data to blob: %v", err)
	}

	fmt.Printf("[UploadFileToBlobStorage] Successfully uploaded JSON data to blob '%s'\n", blobName)
	return nil
}
