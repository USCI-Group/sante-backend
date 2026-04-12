package pdf

import (
	"bytes"
	_ "embed"
	"fmt"

	"encore.app/common"
	"encore.app/customers/customer_common"
	"encore.app/database/models"
	"github.com/jung-kurt/gofpdf"
)

//go:embed images/pretzley_logo.png
var logoPNG []byte

func GenerateInvoicePDF(order *models.Order) ([]byte, error) {
	invoiceID := order.InvoiceNumber
	invoiceDate, err := common.GetTimeInMalaysiaTimezone(order.InvoiceDate)
	if err != nil {
		return nil, err
	}
	invoiceDateString := invoiceDate.Format("Jan 02, 2006")
	companyAddress := customer_common.CombineAddress(order.Outlet.Address)
	orderType := order.OrderType
	paymentMethod := order.PaymentMethod
	customerName := ""
	if order.OrderDetails != nil {
		customerName = *order.OrderDetails.CustomerName
	}
	pointsRewarded := 0
	if order.PointsRewarded != nil {
		pointsRewarded = *order.PointsRewarded
	}
	expRewarded := 0
	if order.ExpRewarded != nil {
		expRewarded = *order.ExpRewarded
	}

	lineSpacing := 8.0

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 18, 20)
	pdf.AddPage()

	// Header Section
	// Draw a line
	// pdf.SetLineWidth(0.5)
	// pdf.SetDrawColor(128, 128, 128) // grey
	// pdf.Line(20, 36, 190, 36)

	pdf.RegisterImageOptionsReader(
		"logo", // image alias (any string)
		gofpdf.ImageOptions{
			ImageType: "PNG", // MUST match actual format
			ReadDpi:   true,
		},
		bytes.NewReader(logoPNG),
	)

	// Follow the margin: Use pdf.GetMargins() to compute correct x for image placement
	left, top, right, _ := pdf.GetMargins()
	imageWidth := 30.0
	// A4 width is 210mm, so printable area is 210 - left - imageWidth
	x := 210 - left - imageWidth
	y := top // stick to Y margin
	pdf.Image(
		"logo", // alias, NOT path
		x, y,   // x, y - align right within margins
		imageWidth, 0, // width, height (0 = auto)
		false,
		"",
		0,
		"",
	)

	currentY := top + 5

	pdf.SetFont("Arial", "", 11)
	if customerName != "" {
		pdf.CellFormat(0, 5, fmt.Sprintf("Customer Name: %s", customerName), "", 0, "L", false, 0, "")
		currentY += 5
	}

	pdf.SetY(currentY)
	// Company Info
	pdf.SetTextColor(0, 0, 0)
	pdf.MultiCell(80, 5, companyAddress, "", "L", false)

	// Invoice title & meta
	currentY += 10
	pdf.SetY(currentY)

	pdf.Ln(lineSpacing)
	pdf.SetFont("Arial", "", 11)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(0, 6, fmt.Sprintf("Invoice number: %s", invoiceID), "", 0, "L", false, 0, "")
	pdf.Ln(lineSpacing)
	pdf.CellFormat(0, 6, fmt.Sprintf("Date: %s", invoiceDateString), "", 0, "L", false, 0, "")
	pdf.Ln(lineSpacing)
	pdf.CellFormat(0, 6, fmt.Sprintf("Order type: %s", orderType), "", 0, "L", false, 0, "")
	pdf.Ln(lineSpacing)
	pdf.CellFormat(0, 6, fmt.Sprintf("Payment method: %s", paymentMethod), "", 0, "L", false, 0, "")

	currentY += lineSpacing*5 + 3
	pdf.SetY(currentY)

	// Order details
	pdf.SetTextColor(150, 150, 150) // lighter grey
	pdf.CellFormat(0, 6, "Order details:", "", 0, "L", false, 0, "")

	currentY += lineSpacing
	pdf.SetY(currentY)

	// Table Headers
	pdf.SetFillColor(255, 255, 255)
	pdf.SetTextColor(0, 0, 0)
	pdf.SetDrawColor(220, 220, 220) // lighter grey border
	pdf.SetLineWidth(0.3)

	columnWidth := (210.0 - left - right) / 5
	colWidths := []float64{columnWidth * 3, columnWidth, columnWidth}
	headers := []string{"Item", "Qty", "Price"}
	pdf.CellFormat(colWidths[0], 10, headers[0], "TB", 0, "L", true, 0, "")
	pdf.CellFormat(colWidths[1], 10, headers[1], "TB", 0, "R", true, 0, "")
	pdf.CellFormat(colWidths[2], 10, headers[2], "TB", 0, "R", true, 0, "")
	pdf.Ln(-1)

	// Table Rows
	pdf.SetFont("Arial", "", 11)
	// pdf.SetFillColor(240, 240, 245)

	for idx, item := range order.OrderItems {
		fill := idx%2 == 0

		pdf.SetTextColor(0, 0, 0)
		pdf.SetFont("Arial", "", 11)

		pdf.CellFormat(colWidths[0], lineSpacing, item.Product.Name, "T", 0, "L", fill, 0, "")
		pdf.CellFormat(colWidths[1], lineSpacing, fmt.Sprintf("x%d", item.Quantity), "T", 0, "R", fill, 0, "")
		pdf.CellFormat(colWidths[2], lineSpacing, fmt.Sprintf("RM %.2f", item.SubTotal), "T", 0, "R", fill, 0, "")
		pdf.Ln(-1)

		pdf.SetTextColor(150, 150, 150)
		pdf.SetFont("Arial", "", 11)

		modifiers := customer_common.CombineModifiers(item.SelectedModifierGroups)

		pdf.MultiCell(colWidths[0], lineSpacing, modifiers, "", "L", fill)
		// pdf.Ln(-1)

	}

	deliveryFee := 0.0
	if order.DeliveryFee != nil {
		deliveryFee = float64(*order.DeliveryFee)
	}

	// Total Section
	pdf.SetTextColor(0, 0, 0)
	pdf.Ln(1)
	pdf.CellFormat(colWidths[0], lineSpacing, "", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[1], lineSpacing, "Points collected", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[2], lineSpacing, fmt.Sprintf("%d pts", pointsRewarded), "0", 1, "R", false, 0, "")

	pdf.CellFormat(colWidths[0], lineSpacing, "", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[1], lineSpacing, "Experience collected", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[2], lineSpacing, fmt.Sprintf("%d exp", expRewarded), "0", 1, "R", false, 0, "")

	pdf.CellFormat(colWidths[0], lineSpacing, "", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[1], lineSpacing, "Amount (inc.SST% 6)", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[2], lineSpacing, fmt.Sprintf("RM %.2f", order.GrossTotal), "0", 1, "R", false, 0, "")

	pdf.CellFormat(colWidths[0], lineSpacing, "", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[1], lineSpacing, "Voucher", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[2], lineSpacing, fmt.Sprintf("RM %.2f", order.DiscountAmount), "0", 1, "R", false, 0, "")

	pdf.CellFormat(colWidths[0], lineSpacing, "", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[1], lineSpacing, "Sub Total", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[2], lineSpacing, fmt.Sprintf("RM %.2f", order.GrossTotal), "0", 1, "R", false, 0, "")

	if order.OrderType == models.OrderTypeDelivery {
		pdf.CellFormat(colWidths[0], lineSpacing, "", "0", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[1], lineSpacing, "Delivery Fee", "0", 0, "R", false, 0, "")
		pdf.CellFormat(colWidths[2], lineSpacing, fmt.Sprintf("RM %.2f", deliveryFee), "0", 1, "R", false, 0, "")
	}

	pdf.CellFormat(colWidths[0], lineSpacing, "", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[1], lineSpacing, "Rounding Adj", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[2], lineSpacing, fmt.Sprintf("RM %.2f", order.RoundedAmount), "0", 1, "R", false, 0, "")

	pdf.CellFormat(colWidths[0], lineSpacing, "", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[1], lineSpacing, "6% SST", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[2], lineSpacing, fmt.Sprintf("RM %.2f", order.TaxCharge), "0", 1, "R", false, 0, "")

	// Set text color to purple (RGB: 128, 0, 128) for Grand Total row
	pdf.SetTextColor(128, 0, 128)
	pdf.SetFont("Arial", "B", 11)
	pdf.CellFormat(colWidths[0], lineSpacing, "", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[1], lineSpacing, "Grand total", "0", 0, "R", false, 0, "")
	pdf.CellFormat(colWidths[2], lineSpacing, fmt.Sprintf("RM %.2f", order.RoundedNetTotal), "0", 1, "R", false, 0, "")

	var buf bytes.Buffer
	if err = pdf.Output(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
