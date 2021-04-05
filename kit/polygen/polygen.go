package polygen

import (
	"os"

	"github.com/pkg/errors"
)

func getFileData(f *os.File) ([]byte, error) {
	// Getting the JPG Data size (ignoring the header)
	fileDataSize, err := f.Seek(0, 2)
	if err != nil {
		return nil, errors.Wrap(err, "failed to seek file to the end")
	}
	_, err = f.Seek(0, 0)
	if err != nil {
		return nil, errors.Wrap(err, "failed to seek file to after the header")
	}

	// Reading the JPG Data (ignoring the header)
	fileData := make([]byte, fileDataSize)
	n, err := f.Read(fileData)
	if err != nil || n != int(fileDataSize) {
		return nil, errors.Wrap(err, "failed to read the file data")
	}

	return fileData, nil
}

func discoverIV(magic, pw, input []byte) ([]byte, error) {
	var iv []byte
	// todo discover IV
	//secret.DecryptAes256()
	return iv, nil
}

func generatePdfMagic(pdf *os.File) ([]byte, []byte, error) {
	var pdfVersion []byte
	var i int
	pdfEndMagic := []byte("endstream\x0aendobj\x0a")

	pdfHeader := make([]byte, 32)
	n, err := pdf.Read(pdfHeader)
	if err != nil || n != 32 {
		return nil, nil, errors.Wrap(err, "failed to read the PDF header")
	}
	for i = 0; i < 32; i++ {
		// finding the PDF marker
		if pdfHeader[i] == 0x50 && pdfHeader[i+1] == 0x44 && pdfHeader[i+2] == 0x46 {
			pdfVersion = pdfHeader[i : i+7]
			break
		}
	}
	if i == 32 {
		return nil, nil, errors.New("failed to find out the PDF version")
	}

	// creating pdf begin magic based in PDF input file version
	pdfBeginMagic := []byte("%")
	pdfBeginMagic = append(pdfBeginMagic, pdfVersion...)
	pdfBeginMagic = append(pdfBeginMagic, "\x0a999 0 obj\x0a<<>>\x0astream\x0a"...)

	return pdfBeginMagic, pdfEndMagic, nil
}

func PdfJpgAes256(jpgFileName, pdfFileName string, bytePassword []byte) ([]byte, []byte, error) {
	var data []byte
	var iv []byte

	// Opening PDF file
	pdf, err := os.OpenFile(pdfFileName, os.O_RDONLY, 0755)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to open the file %s", pdfFileName)
	}
	defer pdf.Close()

	// Reading JPG file with as read only and
	jpg, err := os.OpenFile(jpgFileName, os.O_RDONLY, 0755)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to open the file %s", jpgFileName)
	}
	defer jpg.Close()

	//jpgData, err := getFileData(jpg)
	//if err != nil {
	//return nil, nil, errors.Wrap(err, "failed to get the JPG file data")
	//}

	// Getting the PDF magic
	//pdfBeginMagic, pdfEndMagic, err := generatePdfMagic(pdf)

	//// Finding the intermediate data after AES 256 ECB decryption
	//iv, err = discoverIV(pdfBeginMagic, bytePassword, jpgData[:32])
	//if err != nil {
	//return nil, nil, errors.Wrap(err, "failed to discover the IV")
	//}

	return data, iv, nil
}

func JpgPdf(jpgFileName, pdfFileName string) ([]byte, error) {
	const jpgHeaderSize = 0x14
	var data []byte

	pdf, err := os.OpenFile(pdfFileName, os.O_RDONLY, 0755)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open the file %s", pdfFileName)
	}
	defer pdf.Close()

	pdfBeginMagic, pdfEndMagic, err := generatePdfMagic(pdf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the pdf magic")
	}
	jpgComment := []byte("\xff\xfe\x00\x22\x0a")

	// Reading jpgfile with as read only and
	jpg, err := os.OpenFile(jpgFileName, os.O_RDONLY, 0755)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open the file %s", jpgFileName)
	}
	defer jpg.Close()

	// Reading the JPG header
	jpgHeader := make([]byte, jpgHeaderSize)
	n, err := jpg.Read(jpgHeader)
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
	data = append(data, jpgComment...)
	data = append(data, pdfBeginMagic...)
	data = append(data, jpgData...)
	data = append(data, pdfEndMagic...)
	data = append(data, pdfData...)

	return data, nil
}
