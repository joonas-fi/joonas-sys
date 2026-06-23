if [ "$#" -lt 2 ]; then
  echo "Usage: $0 output.mp4 input1.mp4 input2.mp4 [input3.mp4 ...]"
  exit 1
fi

output_file="$1"
shift

temp_file="$(mktemp concat-filelist-tmp.XXXXXX.txt)"

cleanup() {
  rm -f "$temp_file"
}
trap cleanup EXIT

for input_file in "$@"; do
  if [ ! -f "$input_file" ]; then
    echo "Error: File '$input_file' not found!"
    exit 1
  fi

  escaped_input_file=${input_file//\'/\'\\\'\'}
  printf "file '%s'\n" "$escaped_input_file" >> "$temp_file"
done

ffmpeg -f concat -safe 0 -i "$temp_file" -c copy "$output_file"

echo "Concatenation complete: $output_file"
