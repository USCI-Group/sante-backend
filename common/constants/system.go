package constants

type SystemDataInfoType string

const (
	SystemDataInfoTypeERPClientID                 SystemDataInfoType = "erp_client_id"     // For E-Invoice
	SystemDataInfoTypeERPClientSecret             SystemDataInfoType = "erp_client_secret" // For E-Invoice (Encrypted)
	SystemDataInfoTypeMainTaxpayerTIN             SystemDataInfoType = "main_taxpayer_tin"
	SystemDataInfoTypeGrabFoodPartnerClientID     SystemDataInfoType = "grabfood_partner_client_id"
	SystemDataInfoTypeGrabFoodPartnerClientSecret SystemDataInfoType = "grabfood_partner_client_secret"
	SystemDataInfoTypePretzleyAPNsPrivateKey      SystemDataInfoType = "pretzley_apns_private_key"
	SystemDataInfoTypeAnchorSMSUser               SystemDataInfoType = "anchor_sms_user"
	SystemDataInfoTypeAnchorSMSPassword           SystemDataInfoType = "anchor_sms_password"
)
