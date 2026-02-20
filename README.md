# Event Registration & Ticketing System

## Overview
This is a comprehensive REST API built with Go (Golang) and Gin, acting as an Event Registration and Ticketing System (similar to Eventbrite).

### Core Deliverables Achieved
1. **Complete REST API**: Users can browse events, register for available slots, and organizers can manage events. Custom login authorization correctly divides these roles without cluttering the UI.
2. **Database Schema with Proper Constraints**: The project uses PostgreSQL with GORM models. Correct foreign-key mappings (User -> Event Registrations) and capacity decrement validations exist.
3. **Concurrent Booking Simulation**: Included in `backend/cmd/simulate_concurrency.go` which simulates multiple goroutines trying to claim one single event spot simultaneously.
4. **Preventing Race Conditions Documentation**: Described below.

---

## Concurrency Strategy: Preventing Overbooking (Race Conditions)

In a high-demand ticketing system, multiple users could try to book the exact same final ticket slot within the exact same millisecond. If not handled correctly, the backend could read `AvailableCapacity = 1` for 10 simultaneous requests and allow all 10 users to register, resulting in a severe overbooking condition.

### How We Solved It
Our solution uses a strict **Transaction processing with Database Row-Level Locking**.

#### 1. The Database Transaction Layer (`tx`)
Whenever a user attempts to register, we wrap the operation in a Gorm database transaction. This ensures atomicity (either both the capacity decrement and the registration happen, or neither happens).

```go
err := database.DB.Transaction(func(tx *gorm.DB) error {
    // ... execution
})
```

#### 2. Row-Level Pessimistic Locking (`FOR UPDATE`)
To prevent race conditions, the transaction must acquire an exclusive lock on the *specific Event Row* before checking its capacity. We explicitly trigger PostgreSQL's `SELECT ... FOR UPDATE` clause via Gorm:

```go
if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&event, input.EventID).Error; err != nil {
    return err
}
```
**Why this works:**
If 5 requests hit the backend at identical times, the Postgres engine itself intercepts them. The first request grabs the `"UPDATE"` lock on the queried event. 
The other 4 requests **hang and wait** at the query level until Request 1 finishes and releases the lock. 
Request 1 successfully sees `Capacity = 1`, decrements it to `Capacity = 0`, writes the user's registration ticket, and commits/releases the lock.
Now, Requests 2, 3, 4, and 5 acquire their locks sequentially. However, when they read the database, they will read the **updated** capacity of `0`. The backend safely rejects them with the `no_slots` error and stops them from writing tickets.

#### 3. Duplicate Prevention Constraint
Within the same transaction, a query asserts that the `user_id + event_id` does not already exist within the `Registration` table. If it does, it kicks back an `"already_registered"` constraint error, acting as a secondary idempotency guard.

---

## How to Test Concurrency Validation

I have built a simulated test environment that blasts the main REST API with parallel goroutines locking onto one single available capacity ticket.

To run the concurrent simulator:
1. Ensure your PostgreSQL container and backend Go server are actively running (`go run ./cmd/main.go`).
2. Open a new terminal.
3. Execute the simulation tool:
   ```bash
   cd backend
   go run ./cmd/simulate_concurrency.go
   ```

You will notice an output map explicitly showing that out of all hitting concurrent processes, exactly 1 will intercept the lock cleanly yielding a `SUCCESS`, and the remaining parallel processes will catch the `no_slots` rejection!
