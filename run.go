package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
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
		log.Fatal("File reading error", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "/usr/bin/convert", inputfile, "-resize", "50%", outputFile)

	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		log.Fatal("Command timed out")
		return ctx.Err()
	}
	if err != nil {
		log.Fatalf("Non-zero exit code: %s", err)
		return err
	}
	fmt.Println(string(out))

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
