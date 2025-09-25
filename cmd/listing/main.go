package main

import (
    "context"
    "encoding/json"
    "net/http"
    "github.com/go-resty/resty/v2"
    "github.com/gorilla/mux"
    "github.com/jackc/pgx/v5"
    "github.com/skip2/go-qrcode"
)

type InvoiceResponse struct {
    Bolt11    string `json:"bolt11"`
    QrCode    string `json:"qrCode"`
    InvoiceId string `json:"invoiceId"`
}

var db *pgx.Conn // PostgreSQL connection

func createInvoice(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    listingID := vars["id"]

    // Get user pubkey from JWT (via middleware)
    pubkey := r.Context().Value("pubkey").(string)

    // Validate listing
    var price int64
    err := db.QueryRow(context.Background(), "SELECT price FROM listings WHERE id = $1", listingID).Scan(&price)
    if err != nil {
        http.Error(w, "Listing not found", http.StatusNotFound)
        return
    }

    // Call LNBits
    client := resty.New()
    resp, err := client.R().
        SetHeader("X-Api-Key", "your-lnbits-api-key").
        SetBody(map[string]interface{}{
            "amount": price,
            "unit":   "sat",
            "memo":   "Listing " + listingID,
        }).
        Post("http://lnbits:5000/api/v1/payments")
    if err != nil {
        http.Error(w, "Failed to create invoice", http.StatusInternalServerError)
        return
    }

    var lnbitsResp struct {
        PaymentRequest string `json:"payment_request"`
        PaymentHash    string `json:"payment_hash"`
    }
    json.Unmarshal(resp.Body(), &lnbitsResp)

    qrCode, _ := qrcode.Encode(lnbitsResp.PaymentRequest, qrcode.Medium, 256)

    // Store invoice
    _, err = db.Exec(context.Background(), 
        "INSERT INTO invoices (id, listing_id, bolt11, status, pubkey) VALUES ($1, $2, $3, $4, $5)",
        lnbitsResp.PaymentHash, listingID, lnbitsResp.PaymentRequest, "pending", pubkey)
    if err != nil {
        http.Error(w, "Failed to save invoice", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(InvoiceResponse{
        Bolt11:    lnbitsResp.PaymentRequest,
        QrCode:    string(qrCode),
        InvoiceId: lnbitsResp.PaymentHash,
    })
}

func main() {
    // Initialize db (PostgreSQL)
    var err error
    db, err = pgx.Connect(context.Background(), "postgres://user:pass@postgres:5432/marketplace")
    if err != nil {
        log.Fatal(err)
    }

    router := mux.NewRouter()
    router.HandleFunc("/listings/{id}/invoice", middleware(createInvoice)).Methods("POST")
    log.Fatal(http.ListenAndServe(":8081", router))
}
