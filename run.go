package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"gopkg.in/gographics/imagick.v3/imagick"
)

const (
	inputfile = "input.jpg"
)

func main() {
	if err := resizeExternally(inputfile, "result_ext.jpg"); err != nil {
		fmt.Println("External Error", err)
	}
	if err := resizeInternally(inputfile, "result_int.jpg"); err != nil {
		fmt.Println("Internal Error", err)
	}
}

func resizeExternally(inputfile, outputFile string) error {
	data, err := ioutil.ReadFile(inputfile)
	if err != nil {
		log.Fatal("ioutil.ReadFile error", err)
		return err
	}
	outFileStream, err := os.Create(outputFile)
	if err != nil {
		log.Fatal("os.Create error", err)
		return err
	}
	defer outFileStream.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/usr/bin/convert", "-", "-resize", "50%", "-")

	var wg sync.WaitGroup
	wg.Add(1)

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal("cmd.StdoutPipe", err)
		return err
	}
	go func() {
		defer wg.Done()
		_, err := io.Copy(outFileStream, stdoutPipe)
		if err != nil {
			log.Fatal("io.Copy", err)
		}
	}()

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal("cmd.StdinPipe", err)
		return err
	}
	go func() {
		defer stdinPipe.Close()
		_, err := stdinPipe.Write(data)
		if err != nil {
			log.Fatal("stdinPipe.Write", err)
		}
	}()

	if err := cmd.Start(); err != nil {
		log.Fatal("cmd.Start", err)
		return err
	}

	if ctx.Err() == context.DeadlineExceeded {
		log.Fatal("Command timed out")
		return ctx.Err()
	}
	if err != nil {
		log.Fatalf("Non-zero exit code: %s", err)
		return err
	}

	wg.Wait()

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
