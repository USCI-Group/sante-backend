package constants

type GrabFoodAvailableStatus string

const (
	GRABFOOD_AVAILABLE   GrabFoodAvailableStatus = "AVAILABLE"
	GRABFOOD_UNAVAILABLE GrabFoodAvailableStatus = "UNAVAILABLE"
	GRABFOOD_SOLD_OUT    GrabFoodAvailableStatus = "UNAVAILABLETODAY"
)

type GrabFoodMenuFieldType string

const (
	GRABFOOD_MENU_FIELD_TYPE_PRODUCT  GrabFoodMenuFieldType = "ITEM"
	GRABFOOD_MENU_FIELD_TYPE_MODIFIER GrabFoodMenuFieldType = "MODIFIER"
)

type GrabFoodOrderState string

const (
	GRABFOOD_ORDER_STATE_ACCEPTED         GrabFoodOrderState = "ACCEPTED"
	GRABFOOD_ORDER_STATE_DRIVER_ALLOCATED GrabFoodOrderState = "DRIVER_ALLOCATED"
	GRABFOOD_ORDER_STATE_DRIVER_ARRIVED   GrabFoodOrderState = "DRIVER_ARRIVED"
	GRABFOOD_ORDER_STATE_COLLECTED        GrabFoodOrderState = "COLLECTED"
	GRABFOOD_ORDER_STATE_DELIVERED        GrabFoodOrderState = "DELIVERED"
	GRABFOOD_ORDER_STATE_FAILED           GrabFoodOrderState = "FAILED"
	GRABFOOD_ORDER_STATE_CANCELLED        GrabFoodOrderState = "CANCELLED"
)

type GrabExpressDeliveryStatus string

const (
	GRABEXPRESS_STATUS_ALLOCATING       GrabExpressDeliveryStatus = "ALLOCATING"
	GRABEXPRESS_STATUS_PENDING_PICKUP   GrabExpressDeliveryStatus = "PENDING_PICKUP"
	GRABEXPRESS_STATUS_PICKING_UP       GrabExpressDeliveryStatus = "PICKING_UP"
	GRABEXPRESS_STATUS_PENDING_DROP_OFF GrabExpressDeliveryStatus = "PENDING_DROP_OFF"
	GRABEXPRESS_STATUS_IN_DELIVERY      GrabExpressDeliveryStatus = "IN_DELIVERY"
	GRABEXPRESS_STATUS_IN_RETURN        GrabExpressDeliveryStatus = "IN_RETURN"
	GRABEXPRESS_STATUS_COMPLETED        GrabExpressDeliveryStatus = "COMPLETED"
	GRABEXPRESS_STATUS_CANCELED         GrabExpressDeliveryStatus = "CANCELED"
	GRABEXPRESS_STATUS_RETURNED         GrabExpressDeliveryStatus = "RETURNED"
	GRABEXPRESS_STATUS_FAILED           GrabExpressDeliveryStatus = "FAILED"
)

type GrabFoodMenuSyncState string

const (
	GRABFOOD_MENU_SYNC_STATE_QUEUEING   GrabFoodMenuSyncState = "QUEUEING"
	GRABFOOD_MENU_SYNC_STATE_PROCESSING GrabFoodMenuSyncState = "PROCESSING"
	GRABFOOD_MENU_SYNC_STATE_SUCCESS    GrabFoodMenuSyncState = "SUCCESS"
	GRABFOOD_MENU_SYNC_STATE_FAILED     GrabFoodMenuSyncState = "FAILED"
)
