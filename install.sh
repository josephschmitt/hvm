#!/bin/sh

# This script installs HVM
#
# Quick install: `curl -sfL "https://github.com/josephschmitt/hvm/raw/main/install.sh" | bash`

set -e -u

# The follow environment variables can be optionally exported before running the install script
# in order to configure the installer
#
# HVM_INSTALL_VERSION:
#   The version to install. Defaults to "latest"
# HVM_PLATFORM:
#   A combination of "<os name>-<machine architecture>". The installer does its best to figure this
#   out automatically, but you can set this if it fails or you want to override.
# HVM_INSTALL_LOCATION:
#   Alternate install location for the binary. Defaults to /usr/local/bin

green() {
  printf '\033[0;32m%s\033[0m\n' "$*"
}

red() {
  printf '\033[0;31m%s\033[0m\n' "$*"
}

cyan() {
  printf '\033[0;36m%s\033[0m\n' "$*"
}

yellow() {
  printf '\033[0;33m%s\033[0m\n' "$*"
}

debug() {
  echo "$(cyan DEBUG): $*" >&2
}

warning() {
  echo "$(yellow WARNING): $*" >&2
}

error() {
  echo "$(red ERROR): $*" >&2
}

stderr() {
  echo "$*" >&2
}

platform=''
machine=$(uname -m)

if [ "${HVM_PLATFORM:-x}" != "x" ]; then
  platform="${HVM_PLATFORM}"
else
  case "$(uname -s | tr '[:upper:]' '[:lower:]')" in
    "linux")
      case "${machine}" in
        "arm64"* | "aarch64"* ) platform='linux-arm64' ;;
        "arm"* | "aarch"*) platform='linux-arm' ;;
        *"86") platform='linux-386' ;;
        *"64") platform='linux-amd64' ;;
      esac
      ;;
    "darwin")
      case "${machine}" in
        "arm"*) platform='darwin-arm64' ;;
        *"64") platform='darwin-amd64' ;;
      esac
      ;;
    *"freebsd"*)
      case "${machine}" in
        *"86") platform='freebsd-386' ;;
        *"64") platform='freebsd-amd64' ;;
      esac
      ;;
    "openbsd")
      case "${machine}" in
        *"86") platform='openbsd-386' ;;
        *"64") platform='openbsd-amd64' ;;
      esac
      ;;
    "netbsd")
      case "${machine}" in
        *"86") platform='netbsd-386' ;;
        *"64") platform='netbsd-amd64' ;;
      esac
      ;;
    "msys"*|"cygwin"*|"mingw"*|*"_nt"*|"win"*)
      case "${machine}" in
        *"86") platform='win-386' ;;
        *"64") platform='win-amd64' ;;
      esac
      ;;
  esac
fi

if [ "x${platform}" = "x" ]; then
  cat << 'EOM'
/=====================================\\
|      COULD NOT DETECT PLATFORM      |
\\=====================================/

Uh oh! We couldn't automatically detect your operating system.

To continue with installation, please choose from one of the following values:

- linux-amd64
- linux-arm64
- darwin-amd64
- darwin-arm64
- win-amd64
- win-arm64

Export your selection as the HVM_PLATFORM environment variable, and then
re-run this script.

For example:

  $ export HVM_PLATFORM=linux-amd64
  $ curl $(https://raw.githubusercontent.com/josephschmitt/hvm/main/install.sh) | bash

EOM
  exit 1
fi

version="${HVM_INSTALL_VERSION:-"latest"}"
if [ "${version}" == "latest" ]; then
  version="$(curl --silent "https://api.github.com/repos/josephschmitt/hvm/releases/latest" |
    grep '"tag_name":' |
    sed -E 's/.*"([^"]+)".*/\1/'
  )"
fi

if test "$version" = "${version#v}"; then
  version="v${version}"
fi

if [ "x${platform}" = "win-amd64" ] || [ "x${platform}" = "win-arm64" ]; then
  extension='zip'
else
  extension='tar.gz'
fi

# Download and unzip
tmpdir="${TMPDIR:-/tmp/}hvm"
binName="hvm"
download="${platform}.${extension}"
artifactUrl="https://github.com/josephschmitt/hvm/releases/download/${version}/hvm-${version}-${download}"

mkdir -p "${tmpdir}"

pushd "${tmpdir}" > /dev/null

printf "\nDownloading: %s to %s\n" "$(green "${artifactUrl}")" "$(cyan "${tmpdir}/")"
curl -s -f -L "${artifactUrl}" > "${download}"

case "${extension}" in
  "zip") unzip -j -d "${download}" >/dev/null 2>&1 ;;
  "gz") gzip -d -f "${download}" >/dev/null 2>&1 ;;
  "tar.gz") tar -xzf "${download}" >/dev/null 2>&1 ;;
esac

if [ -f "${download}" ]; then
  # Cleanup downloads
  rm "${download}"
fi

installDir=${HVM_INSTALL_LOCATION:-"/usr/local/bin"}

popd > /dev/null

# Check to make sure we have write perms to the install directory
if ! test -w "${installDir}"; then
  printf "\n%s %s\n%s\n%s %s %s\n" "$(red "Unable to install to")" "$(cyan "${installDir}")" \
    "You don't have write permission to this directory." \
    "Try setting the" "$(yellow "HVM_INSTALL_LOCATION")" "env to a different directory and try again." >&2
  exit 1
fi

# Move into /usr/local/bin
mkdir -p "${installDir}"
printf "\nInstalling to %s\n" "$(cyan "${installDir}")"
printf "  "
chmod +x "${tmpdir}/${binName}"
mv -f -v "${tmpdir}/${binName}" "${installDir}/hvm"

# Clean up tmp folder
rm -r "${tmpdir}"

{
  printf "\nRunning post-install...\n"
  "${installDir}/hvm" "update-repos" 2> /dev/null || printf ""
  "${installDir}/hvm" "install-completions" 2> /dev/null || printf ""
}
printf "HVM %s has been downloaded and installed to $(cyan "${installDir}").\n" "$(yellow "${version}")"
if [ -n "${HVM_INSTALL_LOCATION:-}" ]; then
  printf "Please make sure to add %s to your %s in order for the CLI to be available globally.\n" "$(yellow "${HVM_INSTALL_LOCATION}")" "$(green "PATH")"
fi
