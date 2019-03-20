package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"sync"
	"time"

	a "github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/pkg/errors"
	"gopkg.in/gographics/imagick.v3/imagick"
)

const (
	// megaByte         = 1 << 20
	// defaultBlockSize = 50 * megaByte
	maxRetries                                = 5
	retryDelay                                = 1 * time.Second
	timeout                                   = 10 * time.Second
	environmentVariableNameStorageAccountName = "az_storage_name"
	environmentVariableNameStorageAccountKey  = "az_storage_key"
	executable                                = "/usr/bin/convert"
	containerName                             = "ocirocks3"
	blobName                                  = "20181007-110205-L1016848.jpg"
)

func main() {
	var (
		storageAccountName = os.Getenv(environmentVariableNameStorageAccountName)
		storageAccountKey  = os.Getenv(environmentVariableNameStorageAccountKey)
	)

	sharedKeyCredential, _ := a.NewSharedKeyCredential(storageAccountName, storageAccountKey)
	pipeline := a.NewPipeline(sharedKeyCredential, a.PipelineOptions{
		Retry: a.RetryOptions{
			Policy:     a.RetryPolicyExponential,
			MaxTries:   maxRetries,
			RetryDelay: retryDelay,
		}})

	url, _ := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", storageAccountName))
	serviceURL := a.NewServiceURL(*url, pipeline)
	containerURL := serviceURL.NewContainerURL(containerName)
	blobURL := containerURL.NewBlockBlobURL(blobName)
	fmt.Printf("%s\n", blobURL)

	// blobURL.Download()

	if err := resizeExternally("input.jpg", "result_ext.jpg"); err != nil {
		fmt.Println("External Error", err)
	}

	if err := resizeInternally("input.jpg", "result_int.jpg"); err != nil {
		fmt.Println("Internal Error", err)
	}
}

func resizeExternally(inputFileName, outputFileName string) error {
	inputReader, err := os.Open(inputFileName)
	if err != nil {
		log.Fatalf("ioutil.ReadFile: %s\n", err)
		return err
	}
	defer inputReader.Close()

	outWriter, err := os.Create(outputFileName)
	if err != nil {
		log.Fatalf("os.Create: %s\n", err)
		return err
	}
	defer outWriter.Close()

	return resize(inputReader, outWriter)
}

func resize(inputReader io.Reader, outputWriter io.Writer) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	const (
		inputFile  = "-"
		outputFile = "-"
	)
	args := []string{inputFile, "-resize", "50%", outputFile}

	cmd := exec.CommandContext(ctx, executable, args...)
	if err := execCommandPumpData(cmd, inputReader, outputWriter); err != nil {
		return err
	}

	if ctx.Err() == context.DeadlineExceeded {
		return ctx.Err()
	}

	return nil
}

func execCommandPumpData(cmd *exec.Cmd, inputReader io.Reader, outputWriter io.Writer) error {
	var (
		wg          sync.WaitGroup
		errorBuffer bytes.Buffer
		errorWriter = bufio.NewWriter(&errorBuffer)
	)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("cmd.StdoutPipe(): %s\n", err)
		return err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := io.Copy(outputWriter, stdoutPipe)
		if err != nil {
			log.Fatalf("io.Copy(outputWriter, stdoutPipe): %s\n", err)
		}
	}()

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalf("cmd.StderrPipe(): %s\n", err)
		return err
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := io.Copy(errorWriter, stderrPipe)
		if err != nil {
			log.Fatalf("io.Copy(errorWriter, stderrPipe): %s\n", err)
		}
	}()

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("cmd.StdinPipe(): %s\n", err)
		return err
	}
	go func() {
		defer stdinPipe.Close()
		if _, err := io.Copy(stdinPipe, inputReader); err != nil {
			log.Fatalf("io.Copy(stdinPipe, inputReader): %s\n", err)
		}
	}()

	if err := cmd.Start(); err != nil {
		return errors.Wrapf(err, "cmd.Start()")
	}

	wg.Wait()
	if err := cmd.Wait(); err != nil {
		return errors.Wrapf(err, "cmd.Wait(): %s", string(errorBuffer.Bytes()))
	}

	return nil
}

func resizeInternally(inputfile, outputFile string) error {
	imagick.Initialize()
	defer imagick.Terminate()
	mw := imagick.NewMagickWand()
	err := mw.ReadImage(inputfile)
	if err != nil {
		return err
	}

	width := mw.GetImageWidth()
	height := mw.GetImageHeight()
	hWidth := uint(width / 2)
	hHeight := uint(height / 2)

	err = mw.ResizeImage(hWidth, hHeight, imagick.FILTER_LANCZOS)
	if err != nil {
		return err
	}

	err = mw.SetImageCompressionQuality(95)
	if err != nil {
		return err
	}

	err = mw.WriteImage(outputFile)
	if err != nil {
		return err
	}

	return nil
}
