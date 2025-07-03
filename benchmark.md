# csvlinter benchmarks & tool comparison

## Tool comparison notes

**csvlinter** is a strict CSV validator with JSON Schema support, designed for robust structure and content validation. By contrast, **csvkit** (including tools like `csvstat` and `csvclean`) and **csvlint** have different focuses:

- **csvlinter**: Validates structure, content, and supports JSON Schema. Provides explicit pass/fail and error reporting.
- **csvkit**:
  - `csvstat`: Provides statistics and type inference, but does not strictly validate content or enforce schemas.
  - `csvclean`: Detects malformed rows and structural issues, but does not validate content or support schemas.
- **csvlint**: Supports validation and schema checking, but only with [CSV on the Web (CSVW)](https://w3c.github.io/csvw/) metadata, not JSON Schema.

### When is comparison fair?
- **Parsing/analysis speed**: Comparing `csvlinter` and `csvkit` (`csvstat`) is fair for raw reading and analysis.
- **Structural validation**: Comparing `csvlinter` and `csvkit` (`csvclean`) is fair for detecting malformed CSVs.
- **Schema/content validation**: Only `csvlinter` (JSON Schema) and `csvlint` (CSVW) support this; `csvkit` does not.

### Feature comparison table

| Tool      | Structural validation | Schema validation      | Stats/analysis |
|-----------|----------------------|-----------------------|---------------|
| csvlinter | Yes                  | Yes (JSON Schema)     | No            |
| csvkit    | csvclean: Yes        | No                    | csvstat: Yes  |
| csvlint   | Yes                  | Yes (CSVW only)       | No            |

> **Note:** In the benchmarks below, only scenarios supported by each tool are run. For example, schema validation is only benchmarked for `csvlinter` and `csvlint`.

## Benchmarks

To run the benchmarks (requires Docker):

```bash
./benchmarks/compare_csv_tools.sh
```

This will generate large CSVs in benchmarks/, run all tools, and update this benchmark file.

## Benchmark results

| Tool     | Scenario         | Time (s) | Max RSS (MB) |
|----------|------------------|----------|--------------|
| csvlinter | Valid | .53 | 198 |
| csvlinter | Invalid | .60 | 197 |
| csvlinter | Valid+Schema (should fail) | .22 | 199 |
| csvkit | Valid | 15.36 | 743 |
| csvkit | Invalid | 9.38 | 620 |
| csvkit | Valid+Schema (should fail) | 9.64 | 743 |
| csvlint | Valid | 51.04 | 706 |
| csvlint | Invalid | 50.76 | 703 |
| csvlint | Valid+Schema (should fail) | 49.96 | 705 |
 
