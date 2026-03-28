package app

import (
	"testing"

	"accounting-agent/internal/core"
)

func TestDetectPostingIntent(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected PostingIntent
	}{
		{name: "sales invoice", key: "invoice-order-101", expected: PostingIntentSalesInvoice},
		{name: "customer receipt", key: "payment-order-101", expected: PostingIntentCustomerReceipt},
		{name: "vendor payment", key: "pay-vendor-po-101", expected: PostingIntentVendorPayment},
		{name: "goods receipt movement", key: "goods-receipt-mv-77", expected: PostingIntentGoodsReceipt},
		{name: "goods receipt service", key: "po-10-line-2-service-receipt", expected: PostingIntentGoodsReceipt},
		{name: "goods issue", key: "goods-issue-order-9", expected: PostingIntentGoodsIssue},
		{name: "manual", key: "manual-year-end-1", expected: PostingIntentManualAdjustment},
		{name: "unknown", key: "8f68a48d-b2d6-4f19-95e9-0d9a5c2f7e2d", expected: PostingIntentUnknown},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			intent := detectPostingIntent(core.Proposal{IdempotencyKey: tc.key})
			if intent != tc.expected {
				t.Fatalf("expected %s, got %s", tc.expected, intent)
			}
		})
	}
}

func TestValidateDocumentTypeForIntent(t *testing.T) {
	tests := []struct {
		name        string
		intent      PostingIntent
		docTypeCode string
		wantErr     bool
	}{
		{name: "customer receipt valid", intent: PostingIntentCustomerReceipt, docTypeCode: "RC", wantErr: false},
		{name: "customer receipt invalid JE", intent: PostingIntentCustomerReceipt, docTypeCode: "JE", wantErr: true},
		{name: "vendor payment valid", intent: PostingIntentVendorPayment, docTypeCode: "PV", wantErr: false},
		{name: "unknown intent pass-through non-je", intent: PostingIntentUnknown, docTypeCode: "RC", wantErr: false},
		{name: "unknown intent je blocked", intent: PostingIntentUnknown, docTypeCode: "JE", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateDocumentTypeForIntent(tc.intent, tc.docTypeCode)
			if err == nil {
				err = validateUnknownIntentDocumentTypePolicy(tc.intent, tc.docTypeCode)
			}
			if tc.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}
		})
	}
}
