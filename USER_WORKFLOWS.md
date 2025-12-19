# User Workflows Guide

This guide is designed for genealogists and researchers using the GEDCOM Go tool. It focuses on practical workflows for two main audiences:

1. **Private Family Researchers** (10K–200K individuals)
2. **Community/Population Researchers** (500K–5M individuals)

## Philosophy: "Genealogy Workshop, Not a Compiler"

This tool is designed to be usable by genealogists who want to:
- Find everyone named X around place Y between years A–B
- See how two people might be related
- Identify what's missing or suspicious
- Spot duplicates and contradictions
- Export data for sharing and visualization

You don't need to understand algorithms or technical details. The tool provides:
- **Guided defaults** for complex features
- **Clear explanations** instead of silent failures
- **Helpful warnings** when data has issues
- **Scoped operations** that focus on what you need

---

## For Private Family Researchers

### Your Typical Dataset
- 10,000–200,000 individuals
- Your own family tree or a few related families
- Mix of complete and incomplete records
- Goal: Find duplicates, explore relationships, validate data

### Finding Duplicates

**Basic duplicate detection:**
```bash
gedcom duplicates family.ged --top 200
```

This will:
- Find the top 200 most likely duplicates
- Rank them by similarity score
- Show why each pair matches (same parents, close birth year, similar place, etc.)
- Run in minutes, not hours

**Scoped duplicate detection:**
```bash
# Find duplicates in a specific time period and place
gedcom duplicates family.ged --place "Guyana" --year 1850-1920 --top 200

# Find duplicates for a specific surname
gedcom duplicates family.ged --surname "Smith" --top 100

# Get detailed explanations for each match
gedcom duplicates family.ged --top 50 --explain
```

**What you'll see:**
- Ranked list of potential duplicates
- For each match: similarity score, matched fields, differences
- Warnings if your data has very common surnames (e.g., "Smith" appears 5,000 times)
- Suggestions for how to improve results (add place filter, widen time range, etc.)

### Exploring Relationships

**Interactive exploration:**
```bash
gedcom interactive family.ged
```

Then try:
```
> search John Smith
> individual @I123@
> parents @I123@
> ancestors @I123@ 5
> descendants @I123@ 3
> relationship @I123@ @I456@
> path @I123@ @I456@
```

**What you'll get:**
- Clear relationship paths ("1st Cousin", "Great-Grandfather", etc.)
- Visual family tree navigation
- Fast queries (sub-millisecond for most operations)

### Exporting for Sharing

**Export a family branch:**
```bash
# Export all descendants of a person (8 generations deep)
gedcom export --descendants @I123@ --depth 8 -o branch.json

# Export ancestors
gedcom export --ancestors @I123@ --depth 5 -o ancestors.json

# Export to GEDCOM format for sharing
gedcom export --descendants @I123@ --depth 8 -o branch.ged --format gedcom
```

### Data Quality Checks

**Coming soon:**
```bash
# Get a data quality report
gedcom quality family.ged --report quality.html

# This will show:
# - % missing birth/death dates
# - Most common surnames
# - Most common places
# - Implausible ages (mother at 65, child before parent birth)
# - Disconnected components (families not linked)
# - Time coverage (histogram by year)
```

---

## For Community/Population Researchers

### Your Typical Dataset
- 500,000–5,000,000 individuals
- Whole populations (tribes, villages, congregations, diaspora groups)
- Many records with repetitive names (e.g., "Singh", "Johnson", patronymics)
- Goal: Find duplicates across large datasets, analyze data quality, export manageable subsets

### Understanding Your Dataset First

**Before running duplicate detection, understand your data:**
```bash
# Get data quality report (coming soon)
gedcom quality population.ged --report quality.html
```

This helps you:
- See which surnames are most common
- Identify time periods with most data
- Find regions with best coverage
- Understand data completeness

### Scoped Duplicate Detection

**Never run duplicate detection on the entire dataset without scoping:**

```bash
# Find duplicates for a specific surname in a time period
gedcom duplicates population.ged --surname "Singh" --year 1880-1920 --top 200

# Find duplicates in a specific region
gedcom duplicates population.ged --place "Uttarakhand" --year 1750-1900 --top 200

# Combine filters
gedcom duplicates population.ged \
  --surname "Bisht" \
  --place "Uttarakhand" \
  --year 1750-1900 \
  --top 200
```

