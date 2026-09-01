package sandbox

import (
	"strconv"
	"testing"
)

func BenchmarkCreatePaymentIntent(b *testing.B) {
	svc := NewService()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.CreatePaymentIntent(CreatePaymentIntentRequest{
			Amount:  100,
			Currency: "usd",
		}, "bench-create-"+strconv.Itoa(i), fingerprintString("bench-create-"+strconv.Itoa(i)))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPaymentLifecycle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc := NewService()
		key := "bench-lifecycle-" + strconv.Itoa(i)
		intent, err := svc.CreatePaymentIntent(CreatePaymentIntentRequest{Amount: 100, Currency: "usd", CaptureMethod: "manual"}, key, fingerprintString(key))
		if err != nil {
			b.Fatal(err)
		}
		confirmed, err := svc.ConfirmPaymentIntent(intent.ID, ConfirmPaymentIntentRequest{}, "", key+"-confirm", fingerprintString(key+"-confirm"))
		if err != nil {
			b.Fatal(err)
		}
		if _, err := svc.CapturePaymentIntent(confirmed.PaymentIntent.ID, CapturePaymentIntentRequest{}); err != nil {
			b.Fatal(err)
		}
	}
}
