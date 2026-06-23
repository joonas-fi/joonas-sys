if [ "$#" -eq 0 ]; then
  echo "Usage: $0 <input_file> [--skip-transmux] [--fix-dimensions]"
  exit 1
fi

input_file=""
skip_transmux=0
fix_dimensions=0

for arg in "$@"; do
  case "$arg" in
    --skip-transmux)
      skip_transmux=1
      ;;
    --fix-dimensions)
      fix_dimensions=1
      ;;
    --help | -h)
      echo "Usage: $0 <input_file> [--skip-transmux] [--fix-dimensions]"
      exit 0
      ;;
    --*)
      echo "Error: Unknown option '$arg'"
      exit 1
      ;;
    *)
      if [ -n "$input_file" ]; then
        echo "Error: Multiple input files provided: '$input_file' and '$arg'"
        exit 1
      fi

      input_file="$arg"
      ;;
  esac
done

if [ -z "$input_file" ]; then
  echo "Error: No input file provided!"
  exit 1
fi

if [ ! -f "$input_file" ]; then
  echo "Error: File '$input_file' not found!"
  exit 1
fi

output_file="${input_file%.*}.mp4"
ffmpeg_args=(-i "$input_file")

if [ "$fix_dimensions" -eq 1 ]; then
  ffmpeg_args+=(-vf "scale=trunc(iw/2)*2:trunc(ih/2)*2")
elif [ "$skip_transmux" -eq 0 ]; then
  ffmpeg_args+=(-c copy)
fi

ffmpeg_args+=("$output_file")

printf 'Running command: ffmpeg'
printf ' %q' "${ffmpeg_args[@]}"
printf '\n'

ffmpeg "${ffmpeg_args[@]}"

echo "Process complete: $output_file"
