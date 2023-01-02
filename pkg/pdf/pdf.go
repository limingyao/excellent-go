package pdf

import (
	"github.com/limingyao/excellent-go/pkg/fonts"
	"github.com/signintech/gopdf"
)

// New a4 pdf with msyh font
func New(fontSize int) (*gopdf.GoPdf, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})

	if err := pdf.AddTTFFontByReader("msyh", fonts.MsyhReader()); err != nil {
		return nil, err
	}
	if err := pdf.SetFont("msyh", "", fontSize); err != nil {
		return nil, err
	}

	return &pdf, nil
}
