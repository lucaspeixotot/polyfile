package polygen

import (
	"os"

	"github.com/pkg/errors"
)

func JpgPdf(jpgFileName, pdfFileName string) ([]byte, error) {
	const jpgHeaderSize = 0x14
	var data []byte
	var i int
	var pdfVersion []byte
	pdfEndMagic := []byte("endstream\x0aendobj\x0a")

	// Reading pdffile with as read only and
	pdf, err := os.OpenFile(pdfFileName, os.O_RDONLY, 0755)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open the file %s", pdfFileName)
	}
	defer pdf.Close()

	pdfHeader := make([]byte, 32)
	n, err := pdf.Read(pdfHeader)
	if err != nil || n != 32 {
		return nil, errors.Wrap(err, "failed to read the PDF header")
	}
	for i = 0; i < 32; i++ {
		// finding the PDF marker
		if pdfHeader[i] == 0x50 && pdfHeader[i+1] == 0x44 && pdfHeader[i+2] == 0x46 {
			pdfVersion = pdfHeader[i : i+7]
			break
		}
	}
	if i == 32 {
		return nil, errors.New("failed to find out the PDF version")
	}

	// creating pdf begin magic based in PDF input file version
	pdfBeginMagic := []byte("\xff\xfe\x00\x22\x0a%")
	pdfBeginMagic = append(pdfBeginMagic, pdfVersion...)
	pdfBeginMagic = append(pdfBeginMagic, "\x0a999 0 obj\x0a<<>>\x0astream\x0a"...)

	// Reading jpgfile with as read only and
	jpg, err := os.OpenFile(jpgFileName, os.O_RDONLY, 0755)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open the file %s", jpgFileName)
	}
	defer jpg.Close()

	// Reading the JPG header
	jpgHeader := make([]byte, jpgHeaderSize)
	n, err = jpg.Read(jpgHeader)
	if err != nil || n != jpgHeaderSize {
		return nil, errors.Wrap(err, "failed to read the JPG header")
	}

	// Getting the JPG Data size (ignoring the header)
	jpgDataSize, err := jpg.Seek(0, 2)
	if err != nil {
		return nil, errors.Wrap(err, "failed to seek JPG file to the end")
	}
	jpgDataSize = jpgDataSize - jpgHeaderSize
	_, err = jpg.Seek(jpgHeaderSize, 0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to seek JPG file to after the header")
	}

	// Reading the JPG Data (ignoring the header)
	jpgData := make([]byte, jpgDataSize)
	n, err = jpg.Read(jpgData)
	if err != nil || n != int(jpgDataSize) {
		return nil, errors.Wrap(err, "failed to read the JPG data")
	}

	// Finding the \xff\xdb marker
	for i := 0; i < int(jpgDataSize)-1; i++ {
		if jpgData[i] == 0xff && jpgData[i+1] == 0xdb {
			jpgData = jpgData[i:]
			break
		}
	}

	// Getting the PDF data size
	pdfDataSize, err := pdf.Seek(0, 2)
	if err != nil {
		return nil, errors.Wrap(err, "failed to seek PDF file to the end")
	}
	_, err = pdf.Seek(0, 0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to seek PDF file to the beginning")
	}

	// Reading the PDF data
	pdfData := make([]byte, pdfDataSize)
	n, err = pdf.Read(pdfData)
	if err != nil || n != int(pdfDataSize) {
		return nil, errors.Wrap(err, "failed to read the PDF data")
	}

	// Appending to the result byte file
	data = append(data, jpgHeader...)
	data = append(data, pdfBeginMagic...)
	data = append(data, jpgData...)
	data = append(data, pdfEndMagic...)
	data = append(data, pdfData...)

	return data, nil
}
