package constants

type PaymentMethod string

const (
	PaymentMethodFPX        PaymentMethod = "fpx"
	PaymentMethodEWallet    PaymentMethod = "e-wallet"
	PaymentMethodCash       PaymentMethod = "cash"
	PaymentMethodStaticQR   PaymentMethod = "static_qr"
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	PaymentMethodDebitCard  PaymentMethod = "debit_card"
	PaymentMethodBNPL       PaymentMethod = "bnpl"
)
