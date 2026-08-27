package archive

//
// import (
//	"bytes"
//	"fmt"
//	"image/jpeg"
//	"io"
//
//	"github.com/gen2brain/go-fitz"
//)
//
// type pdfArchive struct {
//	doc   *fitz.Document
//	path  string
//	pages []Page
//}
//
//func openPDF(path string) (Archive, error) {
//	doc, err := fitz.New(path)
//	if err != nil {
//		return nil, fmt.Errorf("failed to open pdf: %w", err)
//	}
//
//	count := doc.NumPage()
//	pages := make([]Page, count)
//	for i := range count {
//		pages[i] = Page{
//			Index:    i,
//			Filename: fmt.Sprintf("page_%04d.jpg", i+1),
//			Size:     0,
//		}
//	}
//
//	return &pdfArchive{
//		doc:   doc,
//		path:  path,
//		pages: pages,
//	}, nil
//}
//
//func (p *pdfArchive) Pages() ([]Page, error) {
//	return p.pages, nil
//}
//
//func (p *pdfArchive) Extract(page Page) (io.ReadCloser, error) {
//	img, err := p.doc.ImageDPI(page.Index, 150.0)
//	if err != nil {
//		return nil, fmt.Errorf("failed to render pdf page %d: %w", page.Index, err)
//	}
//
//	var buf bytes.Buffer
//	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
//		return nil, fmt.Errorf("failed to encode pdf page %d: %w", page.Index, err)
//	}
//
//	return io.NopCloser(&buf), nil
//}
//
//func (p *pdfArchive) Close() error {
//	return p.doc.Close()
//}
