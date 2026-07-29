# Testing notes

Scratch list of test work still outstanding. Delete once it's done.

```bash
go vet ./... && gofmt -s -l . && go test -count=1 ./...
```

Run `go vet` separately — `go test` only runs a small subset of vet's checks, and
the missing ones are the ones that caught real bugs here (`unusedresult`).

Coverage snapshot as of the last session (`go test -coverprofile` +
`go tool cover -func`), total **53.7%**:

| function | coverage |
|---|---|
| `fundCode`, `newMux`, `Equal`, `Compare`, `MarshalJSON`, `ptr`, `getLatestPrices` | 100% |
| `latestPrices` | 90.9% |
| `parseCSV` | 85.2% |
| `writeJSON` | 80% |
| `fundPrices` | 78.6% |
| `UnmarshalJSON`, `parseDate`, `mustDate` | 75% |
| `getFunds`, `getFundPrices`, `writeError`, `allFunds`, `fundByCode` | **0%** |

---

## csv_test.go

### Fix

- **`canonical codes` subtest doesn't check codes are canonical.** It counts
  distinct codes but never inspects them, so 16 wrong codes pass exactly like 16
  right ones. Verified: disabling `fundCode` entirely (return the raw header)
  leaves the whole suite green with codes like `"C Fund"` and `"L Income"`.
  Fix by asserting each parsed code resolves in `seedFunds`:

  ```go
  if _, ok := fundByCode(p.FundCode); !ok {
      t.Errorf("parsed code %q has no fund metadata", p.FundCode)
  }
  ```

  Keep the count check alongside it. Together they mean "16 distinct, all real."
  This also becomes the alert that fires the first night TSP ships an L2080
  column, which is the behavior we chose deliberately.

### Add

- **Sparse row** — `2005-07-28` must yield exactly 5 records (G, F, C, S, I) and
  zero L funds. This is the empty-cell rule stated directly: absent means absent,
  not zero. Most valuable remaining test in the file.
- **Known value spot-check** — `C` on `2026-07-24` is `119.3426`. Catches column
  misalignment, which a count check cannot: the original off-by-one shifted data
  by one column while keeping code and price consistent with each other.
- **Trailing blank row is skipped** — no record should carry a zero `Date`.
- **Error cases**, inline via `strings.NewReader`, no fixture needed:
  - empty input → error, not panic (`"parseCSV: no records"`)
  - non-numeric price → error, not a zero-priced record
  - unparseable date → error
  - The point of each: a malformed value must be an error, while an *empty* one
    is skipped silently. That distinction is the whole design of the parser and
    nothing currently tests it.

---

## store_test.go

Currently only `TestFundPricesSortedNewestFirst`. `allFunds` and `fundByCode` are
at 0%.

- **`fundByCode`** — hit and miss. The miss case must check the bool, not just
  the returned `Fund`, since a miss returns `Fund{}, false` and the zero value
  is easy to mistake for a result.
- **`latestPrices`** — three properties, one is already covered by a handler test
  but not at the store level:
  - all returned rows share the newest date
  - rows are sorted by fund code
  - empty seed data returns an empty slice, not nil (JSON `[]` vs `null`)
- **`fundPrices` query filtering** — `From`/`To` are *inclusive*; a record exactly
  on the boundary must be included. Off-by-one on an inclusive bound is invisible
  without a test that puts a record on the edge.
- **`fundPrices` limit** — the limit applies *after* the sort, so it returns the
  N most recent, not N arbitrary rows. Test with a limit smaller than the result.
- **`fundPrices` unknown code** → empty slice, no panic.
- **Aliasing** — `allFunds()` returns a slice sharing its backing array with
  `seedFunds`, so a caller that sorts it permanently reorders the seed data.
  `fundPrices` by contrast builds a fresh slice and is safe to sort. Worth a test
  documenting which is which, because the difference is invisible at the call
  site and one of them is a landmine.

---

## models_test.go

Does not exist yet. `parseDate`, `mustDate`, and `UnmarshalJSON` are all at 75% —
the missing quarter is the error branch in each.

- **`Date.MarshalJSON`** — produces a bare `"2026-07-24"`, not an RFC 3339
  timestamp. This is the reason the custom type exists; nothing currently pins it.
- **`Date.UnmarshalJSON`** — round-trip a `SharePrice` through
  `json.Marshal`/`Unmarshal` and confirm the date survives. Also the failure
  case: malformed JSON string → error.
- **`parseDate` error path** — `"07/24/2026"` and `""` must both error.
- **`mustDate` panics on bad input** — use `defer func(){ recover() }()`. It's the
  only intentional panic in the codebase and its contract is "trusted literals
  only," so it should be pinned.
- **`SharePrice` JSON field names** — `fund`, `date`, `price`. These are the
  public API's wire format; renaming a Go field silently changes them.
- **`Fund` JSON** — `target_year` is *omitted* for core funds and L Income
  (`omitempty` + nil pointer), and present for dated L funds. The pointer-for-
  optionality decision is only observable here.

---

## main_test.go

Biggest gap in the project: **`getFundPrices` is at 0%** and it is the most
complex handler — path normalization, 404, three query parameters, and four
distinct 400 paths.

- **404** for an unknown fund code.
- **Lowercase path normalization** — `/api/v1/funds/c/prices` must behave
  identically to `/C/`.
- **Known fund, no prices** → `200` with `[]`, not `404` and not `null`.
- **400 paths**, each returning the `{"error": ...}` shape:
  - malformed `from`
  - malformed `to`
  - `from` after `to`
  - `limit` non-numeric or `< 1`
- **Limit defaulting and capping** — omitted limit uses `defaultPriceLimit`; a
  limit above `maxPriceLimit` is clamped rather than rejected.
- **`getFunds`** — 200, and the payload length matches `len(allFunds())`.
- **`writeError`** — response body is exactly `{"error": "..."}`. Every error path
  in the API depends on this shape.
- **`Content-Type: application/json`** on at least one success and one error
  response. Header-before-body ordering fails silently when it's wrong.

Skip `main` (0%, nothing to test) and `testParseCSV` (temporary, gets deleted once
`parseCSV` is wired into `store.go`).

---

## Recurring theme

Every bug found while writing these was **silent** — the code returned a
confident wrong answer rather than crashing:

- record count landed on exactly 68 while emitting 8 phantom records and dropping
  8 real ones
- `LIncome` broke 1 fund out of 16; the other 15 passed
- the `errors.New(...)` line reported nothing, so a missing fixture read as PASS
- swapped `Errorf` arguments printed *got* and *want* backwards

Two habits that follow: assert *shape*, not just size — and before trusting a new
test, break the thing it covers and confirm it goes red.
