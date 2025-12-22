# Query Examples

This document provides examples for all 13 query types that the codebase supports.

## 1. Get All Notes for an Individual (Including Inline)

```go
query, _ := query.NewQuery(tree)
indiQuery := query.Individual("@I1@")
notes, err := indiQuery.GetAllNotes()
if err != nil {
    log.Fatal(err)
}

for _, note := range notes {
    if note.IsInline {
        fmt.Printf("Inline note: %s\n", note.Text)
    } else {
        fmt.Printf("Referenced note [%s]: %s\n", note.XrefID, note.Text)
    }
}
```

## 2. Get All Notes for a Family (Including Inline)

```go
query, _ := query.NewQuery(tree)
famQuery := query.Family("@F1@")
notes, err := famQuery.GetAllNotes()
if err != nil {
    log.Fatal(err)
}

for _, note := range notes {
    fmt.Printf("Note: %s (inline: %v)\n", note.Text, note.IsInline)
}
```

## 3. Get All Records Associated with a Note

```go
graph := query.Graph()
records, err := graph.GetRecordsForNote("@N1@")
if err != nil {
    log.Fatal(err)
}

for _, record := range records {
    fmt.Printf("Record %s of type %s references note @N1@\n", 
        record.ID(), record.NodeType())
}
```

## 4. Get All Records Associated with an Event

```go
graph := query.Graph()
records, err := graph.GetRecordsForEvent("@I1@_BIRT_0")
if err != nil {
    log.Fatal(err)
}

for _, record := range records {
    fmt.Printf("Record %s has event @I1@_BIRT_0\n", record.ID())
}
```

## 5. Get All Events on a Particular Day

```go
graph := query.Graph()
// Get all events on January 15, 1900
events, err := graph.GetEventsOnDate(1900, 1, 15)
if err != nil {
    log.Fatal(err)
}

for _, event := range events {
    fmt.Printf("Event: %s on %s\n", event.EventType, event.Date)
}
```

## 6. Get Events of Specific Type on a Particular Day

```go
graph := query.Graph()
// Get all birthdays on January 15
birthdays, err := graph.GetEventsOnDateByType("BIRT", 0, 1, 15)
if err != nil {
    log.Fatal(err)
}

for _, birthday := range birthdays {
    fmt.Printf("Birthday: %s\n", birthday.Date)
}
```

## 7. Get All Individuals with Last Name X

```go
query, _ := query.NewQuery(tree)
results, err := query.Filter().BySurname("Smith").Execute()
if err != nil {
    log.Fatal(err)
}

for _, indi := range results {
    fmt.Printf("Found: %s\n", indi.GetName())
}
```

## 8. Get All Individuals with First Name X

```go
query, _ := query.NewQuery(tree)
results, err := query.Filter().ByGivenName("John").Execute()
if err != nil {
    log.Fatal(err)
}

for _, indi := range results {
    fmt.Printf("Found: %s\n", indi.GetName())
}
```

## 9. Get Most Common First Name

```go
graph := query.Graph()
commonNames := graph.GetMostCommonGivenNames(10)

for _, nameCount := range commonNames {
    fmt.Printf("%s: %d occurrences\n", nameCount.Name, nameCount.Count)
}
```

## 10. Get Most Common Last Name

```go
graph := query.Graph()
commonSurnames := graph.GetMostCommonSurnames(10)

for _, nameCount := range commonSurnames {
    fmt.Printf("%s: %d occurrences\n", nameCount.Name, nameCount.Count)
}
```

## 11. Get All Events for an Individual

```go
query, _ := query.NewQuery(tree)
indiQuery := query.Individual("@I1@")
events, err := indiQuery.GetEvents()
if err != nil {
    log.Fatal(err)
}

for _, event := range events {
    fmt.Printf("Event: %s on %s at %s\n", 
        event.EventType, event.Date, event.Place)
}
```

## 12. Get All Events for a Family

```go
query, _ := query.NewQuery(tree)
famQuery := query.Family("@F1@")
events, err := famQuery.GetEvents()
if err != nil {
    log.Fatal(err)
}

for _, event := range events {
    fmt.Printf("Family event: %s on %s\n", event.EventType, event.Date)
}
```

## 13. Get Individuals with Birthday in a Particular Month/Day/Year/Range

```go
query, _ := query.NewQuery(tree)

// By month only (any year, any day)
results, _ := query.Filter().ByBirthMonth(1).Execute() // January

// By day only (any month, any year)
results, _ := query.Filter().ByBirthDay(15).Execute() // 15th of any month

// By month and day (any year)
results, _ := query.Filter().ByBirthMonthAndDay(1, 15).Execute() // January 15

// By year (already exists)
results, _ := query.Filter().ByBirthYear(1900).Execute()

// By date range (already exists)
start := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
end := time.Date(1950, 12, 31, 23, 59, 59, 999999999, time.UTC)
results, _ := query.Filter().ByBirthDateRange(start, end).Execute()
```

## Additional Filter Examples

### Exact Name Matching

```go
// Exact surname match
results, _ := query.Filter().BySurnameExact("Smith").Execute()

// Exact given name match
results, _ := query.Filter().ByGivenNameExact("John").Execute()
```

### Combined Filters

```go
// Find all individuals named "John Smith" born in 1900
results, _ := query.Filter().
    ByGivenNameExact("John").
    BySurnameExact("Smith").
    ByBirthYear(1900).
    Execute()
```

## Notes

- All query methods are thread-safe and can be used concurrently
- Results are cached for repeated queries on the same graph
- Inline notes are notes embedded directly in the record (not referenced via xref)
- Event dates support various GEDCOM date formats (exact, about, before, after, between, etc.)
- Name filters are case-insensitive by default
- Analytics methods (most common names) return results sorted by count (descending)

