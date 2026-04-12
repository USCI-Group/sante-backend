package constants

type ShopeeFoodAvailableStatus bool

const (
	SHOPEEFOOD_AVAILABLE   ShopeeFoodAvailableStatus = true
	SHOPEEFOOD_UNAVAILABLE ShopeeFoodAvailableStatus = false
)

type ShopeeFoodOrderState string

const (
	SHOPEEFOOD_ORDER_STATE_ACCEPTED               ShopeeFoodOrderState = "ACCEPT"
	SHOPEEFOOD_ORDER_STATE_REJECT                 ShopeeFoodOrderState = "REJECT"
	SHOPEEFOOD_ORDER_STATE_PICKED_UP              ShopeeFoodOrderState = "DELIVERY_PICKED_UP"
	SHOPEEFOOD_ORDER_STATE_DELIVERED              ShopeeFoodOrderState = "DELIVERED"
	SHOPEEFOOD_ORDER_STATE_CANCELED               ShopeeFoodOrderState = "CANCELED"
	SHOPEEFOOD_ORDER_STATE_CANCELING              ShopeeFoodOrderState = "CANCELING"
	SHOPEEFOOD_ORDER_STATE_CANCEL                 ShopeeFoodOrderState = "CANCEL"
	SHOPEEFOOD_ORDER_STATE_COMPLETED              ShopeeFoodOrderState = "COMPLETED"
	SHOPEEFOOD_ORDER_STATE_DELIVERY_ARRIVED_STORE ShopeeFoodOrderState = "DELIVERY_ARRIVED_STORE"
)
