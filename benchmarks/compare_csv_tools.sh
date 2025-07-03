#!/bin/bash
set -e

DATADIR="$(pwd)/benchmarks"

# Generate large_valid_sample.csv if not present
if [ ! -f "$DATADIR/large_valid_sample.csv" ]; then
  echo "Generating large_valid_sample.csv..."
  go run benchmarks/gen_large_valid_csv/main.go "$DATADIR/large_valid_sample.csv"
else
  echo "large_valid_sample.csv already exists. Skipping generation."
fi

# Generate large_invalid_sample.csv if not present
if [ ! -f "$DATADIR/large_invalid_sample.csv" ]; then
  echo "Generating large_invalid_sample.csv..."
  go run benchmarks/gen_large_invalid_csv/main.go "$DATADIR/large_invalid_sample.csv"
else
  echo "large_invalid_sample.csv already exists. Skipping generation."
fi

# Generate large_schema.json if not present
if [ ! -f "$DATADIR/large_schema.json" ]; then
  echo "Copying large_schema.json..."
  cp benchmarks/large_schema.json "$DATADIR/large_schema.json"
else
  echo "large_schema.json already exists. Skipping copy."
fi

# Generate large_schema.csvw.json if not present
if [ ! -f "$DATADIR/large_schema.csvw.json" ]; then
  echo "Copying large_schema.csvw.json..."
  cp benchmarks/large_schema.csvw.json "$DATADIR/large_schema.csvw.json"
else
  echo "large_schema.csvw.json already exists. Skipping copy."
fi

# Build images

echo "Building csvlinter..."
docker build -f benchmarks/Dockerfile.csvlinter -t csvlinter:latest .
echo "Building csvkit..."
docker build -f benchmarks/Dockerfile.csvkit -t csvkit:latest .
echo "Building csvlint..."
docker build -f benchmarks/Dockerfile.csvlint -t csvlint:latest .

# Run scenarios for each tool
for TOOL in csvlinter csvkit csvlint; do
  for SCEN in valid invalid valid_schema; do
    # Only run valid_schema for csvlinter and csvlint
    if [ "$SCEN" = "valid_schema" ] && [ "$TOOL" = "csvkit" ]; then
      continue
    fi
    F="benchmarks/${TOOL}_${SCEN}.time"
    case $SCEN in
      valid)
        CSV=large_valid_sample.csv
        SCHEMA=""
        ;;
      invalid)
        CSV=large_invalid_sample.csv
        SCHEMA=""
        ;;
      valid_schema)
        CSV=large_valid_sample.csv
        if [ "$TOOL" = "csvlinter" ]; then
          SCHEMA=large_schema.json
        elif [ "$TOOL" = "csvlint" ]; then
          SCHEMA=large_schema.csvw.json
        fi
        ;;
    esac
    echo "Running $TOOL on $CSV${SCHEMA:+ with $SCHEMA}..."
    case $TOOL in
      csvlinter)
        if [ -n "$SCHEMA" ]; then
          docker run --rm --entrypoint "" -v "$DATADIR:/testdata" csvlinter:latest /usr/bin/time -v ./csvlinter validate /testdata/$CSV --schema /testdata/$SCHEMA > /dev/null 2> $F
        else
          docker run --rm --entrypoint "" -v "$DATADIR:/testdata" csvlinter:latest /usr/bin/time -v ./csvlinter validate /testdata/$CSV > /dev/null 2> $F
        fi
        ;;
      csvkit)
        docker run --rm --entrypoint "" -v "$DATADIR:/data" csvkit:latest sh -c "/usr/bin/time -v csvstat /data/$CSV" > /dev/null 2> $F
        ;;
      csvlint)
        if [ -n "$SCHEMA" ]; then
          docker run --rm --entrypoint "" -v "$DATADIR:/data" csvlint:latest /usr/bin/time -v csvlint /data/$CSV --schema=/data/$SCHEMA > /dev/null 2> $F
        else
          docker run --rm --entrypoint "" -v "$DATADIR:/data" csvlint:latest /usr/bin/time -v csvlint /data/$CSV > /dev/null 2> $F
        fi
        ;;
    esac
  done
done

echo "Results saved to benchmarks/csvlinter_*.time, benchmarks/csvkit_*.time, benchmarks/csvlint_*.time in the benchmarks/ directory."

echo "Compiling results into markdown table..."

# Table header
TABLE="| Tool     | Scenario         | Time (s) | Max RSS (MB) |\n|----------|------------------|----------|--------------|\n"

# Helper to extract time in seconds from h:mm:ss or m:ss
parse_time() {
  local t="$1"
  if [[ $t =~ ([0-9]+):([0-9]+)\.([0-9]+) ]]; then
    local min=${BASH_REMATCH[1]}
    local sec=${BASH_REMATCH[2]}
    local frac=${BASH_REMATCH[3]}
    echo "scale=2; $min*60+$sec+0.$frac" | bc
  else
    echo "$t"
  fi
}

for TOOL in csvlinter csvkit csvlint; do
  for SCEN in valid invalid valid_schema; do
    F="benchmarks/${TOOL}_${SCEN}.time"
    if [ -f "$F" ]; then
      # Extract wall clock time
      T=$(grep "Elapsed (wall clock) time" "$F" | awk -F": " '{print $2}')
      T_S=$(parse_time "$T")
      # Extract max RSS
      RSS=$(grep "Maximum resident set size" "$F" | awk -F": " '{print $2}')
      RSS_MB=$((RSS/1024))
      # Pretty scenario name
      case $SCEN in
        valid) SCEN_LABEL="Valid";;
        invalid) SCEN_LABEL="Invalid";;
        valid_schema) SCEN_LABEL="Valid+Schema (should fail)";;
      esac
      # Pretty tool name
      case $TOOL in
        csvlinter) TOOL_LABEL="csvlinter";;
        csvkit) TOOL_LABEL="csvkit";;
        csvlint) TOOL_LABEL="csvlint";;
      esac
      TABLE+="| $TOOL_LABEL | $SCEN_LABEL | $T_S | $RSS_MB |\n"
    fi
  done
done

# Generate benchmark.md from template
TEMPLATE="benchmarks/benchmark_template.md"
BENCHMARK_MD="benchmark.md"

if [ ! -f "$TEMPLATE" ]; then
  echo "Benchmark template not found: $TEMPLATE"
  exit 1
fi

# Read template and replace placeholder
CONTENT=$(cat "$TEMPLATE")
CONTENT="${CONTENT//\{\{BENCHMARK_TABLE\}\}/$TABLE}"
echo -e "$CONTENT" > "$BENCHMARK_MD"

echo "Benchmark results written to $BENCHMARK_MD." 