#!/bin/bash
# Script to add license headers to source files in the CloudSnooze project

set -e

echo "Adding license headers to source files..."
echo "This script will add Apache 2.0 license headers to Go and JavaScript files."

# Get the script's directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Go to project root
cd "$SCRIPT_DIR/.."
PROJECT_ROOT="$(pwd)"

# Header templates
GO_HEADER="// Copyright 2025 Scott Friedman and CloudSnooze Contributors
// SPDX-License-Identifier: Apache-2.0

"

JS_HEADER="/**
 * Copyright 2025 Scott Friedman and CloudSnooze Contributors
 * SPDX-License-Identifier: Apache-2.0
 */

"

# Count variables
go_files_processed=0
js_files_processed=0
html_files_processed=0
md_files_processed=0
files_skipped=0
files_with_header=0

# Function to check if file already has a license header
has_license_header() {
    grep -q "Copyright" "$1" || grep -q "SPDX" "$1"
    return $?
}

# Function to add header to Go files
add_go_header() {
    local file="$1"
    
    if has_license_header "$file"; then
        files_with_header=$((files_with_header + 1))
        return
    fi
    
    # Check if the file starts with a package declaration
    first_line=$(head -1 "$file")
    if [[ "$first_line" == package* ]]; then
        # File starts with package, insert header before it
        echo -e "${GO_HEADER}${first_line}" > "$file.new"
        tail -n +2 "$file" >> "$file.new"
        mv "$file.new" "$file"
        go_files_processed=$((go_files_processed + 1))
    else
        # Insert at the beginning
        echo -e "${GO_HEADER}$(cat "$file")" > "$file.new"
        mv "$file.new" "$file"
        go_files_processed=$((go_files_processed + 1))
    fi
}

# Function to add header to JavaScript files
add_js_header() {
    local file="$1"
    
    if has_license_header "$file"; then
        files_with_header=$((files_with_header + 1))
        return
    fi
    
    # Insert at the beginning
    echo -e "${JS_HEADER}$(cat "$file")" > "$file.new"
    mv "$file.new" "$file"
    js_files_processed=$((js_files_processed + 1))
}

# Function to add header to HTML files
add_html_header() {
    local file="$1"
    
    if has_license_header "$file"; then
        files_with_header=$((files_with_header + 1))
        return
    fi
    
    # Check if file has a DOCTYPE or HTML tag
    if grep -q "<!DOCTYPE" "$file" || grep -q "<html" "$file"; then
        # Insert after the first line (after DOCTYPE or HTML tag)
        head -1 "$file" > "$file.new"
        echo "<!--
  Copyright 2025 Scott Friedman and CloudSnooze Contributors
  SPDX-License-Identifier: Apache-2.0
-->" >> "$file.new"
        tail -n +2 "$file" >> "$file.new"
        mv "$file.new" "$file"
        html_files_processed=$((html_files_processed + 1))
    else
        # Insert at the beginning
        echo "<!--
  Copyright 2025 Scott Friedman and CloudSnooze Contributors
  SPDX-License-Identifier: Apache-2.0
-->" > "$file.new"
        cat "$file" >> "$file.new"
        mv "$file.new" "$file"
        html_files_processed=$((html_files_processed + 1))
    fi
}

# Find and process Go files
echo "Processing Go files..."
find "$PROJECT_ROOT" -name "*.go" -not -path "*/vendor/*" -not -path "*/node_modules/*" | while read -r file; do
    echo "  Adding header to $file"
    add_go_header "$file"
done

# Find and process JavaScript files
echo "Processing JavaScript files..."
find "$PROJECT_ROOT" -name "*.js" -not -path "*/vendor/*" -not -path "*/node_modules/*" -not -path "*/dist/*" | while read -r file; do
    echo "  Adding header to $file"
    add_js_header "$file"
done

# Find and process HTML files
echo "Processing HTML files..."
find "$PROJECT_ROOT" -name "*.html" -not -path "*/vendor/*" -not -path "*/node_modules/*" -not -path "*/dist/*" | while read -r file; do
    echo "  Adding header to $file"
    add_html_header "$file"
done

# Function to add header to Markdown files
add_md_header() {
    local file="$1"
    
    if has_license_header "$file"; then
        files_with_header=$((files_with_header + 1))
        return
    fi
    
    # Check if file starts with a front matter (---)
    if grep -q "^---" "$file"; then
        # Get the end of front matter
        line_num=$(grep -n "^---" "$file" | sed -n '2p' | cut -d: -f1)
        if [ -n "$line_num" ]; then
            # Add header after front matter
            head -n "$line_num" "$file" > "$file.new"
            echo "" >> "$file.new"
            echo "<!--" >> "$file.new"
            echo "Copyright 2025 Scott Friedman and CloudSnooze Contributors" >> "$file.new"
            echo "SPDX-License-Identifier: Apache-2.0" >> "$file.new"
            echo "-->" >> "$file.new"
            echo "" >> "$file.new"
            tail -n +$((line_num + 1)) "$file" >> "$file.new"
            mv "$file.new" "$file"
            md_files_processed=$((md_files_processed + 1))
            return
        fi
    fi
    
    # Check if the file starts with a # heading
    if grep -q "^#" "$file"; then
        # Insert before the first heading
        echo "<!--" > "$file.new"
        echo "Copyright 2025 Scott Friedman and CloudSnooze Contributors" >> "$file.new"
        echo "SPDX-License-Identifier: Apache-2.0" >> "$file.new"
        echo "-->" >> "$file.new"
        echo "" >> "$file.new"
        cat "$file" >> "$file.new"
        mv "$file.new" "$file"
        md_files_processed=$((md_files_processed + 1))
    else
        # Insert at the beginning
        echo "<!--" > "$file.new"
        echo "Copyright 2025 Scott Friedman and CloudSnooze Contributors" >> "$file.new"
        echo "SPDX-License-Identifier: Apache-2.0" >> "$file.new"
        echo "-->" >> "$file.new"
        echo "" >> "$file.new"
        cat "$file" >> "$file.new"
        mv "$file.new" "$file"
        md_files_processed=$((md_files_processed + 1))
    fi
}

# Find and process Markdown files
echo "Processing Markdown files..."
find "$PROJECT_ROOT" -name "*.md" -not -path "*/vendor/*" -not -path "*/node_modules/*" -not -path "*/dist/*" -not -name "LICENSE.md" -not -name "README.md" | while read -r file; do
    echo "  Adding header to $file"
    add_md_header "$file"
done

echo "License header addition complete."
echo "$go_files_processed Go files processed"
echo "$js_files_processed JavaScript files processed"
echo "$html_files_processed HTML files processed"
echo "$md_files_processed Markdown files processed"
echo "$files_with_header files already had headers"
echo "$files_skipped files skipped (errors or other issues)"
echo
echo "Please review the changes before committing."