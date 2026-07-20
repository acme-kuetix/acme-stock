package transitions

import (
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	persistenceStore "github.com/acme-kuetix/acme-std-persistence/modules/persistence/store/transitions"
)

// testIDCounter provides unique IDs for tests that don't go through
// sequence/sequence.NextByCode (i.e. seedLocation/seedStockMove, which
// mirror WSL but stay inside this package's Go test harness).
var testIDCounter int64

func nextTestID() string {
	return strconv.FormatInt(atomic.AddInt64(&testIDCounter, 1), 10)
}

func newStock() *stockTransitions { return &stockTransitions{} }

// resetAll clears both the stock-local state and the persistence store.
// Every test calls this to start from a clean slate.
func resetAll() {
	persistenceStore.ResetStore()
	atomic.StoreInt64(&testIDCounter, 0)
}

// ─── Seed helpers (mirror WSL workflow steps) ────────────────
//
// seedLocation mirrors workflows/solutions/stock/location-create.wsl:
// sequence/sequence.NextByCode(code: "stock-location") → persistence/store.Create("stock_locations", ...).
// Returns the new location id. Uses nextTestID() instead of
// sequence/sequence.NextByCode so this test stays inside the acme-stock
// package without importing acme-sequence (which isn't in go.mod).
// Caller is responsible for calling resetAll() before seeding.
func seedLocation(t *testing.T, code, name, typ, parentId string) string {
	t.Helper()
	id := nextTestID()
	now := time.Now().UTC()
	doc := map[string]interface{}{
		"id":        id,
		"code":      code,
		"name":      name,
		"type":      typ,
		"parentId":  parentId,
		"active":    true,
		"address":   "",
		"createdAt": now,
		"updatedAt": now,
	}
	store := persistenceStore.NewStoreTransitionsConcrete()
	if r := store.Create(locationsCollection, id, doc); !r.Success {
		t.Fatalf("seedLocation store.Create: %v", r.Error)
	}
	return id
}

// seedStockMove mirrors workflows/solutions/stock/move-create.wsl:
// sequence/sequence.NextByCode(code: "stock-move") → persistence/store.Create("stock_moves", ...).
// Returns the new move id. Uses nextTestID() for the same reason as seedLocation.
// The reference is set to a placeholder "SM-<id>" — production WSL uses
// sequence/sequence.NextByCode (prefix "SM-", padding 4).
// Caller is responsible for calling resetAll() before seeding.
func seedStockMove(t *testing.T, sourceId, destinationId string) string {
	t.Helper()
	id := nextTestID()
	reference := "SM-" + id
	now := time.Now().UTC()
	doc := map[string]interface{}{
		"id":            id,
		"reference":     reference,
		"sourceId":      sourceId,
		"destinationId": destinationId,
		"state":         moveStateDraft,
		"scheduledAt":   "",
		"origin":        "",
		"note":          "",
		"lines":         []interface{}{},
		"createdAt":     now,
		"updatedAt":     now,
	}
	store := persistenceStore.NewStoreTransitionsConcrete()
	if r := store.Create(movesCollection, id, doc); !r.Success {
		t.Fatalf("seedStockMove store.Create: %v", r.Error)
	}
	return id
}

// getLocationDoc fetches the raw location doc from persistence/store.
func getLocationDoc(t *testing.T, id string) map[string]interface{} {
	t.Helper()
	store := persistenceStore.NewStoreTransitionsConcrete()
	gr := store.Get(locationsCollection, id)
	if !gr.Success {
		t.Fatalf("getLocationDoc: Get failed: %v", gr.Error)
	}
	return gr.Response.(map[string]interface{})["doc"].(map[string]interface{})
}

// getMoveDoc fetches the raw move doc from persistence/store.
func getMoveDoc(t *testing.T, id string) map[string]interface{} {
	t.Helper()
	store := persistenceStore.NewStoreTransitionsConcrete()
	gr := store.Get(movesCollection, id)
	if !gr.Success {
		t.Fatalf("getMoveDoc: Get failed: %v", gr.Error)
	}
	return gr.Response.(map[string]interface{})["doc"].(map[string]interface{})
}

// ─── GetLocationTree (migrated to WSL via collections/ops/ops.RecurseForest) ───

func TestGetLocationTree(t *testing.T) {
	t.Skip("GetLocationTree migrated to WSL — tree now built by collections/ops/ops.RecurseForest")
}

func TestGetLocationTreeSubtree(t *testing.T) {
	t.Skip("GetLocationTree migrated to WSL — tree now built by collections/ops/ops.RecurseForest")
}

