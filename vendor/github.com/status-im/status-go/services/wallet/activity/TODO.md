# Provide dynamic activity updates

Task: https://github.com/status-im/status-desktop/issues/12120

## Intro

In the current approach only static paginated filtering is possible because the filtering is done in SQL

The updated requirements need to support dynamic updates of the current visualized filter

## Plan

- [ ] Required common (runtime/SQL) infrastructure
  - [-] Refactor into a session based filter
  - [-] Keep a mirror of identities for session
  - [-] Capture events (new downloaded and pending first)
  - [-] Have the simplest filter to handle new and updated and emit wallet event
  - [ ] Handle update filter events in UX and alter the model (add/remove)
- [ ] Asses how the runtime filter grows in complexity/risk
- [ ] Quick prototype of SQL only filter if still make sense
- [ ] Refactor the async handling to fit the session based better (use channels and goroutine)

## How to

I see two ways:

- Keep a **runtime** (go/nim) dynamic in memory filter that is in sync with the SQL filter and use the filter to process transactions updates and propagate to the current visualized model
  - The filter will push changes to the in memory model based on the sorting and filtering criteria
  - If the filter is completely in sync withe the SQL one, then the dynamic updates to the model should have the same content as fetched from scratch from the DB
  - *Advantages*
    - Less memory and performance requirements
  - *Disadvantages*
    - Two sources of truth for the filter
      - With tests for each event this can be mitigated
    - Complexity around the multi-transaction/sub-transaction relation
    - If we miss doing equivalent changes in bot filters (SQL and runtime) the filter might not be in sync with the SQL one and have errors in update
- **Refresh SQL filter** on every transaction (or bulk) update to DB and compare with the current visualized filter to extract differences and push as change notifications
  - This approach is more expensive in terms of memory and performance but will use only one source of truth implementation
  - This way we know for sure that the updated model is in sync with a newly fetched one
  - *Advantages*
    - Less complexity and less risk to be out of sync with the SQL filter
  - *Disadvantages*
    - More memory and performance requirements
      - The real improvement will be to do the postponed refactoring of the activity in DB

## Requirements

Expected filter states to be addressed

- Filter is set
- No Filter
- Filter is cleared
  - How about if only partially cleared?

Expected dynamic events

- **New transactions**
  - Pending
  - Downloaded (external)
  - Multi-transactions?
- **Transaction changed state**
  - Pending to confirmed (new transaction/removed transaction)

Filter criteria

- time interval: start-end
- activity type (send/receive/buy/swap/bridge/contract_deploy/mint)
- status (pending/failed/confirmed/finalized)
- addresses
- tokens
- multi-transaction filtering transaction

## Implementation

### SQL filter

For new events

- keep a mirror of identities on status-go side (optional session based)
- on update events fetch identities and check against the mirror if any is new
- for new entries send the notification with the transaction details
- keep pending changes (not added)
  - remove entries that were processed for this session

For update?

- check if entry is in the mirror and propagate update event

### Mirror filter

For new events

- keep a mirror of identities
- on update events pass them through the filter and if they pass send updates
  - the filter checks criteria and available mirror interval to dismiss from mirror
- sub-transactions challenge
  - TODO
- token challenges
  - TODO

For update?

- check if entry is in the mirror and propagate update event