package wheretopark_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	wheretopark "wheretopark/go"
)

func TestServerClient(t *testing.T) {
	var provider1URL, provider2URL string
	mux := http.NewServeMux()

	mux.HandleFunc("/v1/config/providers", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		providers := []map[string]string{
			{"name": "provider1", "url": provider1URL},
			{"name": "provider2", "url": provider2URL},
		}
		json.NewEncoder(w).Encode(providers)
	})

	mux.HandleFunc("/v1/provider/provider1/parking-lots", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lots := map[string]wheretopark.ParkingLot{"A": sampleParkingLot}
		json.NewEncoder(w).Encode(lots)
	})

	mux.HandleFunc("/v1/provider/provider2/parking-lots", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		lots := map[string]wheretopark.ParkingLot{"B": sampleParkingLot}
		json.NewEncoder(w).Encode(lots)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	provider1URL = server.URL + "/v1/provider/provider1"
	provider2URL = server.URL + "/v1/provider/provider2"

	base, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	c := wheretopark.NewServerClient(base)

	providers, err := c.Providers()
	if err != nil {
		t.Fatal(err)
	}
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}

	allParkingLots := make(map[wheretopark.ID]wheretopark.ParkingLot)
	for _, provider := range providers {
		parkingLots, err := c.GetFrom(provider)
		if err != nil {
			t.Fatalf("failed to fetch from %s: %s", provider.Name, err)
		}
		for id, parkingLot := range parkingLots {
			allParkingLots[id] = parkingLot
		}
	}

	allParkingLots2, err := c.GetFromMany(providers)
	if err != nil {
		t.Fatal(err)
	}

	equalJson[map[wheretopark.ID]wheretopark.ParkingLot](t, allParkingLots, allParkingLots2, "different parking lot maps returned from GetFrom and GetFromMany")
}
