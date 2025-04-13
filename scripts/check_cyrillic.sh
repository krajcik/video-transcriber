#!/bin/bash

found=0

while IFS= read -r file; do
  if grep -n -E '//.*[А-Яа-яЁё]|/\*.*[А-Яа-яЁё].*\*/|".*[А-Яа-яЁё].*"' "$file"; then
    echo "Cyrillic text found in $file"
    found=1
  fi
done < <(find . -type f -name "*.go" ! -path "./vendor/*")

if [ $found -eq 1 ]; then
  echo "Non-English (Cyrillic) text is not allowed in code or comments"
  exit 1
fi

exit 0
