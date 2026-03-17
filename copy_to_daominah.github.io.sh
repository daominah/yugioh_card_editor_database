#!/bin/bash

# Script to copy necessary files from
# dir github.com/daominah/yugioh_card_editor/web to
# dir github.com/daominah/daominah.github.io,
# so we can push the changes to serve on GitHub page

set -euo pipefail

sourceDir="./web"
targetDir="$HOME/go/src/github.com/daominah/daominah.github.io"


# files to copy (one per line),
# note that target "konami_data/konami_db_en.js" updated by github workflow, so must not list here
files=(
    "index.html"
    "index.js"
    "index.css"
    "favicon.ico"
    "konami_data/alt_arts.js"
)

# directories to copy all files inside
directories=(
    "card_frame"
    "external_lib"
    "font"
    "icon"
)


# Check if source and target directories exist, if not exit with error
if [ ! -d "$sourceDir" ]; then
    echo "Error: Source directory '$sourceDir' does not exist"
    exit 1
fi
if [ ! -d "$targetDir" ]; then
    echo "Error: Target directory '$targetDir' does not exist"
    exit 1
fi

echo "Copying files from $sourceDir to $targetDir..."

# Copy files
echo "Copying files..."
for file in "${files[@]}"; do
    source_file="$sourceDir/$file"
    if [ ! -f "$source_file" ]; then
        echo "Warning: Source file '$source_file' does not exist, skipping..."
        continue
    fi
    # Create target directory if file is in subdirectory
    target_file="$targetDir/$file"
    target_dir=$(dirname "$target_file")
    if [ ! -d "$target_dir" ]; then
        mkdir -p "$target_dir"
    fi
    cp "$source_file" "$target_file"
    echo "  Copied: $file"
done

# Copy directories
echo "Copying directories..."
for dir in "${directories[@]}"; do
    source_dir="$sourceDir/$dir"
    if [ ! -d "$source_dir" ]; then
        echo "Warning: Source directory '$source_dir' does not exist, skipping..."
        continue
    fi
    # Create target directory if it doesn't exist
    mkdir -p "$targetDir/$dir"
    # Copy directory contents recursively
    # Using shopt to include hidden files, or fallback to simple cp
    if [ -n "$(ls -A "$source_dir" 2>/dev/null)" ]; then
        # Copy all files including hidden ones
        (cd "$source_dir" && cp -r . "$targetDir/$dir/")
    fi
    echo "  Copied directory: $dir"
done

echo "Copy completed successfully!"
echo "Files copied to: $targetDir"
echo "==============================="
echo "You can check diff in $targetDir then push the changes to serve on GitHub page"
