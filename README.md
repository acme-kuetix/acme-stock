# acme-stock

Stock locations and moves for the Acme ERP: hierarchical warehouse locations, multi-line stock moves with a draft→confirmed→done state machine, and per-line product/UoM tracking.

## Models

- **Location** — a node in the stock location tree. Has a unique code, a type (`warehouse`/`internal`/`supplier`/`customer`/`production`/`transit`), an optional parent, and an active flag. Only `warehouse` and `internal` locations may have children; the others are leaves. A parent ID must reference an existing `warehouse` or `internal` location.
- **StockMove** — a document that moves stock from a source location to a destination location (which must differ and both exist). Starts in `draft`; transitions to `confirmed` (requires at least one line), then `done` (immutable thereafter). A draft or confirmed move may be `cancelled`; a `done` move cannot.
- **StockMoveLine** — a single product entry within a move. Carries a product ID, quantity (positive float), UoM ID, and optional note. Lines can be added or removed only while the move is in `draft` or `confirmed` state.

## Transitions

### Location
| Method | Args | Notes |
|---|---|---|
| `CreateLocation` | `code, name, typ, parentId string, active bool, address string` | 201; unique code; `typ` must be one of warehouse/internal/supplier/customer/production/transit; `parentId` if set must be warehouse/internal |
| `ListLocations` | `locationType string` | 200; optional filter by type |
| `GetLocation` | `locationId string` | 200/404 |
| `UpdateLocation` | `locationId, name, typ, parentId string, active bool, address string` | 200/404/400 — code is immutable; type change to a leaf fails if the location has children; parentId cycle and self-parent checks |
| `DeleteLocation` | `locationId string` | 200/404/400 — fails if it has children |
| `ListChildLocations` | `locationId string` | 200/404 — direct children only |

### StockMove
| Method | Args | Notes |
|---|---|---|
| `CreateStockMove` | `reference, sourceId, destinationId, scheduledAt, origin, note string` | 201; unique reference; source != destination; both must exist; `scheduledAt` if set must be RFC3339; starts in draft |
| `ListStockMoves` | `state, sourceId string` | 200; optional filters |
| `GetStockMove` | `moveId string` | 200/404; includes all lines |
| `AddStockMoveLine` | `moveId, productId string, qty interface{}, uomId, note string` | 201/400/404 — qty must be positive; move must be draft or confirmed |
| `DeleteStockMoveLine` | `moveId, lineId string` | 200/404/400 — move must be draft or confirmed |
| `ConfirmStockMove` | `moveId string` | 200/400/404 — draft→confirmed; requires ≥1 line |
| `DoneStockMove` | `moveId string` | 200/400/404 — confirmed→done; immutable thereafter |
| `CancelStockMove` | `moveId string` | 200/400/404 — draft/confirmed→cancelled; cannot cancel a done move |

## State machine

```
                  ┌─────────────┐
                  │    draft    │
                  └──────┬──────┘
            ┌────────────┼────────────┐
            │            │            │
            ▼            ▼            ▼
       (add line)   (confirm)    (cancel)
            │            │            │
            │            ▼            │
            │      ┌──────────┐      │
            │      │ confirmed│      │
            │      └────┬─────┘      │
            │           │            │
            │      (cancel)          │
            │           │            │
            │           ▼            ▼
            │      ┌──────────┐  ┌───────────┐
            └─────▶│   done   │  │ cancelled │
                   └──────────┘  └───────────┘
                   (immutable)
```

Once a move is `done`, `AddStockMoveLine`, `DeleteStockMoveLine`, and `CancelStockMove` all return 400 with `code: "move_locked"` (or `invalid_state` for cancel).

## WSL workflows

All 14 workflows live in `workflows/solutions/stock/`:

- `location-create`, `location-list`, `location-get`, `location-update`, `location-delete`, `location-children`
- `move-create`, `move-list`, `move-get`, `move-line-add`, `move-line-delete`, `move-confirm`, `move-done`, `move-cancel`

## HTTP endpoints (when composed into erp-app)

| Method | Path | Workflow |
|---|---|---|
| GET | `/stock/locations` | location-list |
| POST | `/stock/locations` | location-create |
| GET | `/stock/locations/{locationId}` | location-get |
| PUT | `/stock/locations/{locationId}` | location-update |
| DELETE | `/stock/locations/{locationId}` | location-delete |
| GET | `/stock/locations/{locationId}/children` | location-children |
| GET | `/stock/moves` | move-list |
| POST | `/stock/moves` | move-create |
| GET | `/stock/moves/{moveId}` | move-get |
| POST | `/stock/moves/{moveId}/lines` | move-line-add |
| DELETE | `/stock/moves/{moveId}/lines/{lineId}` | move-line-delete |
| POST | `/stock/moves/{moveId}/confirm` | move-confirm |
| POST | `/stock/moves/{moveId}/done` | move-done |
| POST | `/stock/moves/{moveId}/cancel` | move-cancel |

## Engine gotchas encountered

- The Go param for the location type field is `typ` not `type` — `type` is a Go keyword. WSL passes `typ: $json.type` and the response map keeps the user-facing `"type"` JSON key.
- The Go param for the location ID is `locationId` not `id` — `id` is consumed by `FlowTransition.ID` during WSL decoding. Same convention used for `moveId`, `lineId`.
- `qty` is `interface{}` not `float64` — JSON numbers arrive as float64 and the engine's `castInput` doesn't convert float64→int/float64 for typed params. The transition uses a `toFloat()` helper.
- All transitions live in a single `stock.go` file because `kue update` generates meta/di from the file containing the `NewStockTransitions` constructor. Splitting into separate files would cause those methods to be silently dropped from the metadata cache.
- `ResetStore` clears both location and move stores (and their ID counters) — tests rely on this for isolation.

## Interchangeability

The transition file uses an in-memory store today (map + RWMutex). The method signatures are stable, so a SQL/Postgres variant can replace the storage layer without touching the WSL workflows that call `stock/stock.*`.
