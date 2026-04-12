package customer_type

import (
	"encore.dev/types/uuid"
)

type ProductWithQuantity struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
}
