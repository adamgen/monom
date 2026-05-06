---
name: /validate-constitution
id: validate-constitution
category: Workflow
description: "Validate staged or recent changes against the monom constitution"
---

1. Read `constitution.md` fully.
2. Review the changes or files specified by the user (or the current git diff if none are specified).
3. For each principle in the constitution, give a clear verdict: **pass**, **violation**, or **not applicable**. For any violation, explain specifically what was done and what the principle requires instead.

## Output format

```
## Constitution Validation

### 1. Go Owns Logic, Shell Owns Surface — PASS / VIOLATION / N/A
<one sentence verdict, or explanation of violation>

### 2. Minimize Subprocess Roundtrips — PASS / VIOLATION / N/A
<one sentence verdict, or explanation of violation>

### 3. User Config Interface — PASS / VIOLATION / N/A
<one sentence verdict, or explanation of violation>

### 4. Testing Is Layered — PASS / VIOLATION / N/A
<one sentence verdict, or explanation of violation>

### 5. Static and Lint Checks — PASS / VIOLATION / N/A
<one sentence verdict, or explanation of violation>

### 6. Terminology — PASS / VIOLATION / N/A
<one sentence verdict, or explanation of violation>

---
**Overall: PASS / VIOLATIONS FOUND**
<If violations: list each one with a concrete fix suggestion>
```

Be direct. Do not soften violations. If something violates the constitution, say so clearly and explain the fix.
