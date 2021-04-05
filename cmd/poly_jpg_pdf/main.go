package main

import (
	"fmt"
	"os"

	"github.com/lucaspeixotot/polyfile/kit/polygen"
	"github.com/pkg/errors"
)

func createOutputFile(data []byte, outputFileName string) error {
	output, err := os.OpenFile(outputFileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to create the output with the filename %s", outputFileName)
	}

	n, err := output.Write(data)
	if err != nil || n != len(data) {
		return errors.Wrap(err, "failed to write the data to the output file")
	}

	defer output.Close()
	return nil
}

func main() {
	args := os.Args[1:]
	if len(args) != 3 {
		fmt.Printf("The program needs three parameters in the exact order: jpg file name, pdf file name, output file name.\n")
		os.Exit(1)
	}

	jpgFileName := args[0]
	pdfFileName := args[1]
	outputFileName := args[2]

	outputBytes, err := polygen.JpgPdf(jpgFileName, pdfFileName)
	if err != nil {
		fmt.Printf("Failed to generate the jpg/pdf file: %+v", err)
		os.Exit(1)
	}

	err = createOutputFile(outputBytes, outputFileName)
	if err != nil {
		fmt.Printf("Failed to create the jpg/pdf file: %+v", err)
	}

	os.Exit(0)
}
