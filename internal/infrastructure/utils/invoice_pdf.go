package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/latiiLA/coop-forex-server/configs"
	"github.com/latiiLA/coop-forex-server/internal/domain/model"
)

const (
	invInkR   = 15
	invInkG   = 23
	invInkB   = 42
	invMutedR = 100
	invMutedG = 116
	invMutedB = 139
	invLineR  = 203
	invLineG  = 213
	invLineB  = 225
	invAccentR = 6
	invAccentG = 147
	invAccentB = 227
	invLeft   = 14.0
	invRight  = 196.0
	invWidth  = 182.0
)

type ReceiptPDFInput struct {
	Request       *model.Request
	ReceiptNumber string
	GeneratedBy   string
}

func resolveReceiptAsset(candidates ...string) string {
	for _, p := range candidates {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func fmtMoney(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func fmtWhen(t *time.Time) string {
	if t == nil {
		return "—"
	}
	return t.Format("02 Jan 2006 15:04")
}

func truncatePDF(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}

// GenerateRequestInvoicePDF renders a single-page official receipt/invoice.
func GenerateRequestInvoicePDF(input ReceiptPDFInput) ([]byte, error) {
	req := input.Request
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(invLeft, 12, 14)
	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	y := drawInvoiceHeader(pdf, input)
	y = drawInvoiceMeta(pdf, req, input, y)
	y = drawInvoiceParties(pdf, req, y)
	y = drawInvoiceAmounts(pdf, req, y)
	y = drawInvoiceTrail(pdf, req, y)
	y = drawInvoiceSignatures(pdf, req, y)
	drawInvoiceFooter(pdf, input, y)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func drawInvoiceHeader(pdf *gofpdf.Fpdf, input ReceiptPDFInput) float64 {
	logo := resolveReceiptAsset(
		filepath.Join("assets", "cooplogo.png"),
		filepath.Join("assets", "coop.gif"),
		filepath.Join("..", "coop-forex", "public", "images", "cooplogo.png"),
	)
	if logo != "" {
		pdf.ImageOptions(logo, invLeft, 10, 16, 0, false, gofpdf.ImageOptions{ReadDpi: true}, 0, "")
	}

	pdf.SetTextColor(invInkR, invInkG, invInkB)
	pdf.SetXY(invLeft+20, 11)
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(90, 6, "FOREX HUB")
	pdf.SetXY(invLeft+20, 17)
	pdf.SetFont("Arial", "", 7.5)
	pdf.SetTextColor(invMutedR, invMutedG, invMutedB)
	pdf.Cell(90, 4, "Cooperative Bank of Oromia")
	pdf.SetXY(invLeft+20, 21)
	pdf.Cell(90, 4, "Official Foreign Currency Receipt")

	pdf.SetTextColor(invAccentR, invAccentG, invAccentB)
	pdf.SetXY(130, 11)
	pdf.SetFont("Arial", "B", 16)
	pdf.CellFormat(66, 7, "RECEIPT", "", 0, "R", false, 0, "")
	pdf.SetTextColor(invInkR, invInkG, invInkB)
	pdf.SetXY(130, 18)
	pdf.SetFont("Arial", "B", 9)
	pdf.CellFormat(66, 5, input.ReceiptNumber, "", 0, "R", false, 0, "")
	pdf.SetXY(130, 23)
	pdf.SetFont("Arial", "", 7)
	pdf.SetTextColor(invMutedR, invMutedG, invMutedB)
	pdf.CellFormat(66, 4, "Date: "+time.Now().Format("02 Jan 2006"), "", 0, "R", false, 0, "")

	pdf.SetDrawColor(invAccentR, invAccentG, invAccentB)
	pdf.SetLineWidth(0.7)
	pdf.Line(invLeft, 29, invRight, 29)
	pdf.SetLineWidth(0.2)
	pdf.SetDrawColor(invLineR, invLineG, invLineB)
	pdf.Line(invLeft, 30.2, invRight, 30.2)

	return 33
}

func drawInvoiceMeta(pdf *gofpdf.Fpdf, req *model.Request, input ReceiptPDFInput, y float64) float64 {
	pdf.SetXY(invLeft, y)
	pdf.SetFont("Arial", "", 8)
	pdf.SetTextColor(invMutedR, invMutedG, invMutedB)
	pdf.Cell(60, 4, "Request code")
	pdf.Cell(62, 4, "Status")
	pdf.Cell(60, 4, "Cross-reference")
	pdf.Ln(4)

	pdf.SetX(invLeft)
	pdf.SetFont("Arial", "B", 9)
	pdf.SetTextColor(invInkR, invInkG, invInkB)
	pdf.Cell(60, 5, truncatePDF(req.RequestCode, 28))
	pdf.Cell(62, 5, truncatePDF(string(req.RequestStatus), 24))
	pdf.Cell(60, 5, truncatePDF(input.ReceiptNumber, 28))
	return y + 12
}

func drawKV(pdf *gofpdf.Fpdf, x, y, w float64, label, value string) float64 {
	pdf.SetXY(x, y)
	pdf.SetFont("Arial", "", 7)
	pdf.SetTextColor(invMutedR, invMutedG, invMutedB)
	pdf.Cell(w, 3.2, label)
	pdf.SetXY(x, y+3.2)
	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(invInkR, invInkG, invInkB)
	pdf.Cell(w, 4, truncatePDF(value, 42))
	return y + 8.5
}

func drawInvoiceParties(pdf *gofpdf.Fpdf, req *model.Request, y float64) float64 {
	colW := 88.0
	gap := 6.0
	boxH := 46.0

	pdf.SetDrawColor(invLineR, invLineG, invLineB)
	pdf.SetFillColor(248, 250, 252)
	pdf.Rect(invLeft, y, colW, boxH, "FD")
	pdf.Rect(invLeft+colW+gap, y, colW, boxH, "FD")

	pdf.SetFont("Arial", "B", 7.5)
	pdf.SetTextColor(invAccentR, invAccentG, invAccentB)
	pdf.SetXY(invLeft+3, y+2)
	pdf.Cell(colW-6, 4, "BILL TO / APPLICANT")
	pdf.SetXY(invLeft+colW+gap+3, y+2)
	pdf.Cell(colW-6, 4, "REQUEST DETAILS")

	leftY := y + 8
	leftY = drawKV(pdf, invLeft+3, leftY, colW-6, "Name", req.ApplicantName)
	leftY = drawKV(pdf, invLeft+3, leftY, colW-6, "Account number", req.ApplicantAccountNumber)
	requestingBranch := "N/A"
	if req.Branch != nil {
		requestingBranch = req.Branch.Name
	}
	leftY = drawKV(pdf, invLeft+3, leftY, colW-6, "Requesting branch", requestingBranch)
	_ = drawKV(pdf, invLeft+3, leftY, colW-6, "Requested as", req.RequestingAs)

	currency := "N/A"
	if req.FcyRequested != nil {
		currency = CurrencyLabel(req.FcyRequested)
	}
	processingBranch := "N/A"
	if req.ProcessingBranch != nil {
		processingBranch = req.ProcessingBranch.Name
	}
	rightY := y + 8
	rightY = drawKV(pdf, invLeft+colW+gap+3, rightY, colW-6, "Currency requested", currency)
	rightY = drawKV(pdf, invLeft+colW+gap+3, rightY, colW-6, "Amount requested", fmtMoney(req.FcyRequestedAmount))
	rightY = drawKV(pdf, invLeft+colW+gap+3, rightY, colW-6, "Processing branch", processingBranch)
	_ = drawKV(pdf, invLeft+colW+gap+3, rightY, colW-6, "Requester / at", PersonName(req.Requester)+" · "+fmtWhen(req.RequestedAt))

	return y + boxH + 4
}

func drawInvoiceAmounts(pdf *gofpdf.Fpdf, req *model.Request, y float64) float64 {
	// Currencies are resolved from approved_currency_ids / accepted_currency_ids in order.
	y = drawAmountBlock(pdf, y, "Approved amounts", req.ApprovedCurrencies, req.ApprovedAmountInCash, req.ApprovedAmountInCard, req.ApprovedAmounts)
	y = drawAmountBlock(pdf, y, "Accepted amounts", req.AcceptedCurrencies, req.AcceptedAmountInCash, req.AcceptedAmountInCard, req.AcceptedAmounts)
	return y
}

func drawAmountBlock(
	pdf *gofpdf.Fpdf,
	y float64,
	title string,
	currencies []model.Currency,
	cashAmounts []float64,
	cardAmounts []float64,
	totalAmounts []float64,
) float64 {
	pdf.SetXY(invLeft, y)
	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(invInkR, invInkG, invInkB)
	pdf.Cell(0, 4, title)
	y += 5

	colW := []float64{70, 37, 37, 38}
	headers := []string{"Currency", "Cash", "Card", "Total"}
	rowH := 5.5

	pdf.SetFillColor(invInkR, invInkG, invInkB)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 7.5)
	x := invLeft
	for i, h := range headers {
		pdf.Rect(x, y, colW[i], rowH, "F")
		pdf.SetXY(x+1.5, y+0.8)
		align := "L"
		if i > 0 {
			align = "R"
		}
		pdf.CellFormat(colW[i]-3, 3.8, h, "", 0, align, false, 0, "")
		x += colW[i]
	}
	y += rowH

	maxRows := len(currencies)
	if len(totalAmounts) > maxRows {
		maxRows = len(totalAmounts)
	}
	if len(cashAmounts) > maxRows {
		maxRows = len(cashAmounts)
	}
	if len(cardAmounts) > maxRows {
		maxRows = len(cardAmounts)
	}
	if maxRows == 0 {
		maxRows = 1
	}
	if maxRows > 3 {
		maxRows = 3
	}

	for i := 0; i < maxRows; i++ {
		if i%2 == 0 {
			pdf.SetFillColor(255, 255, 255)
		} else {
			pdf.SetFillColor(248, 250, 252)
		}
		pdf.SetDrawColor(invLineR, invLineG, invLineB)

		currency := "—"
		if i < len(currencies) {
			currency = CurrencyLabel(&currencies[i])
		}

		cash, card, total := "—", "—", "—"
		if i < len(cashAmounts) {
			cash = fmtMoney(cashAmounts[i])
		}
		if i < len(cardAmounts) {
			card = fmtMoney(cardAmounts[i])
		}
		if i < len(totalAmounts) {
			total = fmtMoney(totalAmounts[i])
		} else if i < len(cashAmounts) || i < len(cardAmounts) {
			var sum float64
			if i < len(cashAmounts) {
				sum += cashAmounts[i]
			}
			if i < len(cardAmounts) {
				sum += cardAmounts[i]
			}
			total = fmtMoney(sum)
		}

		vals := []string{currency, cash, card, total}
		x = invLeft
		for j, v := range vals {
			pdf.Rect(x, y, colW[j], rowH, "FD")
			pdf.SetXY(x+1.5, y+0.8)
			pdf.SetFont("Arial", "", 7.5)
			pdf.SetTextColor(invInkR, invInkG, invInkB)
			align := "L"
			if j > 0 {
				align = "R"
			}
			pdf.CellFormat(colW[j]-3, 3.8, truncatePDF(v, 28), "", 0, align, false, 0, "")
			x += colW[j]
		}
		y += rowH
	}

	return y + 4
}

func drawInvoiceTrail(pdf *gofpdf.Fpdf, req *model.Request, y float64) float64 {
	pdf.SetXY(invLeft, y)
	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(invInkR, invInkG, invInkB)
	pdf.Cell(0, 4, "Authorization trail")
	y += 5

	rows := [][2]string{
		{"Requested", PersonName(req.Requester) + "  ·  " + fmtWhen(req.RequestedAt)},
		{"Authorized", PersonName(req.Authorizer) + "  ·  " + fmtWhen(req.AuthorizedAt)},
		{"Validated", PersonName(req.Validater) + "  ·  " + fmtWhen(req.ValidatedAt)},
		{"Approved", PersonName(req.Approver) + "  ·  " + fmtWhen(req.ApprovedAt)},
		{"Accepted", PersonName(req.Accepter) + "  ·  " + fmtWhen(req.AcceptedAt)},
	}

	processingBranch := "N/A"
	if req.ProcessingBranch != nil {
		processingBranch = req.ProcessingBranch.Name
	}
	rows = append(rows, [2]string{"Processor branch", processingBranch})

	rowH := 4.8
	for i, row := range rows {
		if i%2 == 0 {
			pdf.SetFillColor(248, 250, 252)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		pdf.SetDrawColor(invLineR, invLineG, invLineB)
		pdf.Rect(invLeft, y, invWidth, rowH, "FD")
		pdf.SetXY(invLeft+2, y+0.6)
		pdf.SetFont("Arial", "B", 7)
		pdf.SetTextColor(invMutedR, invMutedG, invMutedB)
		pdf.Cell(34, 3.5, row[0])
		pdf.SetFont("Arial", "", 7.5)
		pdf.SetTextColor(invInkR, invInkG, invInkB)
		pdf.Cell(invWidth-38, 3.5, truncatePDF(row[1], 72))
		y += rowH
	}
	return y + 4
}

func drawInvoiceSignatures(pdf *gofpdf.Fpdf, req *model.Request, y float64) float64 {
	boxW := 88.0
	boxH := 32.0
	gap := 6.0

	pdf.SetDrawColor(invLineR, invLineG, invLineB)
	pdf.Rect(invLeft, y, boxW, boxH, "D")
	pdf.Rect(invLeft+boxW+gap, y, boxW, boxH, "D")

	pdf.SetFont("Arial", "B", 7)
	pdf.SetTextColor(invMutedR, invMutedG, invMutedB)
	pdf.SetXY(invLeft+2.5, y+2)
	pdf.Cell(boxW-5, 3, "APPROVER SIGNATURE")
	pdf.SetXY(invLeft+boxW+gap+2.5, y+2)
	pdf.Cell(boxW-5, 3, "OFFICIAL STAMP")

	sigPath := ""
	if req.Approver != nil && req.Approver.Signature != nil && *req.Approver.Signature != "" {
		sig := *req.Approver.Signature
		sigPath = resolveReceiptAsset(
			sig,
			filepath.Join(configs.FileUploadPath, filepath.Base(sig)),
			filepath.Join("uploads", filepath.Base(sig)),
			filepath.Join("assets", "signature.jpg"),
			filepath.Join("..", "coop-forex", "public", "images", "signature.jpg"),
		)
	}
	stampPath := resolveReceiptAsset(
		filepath.Join(configs.FileUploadPath, "reciept_files", "stamp.jpeg"),
		filepath.Join("uploads", "reciept_files", "stamp.jpeg"),
		filepath.Join("assets", "stamp.jpg"),
		filepath.Join("..", "coop-forex", "public", "images", "stamp.jpg"),
	)

	// Fixed max height keeps the document on one page.
	if sigPath != "" {
		pdf.ImageOptions(sigPath, invLeft+18, y+6, 0, 16, false, gofpdf.ImageOptions{ReadDpi: true}, 0, "")
	}
	if stampPath != "" {
		pdf.ImageOptions(stampPath, invLeft+boxW+gap+26, y+5, 0, 18, false, gofpdf.ImageOptions{ReadDpi: true}, 0, "")
	}

	pdf.SetDrawColor(invLineR, invLineG, invLineB)
	pdf.Line(invLeft+8, y+boxH-7, invLeft+boxW-8, y+boxH-7)
	pdf.Line(invLeft+boxW+gap+8, y+boxH-7, invLeft+boxW+gap+boxW-8, y+boxH-7)

	pdf.SetFont("Arial", "", 7.5)
	pdf.SetTextColor(invInkR, invInkG, invInkB)
	pdf.SetXY(invLeft+2, y+boxH-5.5)
	pdf.CellFormat(boxW-4, 3.5, PersonName(req.Approver), "", 0, "C", false, 0, "")
	pdf.SetXY(invLeft+boxW+gap+2, y+boxH-5.5)
	pdf.CellFormat(boxW-4, 3.5, "Authorized stamp", "", 0, "C", false, 0, "")

	return y + boxH + 4
}

func drawInvoiceFooter(pdf *gofpdf.Fpdf, input ReceiptPDFInput, y float64) {
	if y < 248 {
		y = 248
	}
	if y > 258 {
		y = 258
	}

	pdf.SetDrawColor(invAccentR, invAccentG, invAccentB)
	pdf.SetLineWidth(0.5)
	pdf.Line(invLeft, y, invRight, y)

	pdf.SetXY(invLeft, y+2.5)
	pdf.SetFont("Arial", "B", 8)
	pdf.SetTextColor(invInkR, invInkG, invInkB)
	pdf.Cell(0, 4, "Document reference")

	pdf.SetXY(invLeft, y+7)
	pdf.SetFont("Arial", "", 7)
	pdf.SetTextColor(invMutedR, invMutedG, invMutedB)
	pdf.MultiCell(invWidth, 3.3,
		fmt.Sprintf(
			"Receipt No: %s   |   Request: %s   |   Issued %s by %s.\nThis document is an official Forex Hub FCY approval receipt for documentation purposes.",
			input.ReceiptNumber,
			input.Request.RequestCode,
			time.Now().Format("02 Jan 2006 15:04"),
			input.GeneratedBy,
		),
		"", "L", false,
	)
}
