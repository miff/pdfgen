package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/phpdave11/gofpdf"
	"github.com/skip2/go-qrcode"
)

type Person struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

func main() {
	const numWorkers = 10
	var wg sync.WaitGroup
	jobs := make(chan string, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for name := range jobs {
				generateFile(name)
			}
		}()
	}

	// for i := 0; i < 1000; i++ {
	name := fmt.Sprintf("zaposleni-%d", 0)
	jobs <- name
	// }
	close(jobs)
	wg.Wait()
}

func generateFile(name string) {
	jsonFile, err := os.Open("generated.json")
	if err != nil {
		fmt.Printf("Error opening JSON file for %s: %v\n", name, err)
		return
	}
	defer jsonFile.Close()

	jsonData, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("Error reading JSON file for %s: %v\n", name, err)
		return
	}

	var people []Person
	err = json.Unmarshal(jsonData, &people)
	if err != nil {
		fmt.Printf("Error unmarshaling JSON for %s: %v\n", name, err)
		return
	}

	pdf := gofpdf.New("P", "mm", "A4", "")

	pdf.SetHeaderFunc(func() {
		pdf.SetFont("Arial", "B", 12)
		pdf.Image("assets/logo.png", 10, 0, 30, 0, false, "", 0, "")
		pdf.MoveTo(50, 10)
		pdf.Cell(10, 0, "Zaposleni - Header")
		pdf.Ln(20)
	})

	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Arial", "I", 8)
		pdf.Cell(0, 10, fmt.Sprintf("Strana %d", pdf.PageNo()))
	})

	qrCode, err := generateQRCode("https://miff.me")
	if err != nil {
		fmt.Println("Can't generate QR code!")
	}

	lineHeight := 10.0

	addPage := func() {
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(40, 10, "Zaposleni")
		pdf.ImageOptions(qrCode, 180, 0, 30, 30, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
		pdf.Ln(12)
		pdf.SetFont("Arial", "", 8)
		pdf.Cell(30, lineHeight, "Ime")
		pdf.Cell(10, lineHeight, "Godine")
		pdf.Cell(40, lineHeight, "Email")
		pdf.Cell(30, lineHeight, "Phone")
		pdf.Cell(80, lineHeight, "Address")
		pdf.Ln(10)
	}

	addPage()

	for _, person := range people {
		pdf.Cell(30, lineHeight, person.Name)
		pdf.Cell(10, lineHeight, fmt.Sprintf("%d", person.Age))
		pdf.Cell(40, lineHeight, person.Email)
		pdf.Cell(30, lineHeight, person.Phone)
		pdf.Cell(80, lineHeight, person.Address)
		pdf.Ln(lineHeight)
	}

	outputDir := "pdfs"
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		fmt.Printf("Error creating output directory for %s: %v\n", name, err)
		return
	}

	outputPath := filepath.Join(outputDir, name+".pdf")
	err = pdf.OutputFileAndClose(outputPath)
	if err != nil {
		fmt.Printf("Error saving PDF for %s: %v\n", name, err)
		return
	}

	fmt.Printf("PDF %s generated successfully!\n", outputPath)
}

func generateQRCode(data string) (string, error) {
	qrCode, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return "", err
	}

	tempFile, err := os.CreateTemp("", "qrcode-*.png")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	err = qrCode.WriteFile(256, tempFile.Name())
	if err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}
