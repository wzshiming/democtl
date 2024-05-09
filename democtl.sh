#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

COLS=86
ROWS=24
SVG_TERM=""
SVG_PROFILE=""
SIM_ENV=()

export PLAY_PS1="$ "

CACHE_DIR=${TMPDIR:-/tmp}/democtl
ARGS=()

export ASCIINEMA_PATH="${CACHE_DIR}/asciinema"
export PLAYPTY_PATH="${CACHE_DIR}/playpty"
export PYTHONPATH="${ASCIINEMA_PATH}:${PLAYPTY_PATH}"
export PATH="${ASCIINEMA_PATH}/bin:${PLAYPTY_PATH}/bin:${CACHE_DIR}/node_modules/.bin:${PATH}"

function usage() {
  echo "Usage: ${0} <input> <output> [--help] [options...]"
  echo "  <input> input file"
  echo "  <output> output file"
  echo "  --help show this help"
  echo "  --cols=${COLS} cols of the terminal"
  echo "  --rows=${ROWS} rows of the terminal"
  echo "  --ps1=${PLAY_PS1} ps1 of the recording"
  echo "  --term=${SVG_TERM} terminal type"
  echo "  --profile=${SVG_PROFILE} terminal profile"
  echo "  --env=${SIM_ENV[*]} environment variables will be passed to the simulation script"
}

# args parses the arguments.
function args() {
  local arg

  while [[ $# -gt 0 ]]; do
    arg="$1"
    case "${arg}" in
    --cols | --cols=*)
      [[ "${arg#*=}" != "${arg}" ]] && COLS="${arg#*=}" || { COLS="${2}" && shift; } || :
      shift
      ;;
    --rows | --rows=*)
      [[ "${arg#*=}" != "${arg}" ]] && ROWS="${arg#*=}" || { ROWS="${2}" && shift; } || :
      shift
      ;;
    --ps1 | --ps1=*)
      [[ "${arg#*=}" != "${arg}" ]] && PLAY_PS1="${arg#*=}" || { PLAY_PS1="${2}" && shift; } || :
      shift
      ;;
    --term | --term=*)
      [[ "${arg#*=}" != "${arg}" ]] && SVG_TERM="${arg#*=}" || { SVG_TERM="${2}" && shift; } || :
      shift
      ;;
    --profile | --profile=*)
      [[ "${arg#*=}" != "${arg}" ]] && SVG_PROFILE="${arg#*=}" || { SVG_PROFILE="${2}" && shift; } || :
      shift
      ;;
    --env | --env=*)
      [[ "${arg#*=}" != "${arg}" ]] && SIM_ENV+=("${arg#*=}") || { SIM_ENV+=("${2}") && shift; } || :
        shift
        ;;
    --help)
      usage
      exit 0
      ;;
    --*)
      echo "Unknown argument: ${arg}"
      usage
      exit 1
      ;;
    *)
      ARGS+=("${arg}")
      shift
      ;;
    esac
  done
}

# command_exist checks if the command exists.
function command_exist() {
  local command="${1}"
  type "${command}" >/dev/null 2>&1
}

# install_playpty installs playpty.
function install_playpty() {
  if command_exist playpty; then
    return 0
  elif command_exist pip3; then
    pip3 install playpty --target "${PLAYPTY_PATH}" >&2
  else
    echo "playpty is not installed" >&2
    return 1
  fi
}

# install_asciinema installs asciinema.
function install_asciinema() {
  if command_exist asciinema; then
    return 0
  elif command_exist pip3; then
    pip3 install asciinema --target "${ASCIINEMA_PATH}" >&2
  else
    echo "asciinema is not installed" >&2
    return 1
  fi
}

# install_svg_term_cli installs svg-term-cli.
function install_svg_term_cli() {
  if command_exist svg-term; then
    return 0
  elif command_exist npm; then
    npm install --save-dev svg-term-cli --prefix "${CACHE_DIR}" >&2
  else
    echo "svg-term is not installed" >&2
    return 1
  fi
}

# install_svg_to_video installs svg-to-video.
function install_svg_to_video() {
  if command_exist svg-to-video; then
    return 0
  elif command_exist npm; then
    npm install --save-dev https://github.com/wzshiming/svg-to-video --prefix "${CACHE_DIR}" >&2
  else
    echo "svg-to-video is not installed" >&2
    return 1
  fi
}