**Why scoping matters:**
- Large datasets (1M+) can take 10-20 minutes even with optimization
- Very common surnames create "giant blocks" that can't be efficiently processed
- The tool will warn you when this happens and suggest alternatives

**What warnings mean:**
```
⚠️ WARNING: Duplicate detection could not evaluate 500,000 records (50.0%) 
because the dataset has extremely common surnames/years (largest block: 150,000 people). 
Try adding a place filter, widening given-name prefix matching, or running per-region.
```

This means:
- Your dataset has very repetitive names (e.g., everyone named "Singh" born in 1900)
- The tool skipped those blocks to avoid performance issues
- **Solution**: Add filters (place, time period, given name) to narrow the scope

### Exporting Manageable Subsets

**Export by criteria:**
```bash
# Export by surname and place
gedcom export --surname "Bisht" --place "Uttarakhand" --year 1750-1900 -o bisht_family.json

# Export a disconnected component (family cluster)
gedcom export --component 3 -o cluster3.json

# Export descendants of a key individual
gedcom export --descendants @I12345@ --depth 10 -o lineage.json
```

**Why this matters:**
- You can't work with 5M records at once
- Export subsets for:
  - Visualization tools
  - Sharing with collaborators
  - Focused analysis
  - Database imports

### Working with Repetitive Names

**Common scenarios:**
- Patronymics (everyone named "Johnson" or "O'Brien")
- Assigned surnames (enslaved communities, colonial records)
- Cultural naming patterns ("Singh/Kaur", "van der Berg")

**Strategies:**
1. **Filter by place first**: Most effective way to narrow scope
2. **Filter by time period**: Focus on specific generations
3. **Use given name prefixes**: Combine surname + given name initial
4. **Work regionally**: Process one village/tribe at a time
5. **Use data quality report**: Identify which regions/periods have best data

**Example workflow:**
```bash
# Step 1: Understand your data
gedcom quality population.ged --report quality.html

# Step 2: Identify regions with good data
# (Look at the report to see which places have most records)

# Step 3: Process one region at a time
gedcom duplicates population.ged --place "Village A" --year 1800-1900 --top 200
gedcom duplicates population.ged --place "Village B" --year 1800-1900 --top 200

# Step 4: Export results for analysis
gedcom export --place "Village A" --year 1800-1900 -o village_a.json
```

---

## Common Patterns

### Pattern 1: "I think these two people might be the same"

```bash
# Find all potential matches for a specific person
gedcom duplicates family.ged --individual @I123@ --top 20 --explain
```

### Pattern 2: "I want to see all descendants of my great-grandfather"

```bash
# Interactive mode
gedcom interactive family.ged
> descendants @I123@ 5

# Or export
gedcom export --descendants @I123@ --depth 5 -o great_grandfather_branch.json
```

### Pattern 3: "I'm merging two family trees and want to find duplicates"

```bash
# Compare two trees (coming soon)
gedcom duplicates --compare tree1.ged tree2.ged --top 200
```

### Pattern 4: "I have a large dataset and want to find duplicates efficiently"

```bash
# Always scope first
gedcom duplicates large.ged --place "Region A" --year 1800-1900 --top 200

# If you get warnings about giant blocks, add more filters
gedcom duplicates large.ged \
  --place "Region A" \
  --surname "CommonName" \
  --year 1800-1900 \
  --top 200
```

---

## Performance Expectations

### Private Family Researchers (10K–200K)
- **Duplicate detection**: 1-5 minutes
- **Graph queries**: Instant (< 1 second)
- **Export**: Seconds

### Community Researchers (500K–5M)
- **Duplicate detection**: 10-20 minutes (for full dataset, but you should scope!)
- **Scoped duplicate detection**: 1-5 minutes (with place/time filters)
- **Graph queries**: Still fast (< 1 second)
- **Export**: Minutes for large subsets

**Key insight**: Always scope your duplicate detection. The tool is optimized, but 5M records is still 5M records. Use filters.

---

## Getting Help

If you see warnings or unexpected results:

1. **Check the warnings**: They explain what's happening and suggest solutions
2. **Use data quality report**: Understand your dataset first
3. **Scope your operations**: Add place, time, or surname filters
4. **Start small**: Test with a subset before processing the full dataset

The tool is designed to guide you, not to fail silently. If something doesn't work as expected, the warnings and explanations should help you understand why and what to do next.

