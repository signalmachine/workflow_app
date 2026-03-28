package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"accounting-agent/internal/app"
	"accounting-agent/internal/core"

	"github.com/shopspring/decimal"
)

// Run executes a one-shot CLI command and exits.
// args is os.Args[1:] — the first element is the subcommand name.
func Run(ctx context.Context, svc app.ApplicationService, args []string) {
	company, err := svc.LoadDefaultCompany(ctx)
	if err != nil {
		log.Fatalf("Failed to load company: %v", err)
	}

	switch args[0] {
	case "propose", "prop", "p":
		if len(args) < 2 {
			log.Fatal("Usage: app propose \"<event description>\"")
		}
		event := args[1]
		result, err := svc.InterpretEvent(ctx, event, company.CompanyCode)
		if err != nil {
			log.Fatalf("Agent error: %v", err)
		}
		if result.IsClarification {
			fmt.Fprintln(os.Stderr, "AI needs clarification:", result.ClarificationMessage)
			os.Exit(1)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(result.Proposal)

	case "validate", "val", "v":
		var proposal core.Proposal
		if err := json.NewDecoder(os.Stdin).Decode(&proposal); err != nil {
			log.Fatalf("Invalid JSON: %v", err)
		}
		validateCtx := app.WithProposalSource(ctx, app.ProposalSourceCLI)
		if err := svc.ValidateProposal(validateCtx, proposal); err != nil {
			log.Fatalf("Validation failed: %v", err)
		}
		fmt.Println("Proposal is valid.")

	case "commit", "com", "c":
		var proposal core.Proposal
		if err := json.NewDecoder(os.Stdin).Decode(&proposal); err != nil {
			log.Fatalf("Invalid JSON: %v", err)
		}
		commitCtx := app.WithProposalSource(ctx, app.ProposalSourceCLI)
		if err := svc.CommitProposal(commitCtx, proposal); err != nil {
			log.Fatalf("Commit failed: %v", err)
		}
		fmt.Println("Transaction Committed.")

	case "bal", "balances":
		result, err := svc.GetTrialBalance(ctx, company.CompanyCode)
		if err != nil {
			log.Fatalf("Failed to get balances: %v", err)
		}
		printTrialBalance(result)

	case "vendor-invoice-record":
		if len(args) < 7 {
			log.Fatal("Usage: app vendor-invoice-record <vendor_id> <invoice_number> <invoice_date> <invoice_amount> <expense_account_code> <line_amount> [line_description]")
		}
		vendorID, err := strconv.Atoi(args[1])
		if err != nil || vendorID <= 0 {
			log.Fatalf("invalid vendor_id: %s", args[1])
		}
		invoiceDate, err := time.Parse("2006-01-02", args[3])
		if err != nil {
			log.Fatalf("invalid invoice_date: %v", err)
		}
		invoiceAmount, err := decimal.NewFromString(args[4])
		if err != nil {
			log.Fatalf("invalid invoice_amount: %v", err)
		}
		lineAmount, err := decimal.NewFromString(args[6])
		if err != nil {
			log.Fatalf("invalid line_amount: %v", err)
		}
		lineDescription := ""
		if len(args) > 7 {
			lineDescription = args[7]
		}
		result, err := svc.RecordDirectVendorInvoice(ctx, app.DirectVendorInvoiceRequest{
			CompanyCode:   company.CompanyCode,
			VendorID:      vendorID,
			InvoiceNumber: args[2],
			InvoiceDate:   invoiceDate,
			PostingDate:   invoiceDate,
			DocumentDate:  invoiceDate,
			InvoiceAmount: invoiceAmount,
			Lines: []app.DirectVendorInvoiceLineInput{
				{
					Description:        lineDescription,
					ExpenseAccountCode: args[5],
					Amount:             lineAmount,
				},
			},
		})
		if err != nil {
			log.Fatalf("vendor-invoice-record failed: %v", err)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result.VendorInvoice)

	case "vendor-invoice-pay":
		if len(args) < 5 {
			log.Fatal("Usage: app vendor-invoice-pay <vendor_invoice_id> <bank_account_code> <amount> <payment_date>")
		}
		vendorInvoiceID, err := strconv.Atoi(args[1])
		if err != nil || vendorInvoiceID <= 0 {
			log.Fatalf("invalid vendor_invoice_id: %s", args[1])
		}
		amount, err := decimal.NewFromString(args[3])
		if err != nil {
			log.Fatalf("invalid amount: %v", err)
		}
		paymentDate, err := time.Parse("2006-01-02", args[4])
		if err != nil {
			log.Fatalf("invalid payment_date: %v", err)
		}
		result, err := svc.PayVendorInvoice(ctx, app.PayVendorInvoiceRequest{
			CompanyCode:     company.CompanyCode,
			VendorInvoiceID: vendorInvoiceID,
			BankAccountCode: args[2],
			Amount:          amount,
			PaymentDate:     paymentDate,
		})
		if err != nil {
			log.Fatalf("vendor-invoice-pay failed: %v", err)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result.VendorInvoice)

	case "po-close":
		if len(args) < 3 {
			log.Fatal("Usage: app po-close <po_id> <close_reason>")
		}
		poID, err := strconv.Atoi(args[1])
		if err != nil || poID <= 0 {
			log.Fatalf("invalid po_id: %s", args[1])
		}
		result, err := svc.ClosePurchaseOrder(ctx, app.ClosePurchaseOrderRequest{
			CompanyCode: company.CompanyCode,
			POID:        poID,
			CloseReason: args[2],
		})
		if err != nil {
			log.Fatalf("po-close failed: %v", err)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result.PurchaseOrder)

	default:
		log.Fatalf("Unknown command: %s\nAvailable: propose, validate, commit, bal, vendor-invoice-record, vendor-invoice-pay, po-close", args[0])
	}
}

func printTrialBalance(result *app.TrialBalanceResult) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 62))
	fmt.Printf("  %-58s\n", "TRIAL BALANCE")
	fmt.Printf("  Company  : %s — %s\n", result.CompanyCode, result.CompanyName)
	fmt.Printf("  Currency : %s\n", result.Currency)
	fmt.Println(strings.Repeat("=", 62))
	fmt.Printf("  %-10s %-30s %15s\n", "CODE", "NAME", "BALANCE")
	fmt.Println(strings.Repeat("-", 62))
	for _, b := range result.Accounts {
		fmt.Printf("  %-10s %-30s %15s\n", b.Code, b.Name, b.Balance.StringFixed(2))
	}
	fmt.Println(strings.Repeat("=", 62))
}
