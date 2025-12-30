package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type SubscriptionStatus string

const (
	SubscriptionStatusActive     SubscriptionStatus = "active"
	SubscriptionStatusTrialing   SubscriptionStatus = "trialing"
	SubscriptionStatusCanceled   SubscriptionStatus = "canceled"
	SubscriptionStatusPastDue    SubscriptionStatus = "past_due"
	SubscriptionStatusUnpaid     SubscriptionStatus = "unpaid"
	SubscriptionStatusIncomplete SubscriptionStatus = "incomplete"
)

type BillingCycle string

const (
	BillingCycleMonthly BillingCycle = "monthly"
	BillingCycleYearly  BillingCycle = "yearly"
)

type Subscription struct {
	Base            `json:",inline" bson:",inline"`
	OrganizationID  bson.ObjectID      `json:"organization_id" bson:"organization_id"`
	StripeSubID     string             `json:"stripe_sub_id" bson:"stripe_sub_id"`
	StartDate       time.Time          `json:"started_at" bson:"started_at"`
	EndDate         time.Time          `json:"ended_at" bson:"ended_at"`
	TrialPeriodDays int                `json:"trial_period_days" bson:"trial_period_days"`
	BillingCycle    BillingCycle       `json:"billing_cycle" bson:"billing_cycle"` // e.g., "monthly", "yearly"
	Status          SubscriptionStatus `json:"status" bson:"status"`               // e.g., "active", "trialing", "canceled", "past_due", "unpaid", "canceled", "incomplete"
	IsCurrent       bool               `json:"is_current" bson:"is_current"`
}

type CreateSubscription struct {
	OrganizationID  bson.ObjectID `json:"organization_id" bson:"organization_id"`
	StripeSubID     string        `json:"stripe_sub_id" bson:"stripe_sub_id"`
	StartDate       time.Time     `json:"started_at" bson:"started_at"`
	EndDate         time.Time     `json:"ended_at" bson:"ended_at"`
	TrialPeriodDays int           `json:"trial_period_days" bson:"trial_period_days"`
	BillingCycle    BillingCycle  `json:"billing_cycle" bson:"billing_cycle"` // e.g., "monthly", "yearly"

}

type UpdateSubscription struct {
	BillingPlanID bson.ObjectID      `json:"billing_plan_id" bson:"billing_plan_id"`
	EndDate       time.Time          `json:"ended_at" bson:"ended_at"`
	Status        SubscriptionStatus `json:"status" bson:"status"`
}

type GetFreeTrialInfo struct {
	RecordCount    int `json:"record_count" bson:"record_count"`
	DaysLeft       int `json:"days_left" bson:"days_left"`
	RemainingScans int `json:"remaining_scans" bson:"remaining_scans"`
}
