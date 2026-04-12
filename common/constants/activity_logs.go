package constants

type ActivityLogStatus string

type LogAction string

const (
	LOG_STATUS_SUCCESS ActivityLogStatus = "success"
	LOG_STATUS_FAILED  ActivityLogStatus = "failed"
)

// Order Log Actions
const (
	LOG_ACTION_CREATE_ORDER        LogAction = "Create Order"
	LOG_ACTION_CHANGE_ORDER_STATUS LogAction = "Change Order Status"
	LOG_ACTION_REFUND_ORDER        LogAction = "Refund Order"
)

const (
	LOG_ACTION_LOGIN               LogAction = "Login"
	LOG_ACTION_STAFF_LOGIN         LogAction = "Staff Login"
	LOG_ACTION_LOGOUT              LogAction = "Logout"
	LOG_ACTION_STAFF_LOGOUT        LogAction = "Staff Logout"
	LOG_ACTION_UPDATE_STOCK_REPORT LogAction = "Update Stock Report"
	LOG_ACTION_RENEW_JWT_TOKEN     LogAction = "Renew JWT Token"
)

// GrabFood Log Actions
const (
	LOG_ACTION_GRABFOOD_MENU_SYNC_FAILED LogAction = "GrabFood Menu Sync Failed"
)

// GrabExpress Log Actions
const (
	LOG_ACTION_GRABEXPRESS_QUOTE_DELIVERY   LogAction = "GrabExpress Quote Delivery"
	LOG_ACTION_GRABEXPRESS_CREATE_DELIVERY  LogAction = "GrabExpress Create Delivery"
	LOG_ACTION_GRABEXPRESS_TRACKING_WEBHOOK LogAction = "GrabExpress Tracking Webhook"
)

// ShopeeFood Log Actions
const (
	LOG_ACTION_SHOPEEFOOD_ORDER_SUBMISSION    LogAction = "[ShopeeFood] Order Submission"
	LOG_ACTION_SHOPEEFOOD_MENU_SYNC           LogAction = "[ShopeeFood] Menu Sync"
	LOG_ACTION_SHOPEEFOOD_MENU_GET_MENU       LogAction = "[ShopeeFood] Get Menu Request"
	LOG_ACTION_SHOPEEFOOD_ORDER_STATUS_UPDATE LogAction = "[ShopeeFood] Order Status Update Request"
)

// Cloud ECR Log Actions
const (
	LOG_ACTION_CLOUD_ECR_TRANSFER_REQUEST  LogAction = "[Cloud ECR] Transfer Request"
	LOG_ACTION_CLOUD_ECR_TRANSFER_RESPONSE LogAction = "[Cloud ECR] Transfer Response"
	LOG_ACTION_CLOUD_ECR_TRANSFER_ERROR    LogAction = "[Cloud ECR] Transfer Error"
)

// IOS Live Activity Log Actions
const (
	LOG_ACTION_IOS_LIVE_ACTIVITY_STORE_TOKEN       LogAction = "[IOS Live Activity] Store Token"
	LOG_ACTION_IOS_LIVE_ACTIVITY_SEND_NOTIFICATION LogAction = "[IOS Live Activity] Send Notification"
)

// Payment Log Actions
const (
	LOG_ACTION_PAYMENT_CALLBACK                LogAction = "Payment Callback"
	LOG_ACTION_PAYMENT_NOTIFICATION_SUCCESS    LogAction = "Payment Notification Success"
	LOG_ACTION_PAYMENT_NOTIFICATION_FAILED     LogAction = "Payment Notification Failed"
	LOG_ACTION_PAYMENT_SUCCESS_WEBHOOK_SUCCESS LogAction = "Payment Success Webhook Success"
	LOG_ACTION_PAYMENT_SUCCESS_WEBHOOK_FAILED  LogAction = "Payment Success Webhook Failed"
)
