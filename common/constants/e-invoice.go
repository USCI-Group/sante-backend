package constants

type IDType string
type EInvoiceClassificationCode string
type EInvoiceTaxableType string
type EInvoiceType string

type EInvoiceStatus string

const (
	EInvoiceStatusQueued    EInvoiceStatus = "QUEUED"
	EInvoiceStatusSubmitted EInvoiceStatus = "SUBMITTED" // Submitted to LHDN (in progress)
	EInvoiceStatusRejected  EInvoiceStatus = "REJECTED"  // Rejected by LHDN (submission rejected)
	EInvoiceStatusValid     EInvoiceStatus = "VALID"     // Validated by LHDN (final success)
	EInvoiceStatusInvalid   EInvoiceStatus = "INVALID"   // Invalidated by LHDN (final failure)
	EInvoiceStatusFailed    EInvoiceStatus = "FAILED"    // Failed to submit to LHDN (submission failed)
)

const (
	DEFAULT_E_INVOICE_BUYER_TIN = "EI00000000010"
)

const (
	IDTypeNRIC     IDType = "NRIC"
	IDTypeBRN      IDType = "BRN"
	IDTypePassport IDType = "PASSPORT"
	IDTypeArmy     IDType = "ARMY"
)

const (
	BREASTFEEDING_EQUIPMENT                                                                                                                                                                                                                                                                                                           EInvoiceClassificationCode = "001"
	CHILD_CARE_CENTRES_AND_KINDERGARTENS_FEES                                                                                                                                                                                                                                                                                         EInvoiceClassificationCode = "002"
	COMPUTER_SMARTPHONE_OR_TABLET                                                                                                                                                                                                                                                                                                     EInvoiceClassificationCode = "003"
	CONSOLIDATED_E_INVOICE                                                                                                                                                                                                                                                                                                            EInvoiceClassificationCode = "004"
	CONSTRUCTION_MATERIALS                                                                                                                                                                                                                                                                                                            EInvoiceClassificationCode = "005"
	DISBURSEMENT                                                                                                                                                                                                                                                                                                                      EInvoiceClassificationCode = "006"
	DONATION                                                                                                                                                                                                                                                                                                                          EInvoiceClassificationCode = "007"
	E_COMMERCE_E_INVOICE_TO_BUYER_PURCHASER                                                                                                                                                                                                                                                                                           EInvoiceClassificationCode = "008"
	E_COMMERCE_SELF_BILLED_E_INVOICE_TO_SELLER_LOGISTICS_ETC                                                                                                                                                                                                                                                                          EInvoiceClassificationCode = "009"
	EDUCATION_FEES                                                                                                                                                                                                                                                                                                                    EInvoiceClassificationCode = "010"
	GOODS_ON_CONSIGNMENT_CONSIGNOR                                                                                                                                                                                                                                                                                                    EInvoiceClassificationCode = "011"
	GOODS_ON_CONSIGNMENT_CONSIGNEE                                                                                                                                                                                                                                                                                                    EInvoiceClassificationCode = "012"
	GYM_MEMBERSHIP                                                                                                                                                                                                                                                                                                                    EInvoiceClassificationCode = "013"
	INSURANCE_EDUCATION_AND_MEDICAL_BENEFITS                                                                                                                                                                                                                                                                                          EInvoiceClassificationCode = "014"
	INSURANCE_TAKAFUL_OR_LIFE_INSURANCE                                                                                                                                                                                                                                                                                               EInvoiceClassificationCode = "015"
	INTEREST_AND_FINANCING_EXPENSES                                                                                                                                                                                                                                                                                                   EInvoiceClassificationCode = "016"
	INTERNET_SUBSCRIPTION                                                                                                                                                                                                                                                                                                             EInvoiceClassificationCode = "017"
	LAND_AND_BUILDING                                                                                                                                                                                                                                                                                                                 EInvoiceClassificationCode = "018"
	MEDICAL_EXAMINATION_FOR_LEARNING_DISABILITIES_AND_EARLY_INTERVENTION_OR_REHABILITATION_TREATMENTS_OF_LEARNING_DISABILITIES                                                                                                                                                                                                        EInvoiceClassificationCode = "019"
	MEDICAL_EXAMINATION_OR_VACCINATION_EXPENSES                                                                                                                                                                                                                                                                                       EInvoiceClassificationCode = "020"
	MEDICAL_EXPENSES_FOR_SERIOUS_DISEASES                                                                                                                                                                                                                                                                                             EInvoiceClassificationCode = "021"
	OTHERS                                                                                                                                                                                                                                                                                                                            EInvoiceClassificationCode = "022"
	PETROLEUM_OPERATIONS                                                                                                                                                                                                                                                                                                              EInvoiceClassificationCode = "023"
	PRIVATE_RETIREMENT_SCHEME_OR_DEFERRED_ANNUITY_SCHEME                                                                                                                                                                                                                                                                              EInvoiceClassificationCode = "024"
	MOTOR_VEHICLE                                                                                                                                                                                                                                                                                                                     EInvoiceClassificationCode = "025"
	SUBSCRIPTION_OF_BOOKS_JOURNALS_MAGAZINES_NEWSPAPERS_OR_OTHER_SIMILAR_PUBLICATIONS                                                                                                                                                                                                                                                 EInvoiceClassificationCode = "026"
	REIMBURSEMENT                                                                                                                                                                                                                                                                                                                     EInvoiceClassificationCode = "027"
	RENTAL_OF_MOTOR_VEHICLE                                                                                                                                                                                                                                                                                                           EInvoiceClassificationCode = "028"
	EV_CHARGING_FACILITIES_INSTALLATION_RENTAL_SALE_PURCHASE_OR_SUBSCRIPTION_FEES                                                                                                                                                                                                                                                     EInvoiceClassificationCode = "029"
	REPAIR_AND_MAINTENANCE                                                                                                                                                                                                                                                                                                            EInvoiceClassificationCode = "030"
	RESEARCH_AND_DEVELOPMENT                                                                                                                                                                                                                                                                                                          EInvoiceClassificationCode = "031"
	FOREIGN_INCOME                                                                                                                                                                                                                                                                                                                    EInvoiceClassificationCode = "032"
	SELF_BILLED_BETTING_AND_GAMING                                                                                                                                                                                                                                                                                                    EInvoiceClassificationCode = "033"
	SELF_BILLED_IMPORTATION_OF_GOODS                                                                                                                                                                                                                                                                                                  EInvoiceClassificationCode = "034"
	SELF_BILLED_IMPORTATION_OF_SERVICES                                                                                                                                                                                                                                                                                               EInvoiceClassificationCode = "035"
	SELF_BILLED_OTHERS                                                                                                                                                                                                                                                                                                                EInvoiceClassificationCode = "036"
	SELF_BILLED_MONETARY_PAYMENT_TO_AGENTS_DEALERS_OR_DISTRIBUTORS                                                                                                                                                                                                                                                                    EInvoiceClassificationCode = "037"
	SPORTS_EQUIPMENT_RENTAL_ENTRY_FEES_FOR_SPORTS_FACILITIES_REGISTRATION_IN_SPORTS_COMPETITION_OR_SPORTS_TRAINING_FEES_IMPOSED_BY_ASSOCIATIONS_SPORTS_CLUBS_COMPANIES_REGISTERED_WITH_THE_SPORTS_COMMISSIONER_OR_COMPANIES_COMMISSION_OF_MALAYSIA_AND_CARRIING_OUT_SPORTS_ACTIVITIES_AS_LISTED_UNDER_THE_SPORTS_DEVELOPMENT_ACT_1997 EInvoiceClassificationCode = "038"
	SUPPORTING_EQUIPMENT_FOR_DISABLED_PERSON                                                                                                                                                                                                                                                                                          EInvoiceClassificationCode = "039"
	VOLUNTARY_CONTRIBUTION_TO_APPROVED_PROVIDENT_FUND                                                                                                                                                                                                                                                                                 EInvoiceClassificationCode = "040"
	DENTAL_EXAMINATION_OR_TREATMENT                                                                                                                                                                                                                                                                                                   EInvoiceClassificationCode = "041"
	FERTILITY_TREATMENT                                                                                                                                                                                                                                                                                                               EInvoiceClassificationCode = "042"
	TREATMENT_AND_HOME_CARE_NURSING_DAYCARE_CENTRES_AND_RESIDENTIAL_CARE_CENTERS                                                                                                                                                                                                                                                      EInvoiceClassificationCode = "043"
	VOUCHERS_GIFT_CARDS_LOYALTY_POINTS_ETC                                                                                                                                                                                                                                                                                            EInvoiceClassificationCode = "044"
	SELF_BILLED_NON_MONETARY_PAYMENT_TO_AGENTS_DEALERS_OR_DISTRIBUTORS                                                                                                                                                                                                                                                                EInvoiceClassificationCode = "045"
)

const (
	SalesTax                EInvoiceTaxableType = "01"
	ServiceTax              EInvoiceTaxableType = "02"
	TourismTax              EInvoiceTaxableType = "03"
	HighValueGoodsTax       EInvoiceTaxableType = "04"
	SalesTaxOnLowValueGoods EInvoiceTaxableType = "05"
	NotApplicable           EInvoiceTaxableType = "06"
	TaxExemption            EInvoiceTaxableType = "E"
)

const (
	Invoice              EInvoiceType = "01"
	CreditNote           EInvoiceType = "02"
	DebitNote            EInvoiceType = "03"
	RefundNote           EInvoiceType = "04"
	SelfBilledInvoice    EInvoiceType = "11"
	SelfBilledCreditNote EInvoiceType = "12"
	SelfBilledDebitNote  EInvoiceType = "13"
	SelfBilledRefundNote EInvoiceType = "14"
)