# ext_file returns the extension of the input file.
function ext_file() {
  local file="${1}"
  echo "${file##*.}"
}

# ext_replace replaces the extension of the input file with the output extension.
function ext_replace() {
  local file="${1}"
  local ext="${2}"
  echo "${file%.*}.${ext}"
}

# demo2cast converts the input demo file to the output cast file.
function demo2cast() {
  local input="${1}"
  local output="${2}"
  echo "Recording ${input} to ${output}" >&2

  asciinema rec \
    "${output}" \
    --overwrite \
    --cols "${COLS}" \
    --rows "${ROWS}" \
    --env "" \
    --command "playpty ${input} --ps1='${PLAY_PS1}' --cols=${COLS} --rows=${ROWS} --env ${SIM_ENV[*]}"
}

# cast2svg converts the input cast file to the output svg file.
function cast2svg() {
  local input="${1}"
  local output="${2}"
  local args=()
  echo "Converting ${input} to ${output}" >&2

  if [[ "${SVG_TERM}" != "" ]]; then
    args+=("--term" "${SVG_TERM}")
  fi

  if [[ "${SVG_PROFILE}" != "" ]]; then
    args+=("--profile" "${SVG_PROFILE}")
  fi
  svg-term \
    --in "${input}" \
    --out "${output}" \
    --window \
    "${args[@]}"
}

# svg2video converts the input svg file to the output video file.
function svg2video() {
  local input="${1}"
  local output="${2}"
  echo "Converting ${input} to ${output}" >&2

  svg-to-video \
    "${input}" \
    "${output}" \
    --delay-start 5 \
    --headless
}

# convert converts the input file to the output file.
# The input file can be a demo, cast, or svg file.
# The output file can be a cast, svg, or mp4 file.
function convert() {
  local input="${1}"
  local output="${2}"

  local castfile
  local viedofile

  local outext
  local inext

  outext=$(ext_file "${output}")
  inext=$(ext_file "${input}")
  case "${outext}" in
  cast)
    case "${inext}" in
    demo)
      install_asciinema
      install_playpty

      demo2cast "${input}" "${output}"
      return 0
      ;;
    *)
      echo "Unsupported input file type: ${inext}"
      return 1
      ;;
    esac
    ;;
  svg)
    case "${inext}" in
    cast)
      install_svg_term_cli

      cast2svg "${input}" "${output}"
      return 0
      ;;
    demo)
      install_asciinema
      install_playpty
      install_svg_term_cli

      castfile=$(ext_replace "${output}" "cast")
      demo2cast "${input}" "${castfile}"
      cast2svg "${castfile}" "${output}"
      return 0
      ;;
    *)
      echo "Unsupported input file type: ${inext}"
      return 1
      ;;
    esac
    ;;
  mp4)
    case "${inext}" in
    svg)
      install_svg_to_video

      svg2video "${input}" "${output}"
      return 0
      ;;
    cast)
      install_svg_term_cli
      install_svg_to_video

      viedofile=$(ext_replace "${output}" "svg")
      cast2svg "${input}" "${viedofile}"
      svg2video "${viedofile}" "${output}"
      return 0
      ;;
    demo)
      install_asciinema
      install_playpty
      install_svg_term_cli
      install_svg_to_video

      viedofile=$(ext_replace "${output}" "svg")
      castfile=$(ext_replace "${output}" "cast")
      demo2cast "${input}" "${castfile}"
      cast2svg "${castfile}" "${viedofile}"
      svg2video "${viedofile}" "${output}"
      return 0
      ;;
    *)
      echo "Unsupported input file type: ${inext}"
      return 1
      ;;
    esac
    ;;
  *)
    echo "Unsupported output file type: ${outext}"
    return 1
    ;;
  esac
}

function main() {
  if [[ "${#ARGS[*]}" -lt 1 ]]; then
    usage
    exit 1
  fi

  INPUT_FILE="${ARGS[0]}"

  if [[ "${#ARGS[*]}" -gt 1 ]]; then
    OUTPUT_FILE="${ARGS[1]}"
  else
    # If the output file is not specified, use the same name as the input file.
    OUTPUT_FILE="$(ext_replace "${INPUT_FILE}" "svg")"
  fi

  convert "${INPUT_FILE}" "${OUTPUT_FILE}"
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  args "$@"
  main
fi
