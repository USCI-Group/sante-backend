package customer_orders

import (
	"fmt"
	"net/http"

	"encore.app/common"
	"encore.app/common/pdf"
	"encore.app/customers/customer_common"
)

// encore:api public raw method=GET path=/api/invoice
func (s *Service) GetInvoicePDFHandler(w http.ResponseWriter, r *http.Request) {

	orderID := r.URL.Query().Get("order_id")

	orderIDUUID, err := common.StringToUUID(orderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	order, err := customer_common.GetOrderByID(s.db, orderIDUUID, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pdfBytes, err := pdf.GenerateInvoicePDF(order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileName := fmt.Sprintf("invoice_%s.pdf", order.InvoiceNumber)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s", fileName))
	w.WriteHeader(http.StatusOK)
	w.Write(pdfBytes)
}
