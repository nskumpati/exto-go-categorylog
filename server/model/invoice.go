package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Invoice struct {
	Base               `json:",inline" bson:",inline"`
	OrganizationID     bson.ObjectID `json:"organization_id" bson:"organization_id"`
	SubscriptionID     bson.ObjectID `json:"subscription_id" bson:"subscription_id"`
	StripeInvoiceID    string        `json:"stripe_invoice_id" bson:"stripe_invoice_id"`
	InvoiceNumber      string        `json:"invoice_number" bson:"invoice_number"`
	TotalAmount        float64       `json:"total_amount" bson:"total_amount"`
	BillingPeriodStart time.Time     `json:"billing_period_start" bson:"billing_period_start"`
	BillingPeriodEnd   time.Time     `json:"billing_period_end" bson:"billing_period_end"`
	PDF_URL            string        `json:"pdf_url" bson:"pdf_url"`
	Status             string        `json:"status" bson:"status"` // e.g., "paid", "unpaid", "void"
}

type InvoiceStatus string

const (
	InvoicePaid   InvoiceStatus = "paid"
	InvoiceUnpaid InvoiceStatus = "unpaid"
	InvoiceVoid   InvoiceStatus = "void"
)
