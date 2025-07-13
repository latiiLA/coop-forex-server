package migration

import (
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// unseedCurrencies hard-deletes the seeded currencies from the currencies collection.
func unseedCurrencies(ctx context.Context, db *mongo.Database) error {
	// List of short_codes to unseed
	shortCodes := []string{
		"USD", "CAD", "EUR", "AED", "AFN", "ALL", "AMD", "ARS", "AUD", "AZN",
		"BAM", "BDT", "BGN", "BHD", "BIF", "BND", "BOB", "BRL", "BWP", "BYN",
		"BZD", "CDF", "CHF", "CLP", "CNY", "COP", "CRC", "CVE", "CZK", "DJF",
		"DKK", "DOP", "DZD", "EEK", "EGP", "ERN", "ETB", "GBP", "GEL", "GHS",
		"GNF", "GTQ", "HKD", "HNL", "HRK", "HUF", "IDR", "ILS", "INR", "IQD",
		"IRR", "ISK", "JMD", "JOD", "JPY", "KES", "KHR", "KMF", "KRW", "KWD",
		"KZT", "LBP", "LKR", "LTL", "LVL", "LYD", "MAD", "MDL", "MGA", "MKD",
		"MMK", "MOP", "MUR", "MXN", "MYR", "MZN", "NAD", "NGN", "NIO", "NOK",
		"NPR", "NZD", "OMR", "PAB", "PEN", "PHP", "PKR", "PLN", "PYG", "QAR",
		"RON", "RSD", "RUB", "RWF", "SAR", "SDG", "SEK", "SGD", "SOS", "SYP",
		"THB", "TND", "TOP", "TRY", "TTD", "TWD", "TZS", "UAH", "UGX", "UYU",
		"UZS", "VEF", "VND", "XAF", "XOF", "YER", "ZAR", "ZMK", "ZWL",
	}

	collection := db.Collection("currencies")

	// Hard delete: remove documents with matching short_codes
	result, err := collection.DeleteMany(
		ctx,
		bson.M{"short_code": bson.M{"$in": shortCodes}},
	)
	if err != nil {
		return fmt.Errorf("failed to hard-delete currencies: %w", err)
	}

	log.Printf("✅ Hard-deleted %d currencies", result.DeletedCount)
	return nil
}
