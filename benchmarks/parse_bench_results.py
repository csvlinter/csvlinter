import re
import sys

files = [
    ("csvlinter", "benchmarks/csvlinter.time"),
    ("csvkit", "benchmarks/csvkit.time"),
    ("csvlint", "benchmarks/csvlint.time"),
]

results = []

for tool, fname in files:
    try:
        with open(fname) as f:
            content = f.read()
    except FileNotFoundError:
        results.append((tool, "N/A", "N/A", "N/A", "File not found"))
        continue
    user_time = re.search(r"User time \(seconds\): ([0-9.]+)", content)
    elapsed = re.search(r"Elapsed \(wall clock\) time \(h:mm:ss or m:ss\): ([0-9:.]+)", content)
    max_mem = re.search(r"Maximum resident set size \(kbytes\): ([0-9]+)", content)
    errors = re.findall(r"error|invalid|fail|warn", content, re.IGNORECASE)
    results.append((
        tool,
        user_time.group(1) if user_time else "-",
        elapsed.group(1) if elapsed else "-",
        max_mem.group(1) if max_mem else "-",
        len(errors)
    ))

print("| Tool      | User Time (s) | Elapsed Time | Max Memory (KB) | Errors/Warnings |")
print("|-----------|---------------|--------------|-----------------|-----------------|")
for row in results:
    print(f"| {row[0]:<9} | {row[1]:<13} | {row[2]:<12} | {row[3]:<15} | {row[4]:<15} |") 