package yoocommon

type PaymentModeType string

const (
	PaymentSubjectTypeCommodity            PaymentSubjectType = "commodity"
	PaymentSubjectTypeExcise               PaymentSubjectType = "excise"
	PaymentSubjectTypeJob                  PaymentSubjectType = "job"
	PaymentSubjectTypeService              PaymentSubjectType = "service"
	PaymentSubjectTypeGamblingBet          PaymentSubjectType = "gambling_bet"
	PaymentSubjectTypeGamblingPrize        PaymentSubjectType = "gambling_prize"
	PaymentSubjectTypeLottery              PaymentSubjectType = "lottery"
	PaymentSubjectTypeLotteryPrize         PaymentSubjectType = "lottery_prize"
	PaymentSubjectTypeIntellectualActivity PaymentSubjectType = "intellectual_activity"
	PaymentSubjectTypePayment              PaymentSubjectType = "payment"
	PaymentSubjectTypeAgentCommission      PaymentSubjectType = "agent_commission"
	PaymentSubjectTypePropertyRight        PaymentSubjectType = "property_right"
	PaymentSubjectTypenOnOperatingGain     PaymentSubjectType = "non_operating_gain"
	PaymentSubjectTypeInsurancePremium     PaymentSubjectType = "insurance_premium"
	PaymentSubjectTypeSalesTax             PaymentSubjectType = "sales_tax"
	PaymentSubjectTypeResortFee            PaymentSubjectType = "resort_fee"
	PaymentSubjectTypeComposite            PaymentSubjectType = "composite"
	PaymentSubjectTypeAnother              PaymentSubjectType = "another"
)

type PaymentSubjectType string

const (
	PaymentModeTypeFullPrepayment    PaymentModeType = "full_prepayment"
	PaymentModeTypePartialPrepayment PaymentModeType = "partial_prepayment"
	PaymentModeTypeAdvance           PaymentModeType = "advance"
	PaymentModeTypeFullPayment       PaymentModeType = "full_payment"
	PaymentModeTypePartialPayment    PaymentModeType = "partial_payment"
	PaymentModeTypeCredit            PaymentModeType = "credit"
	PaymentModeTypeCreditPayment     PaymentModeType = "credit_payment"
)

type Item struct {
	// parameter with the name of the product or service
	Description string `json:"description"`

	// parameter with the amount per unit of product
	Quantity float64 `json:"quantity"`

	// parameter specifying the quantity of goods (only integers, for example 1)
	Amount *Amount `json:"amount"`

	// parameter with the fixed value 1 (price without VAT)
	VatCode int `json:"vat_code"`

	//parameter is the product’s category for the Tax Service.
	PaymentMode PaymentModeType `json:"payment_mode"`

	//parameter is the payment method’s category for the Tax Service.
	PaymentSubject PaymentSubjectType `json:"payment_subject"`

	//measure                          string
	//country_of_origin_code           string
	//customs_declaration_number       string
	//excise                           string
	//product_code                     string
	//mark_mode                        string
	//additional_payment_subject_props string
	//agent_type                       string

	//mark_quantity                    obj
	//mark_code_info                   obj
	//supplier                         obj
	//payment_subject_industry_details []obj
}
