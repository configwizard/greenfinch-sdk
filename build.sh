#!/bin/bash

# Check for the necessary arguments
if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <starting-directory> <build-tag>"
    exit 1
fi

# Starting directory and build tag
start_dir=$1
build_tag=$2

# Delete any existing build_errors.log files in the directories
find "$start_dir" -name 'build_results.log' -delete
find "$start_dir" -name 'build_errors.log' -delete
# Find all directories recursively and attempt to build Go programs inside them in parallel
pids=()
while IFS= read -r dir; do
    (
          echo "Attempting build in directory $dir with tag $build_tag..."
          binary_name=$(basename "$dir")

          # Capture errors in a variable
          errors=$(go build -tags="$build_tag" -o "$dir/$binary_name" "$dir" 2>&1)

          # Check for "no Go files" error and don't record these
          if [[ $errors =~ "no Go files in" ]]; then
              exit
          fi

          # Record the outcome to a results log
          if [ -z "$errors" ]; then
              echo -e "SUCCESS\n" > "$dir/build_results.log"
          else
              echo -e "ERROR\n" > "$dir/build_results.log"
              echo "$errors" >> "$dir/build_results.log"
          fi
    ) &
    pids+=($!)
done < <(find "$start_dir" -type d)

# Wait for all builds to complete
for pid in "${pids[@]}"; do
    wait $pid
done
echo "All builds completed!"
# Tally up the results
total_builds=0
errored_builds=0
errored_dirs=()

while IFS= read -r result_file; do
    result=$(head -n 1 "$result_file")

    total_builds=$((total_builds+1))

    if [[ "$result" == "ERROR" ]]; then
        errored_builds=$((errored_builds+1))
        errored_dirs+=($(dirname "$result_file"))
    fi
done < <(find "$start_dir" -name 'build_results.log')


# Display the summary

echo "Total builds attempted: $total_builds"
echo "Total builds with errors: $errored_builds"

if [ "$errored_builds" -gt 0 ]; then
    echo "Directories with errors:"
    for dir in "${errored_dirs[@]}"; do
        echo " - $dir"
    done
fi
