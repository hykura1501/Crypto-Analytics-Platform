#!/bin/bash
# Render Mermaid diagrams to PNG cho báo cáo
# Cách 1: npm install -g @mermaid-js/mermaid-cli  rồi chạy: ./export_patterns.sh
# Cách 2: Copy nội dung từng file .md vào https://mermaid.live và Export PNG

OUT_DIR="../reports/photo"
mkdir -p "$OUT_DIR"

for f in pattern_singleton.md pattern_adapter.md pattern_observer.md pattern_strategy.md pattern_template_method.md; do
  name="${f%.md}"
  if command -v mmdc &> /dev/null; then
    mmdc -i "$f" -o "$OUT_DIR/${name}.png"
    echo "OK: $name.png"
  else
    echo "mmdc chưa cài. Paste nội dung $f vào mermaid.live để export PNG -> $OUT_DIR/${name}.png"
  fi
done
