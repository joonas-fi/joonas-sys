if [ "$#" -ne 1 ]; then
  echo "Usage: $0 video_filename"
  exit 1
fi

filename="$1"
output="output.mp4"

if [ ! -f "$filename" ]; then
  echo "Error: File '$filename' not found!"
  exit 1
fi

echo "From (time in format hh:mm:ss):"
echo -n "> "
read -r from

echo "To (same format, end time NOT length of video-to-be-kept)"
echo -n "> "
read -r to

ffmpeg -ss "$from" -to "$to" -i "$filename" -c copy "$output"

echo "saved to $output"
